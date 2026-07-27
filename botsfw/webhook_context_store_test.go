package botsfw

import (
	"context"
	"net/http"
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw-store/botsfwstore"
	"github.com/bots-go-framework/bots-fw-store/botsfwstore/botsfwstoretest"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/strongo/i18n"
)

type storeTestFieldsSetter struct{}

func (storeTestFieldsSetter) Platform() string { return "test" }

func (storeTestFieldsSetter) SetBotUserFields(data botsfwmodels.PlatformUserData, _ botinput.Sender, _, _, appUserID string) error {
	data.SetAppUserID(appUserID)
	return nil
}

func (storeTestFieldsSetter) SetBotChatFields(data botsfwmodels.BotChatData, _ botinput.Chat, _, _, appUserID string, accessGranted bool) error {
	data.SetAppUserID(appUserID)
	data.SetAccessGranted(accessGranted)
	return nil
}

func TestWebhookContextBase_UsesStateStoreForIdentityAndChat(t *testing.T) {
	var ensureCalls, saveCalls, accessCalls int
	var linkedIdentity botsfwstore.Identity
	store := &botsfwstoretest.FakeStateStore{}
	store.EnsureLinkedFunc = func(_ context.Context, request botsfwstore.LinkRequest) (botsfwstore.LinkedIdentity, error) {
		ensureCalls++
		linkedIdentity = request.Identity
		platformData, err := request.NewPlatformUserData("app-1")
		if err != nil {
			return botsfwstore.LinkedIdentity{}, err
		}
		chatData, err := request.NewChatData("app-1", true)
		if err != nil {
			return botsfwstore.LinkedIdentity{}, err
		}
		return botsfwstore.LinkedIdentity{
			AppUser:      botsfwstore.AppUser{ID: "app-1"},
			PlatformUser: botsfwstore.PlatformUser{ID: request.Identity.BotUserID, Data: platformData},
			ChatData:     chatData,
		}, nil
	}
	store.SaveChatFunc = func(_ context.Context, identity botsfwstore.Identity, data botsfwmodels.BotChatData) error {
		saveCalls++
		if identity != linkedIdentity || data.GetAppUserID() != "app-1" {
			t.Fatalf("SaveChat() received identity %#v and data %#v", identity, data)
		}
		return nil
	}
	store.SetPlatformUserAccessGrantedFunc = func(_ context.Context, identity botsfwstore.Identity, newData func() botsfwmodels.PlatformUserData, value bool) (botsfwstore.PlatformUser, error) {
		accessCalls++
		data := newData()
		data.SetAppUserID("app-1")
		data.SetAccessGranted(value)
		return botsfwstore.PlatformUser{ID: identity.BotUserID, Data: data}, nil
	}

	request, err := http.NewRequest(http.MethodPost, "/webhook", nil)
	if err != nil {
		t.Fatal(err)
	}
	profile := newTestProfile("store-test")
	settings := &BotSettings{Code: "bot-1", Token: "token", Locale: i18n.LocaleEnUS, Profile: profile, Store: store}
	input := &moreTestInputMessage{inputType: botinput.TypeText, chatID: "chat-1", senderID: "user-1", firstName: "Ada", lastName: "Lovelace", language: "en"}
	whc, err := NewWebhookContextBase(
		CreateWebhookContextArgs{
			HttpRequest:  request,
			AppContext:   testAppContext{},
			BotContext:   BotContext{BotHost: testBotHost{}, BotSettings: settings},
			WebhookInput: input,
			Store:        store,
		},
		testBotPlatformMore{},
		storeTestFieldsSetter{},
		func() (bool, error) { return false, nil },
		func(context.Context) (string, string, error) { return "", "chat-1", nil },
	)
	if err != nil {
		t.Fatalf("NewWebhookContextBase() error: %v", err)
	}

	chat := whc.ChatData()
	if chat == nil || chat.GetAppUserID() != "app-1" || whc.AppUserID() != "app-1" {
		t.Fatalf("linked context = chat:%#v appUser:%q", chat, whc.AppUserID())
	}
	if ensureCalls != 1 || linkedIdentity.PlatformID != "test" || linkedIdentity.BotID != "bot-1" || linkedIdentity.BotUserID != "user-1" || linkedIdentity.ChatID != "chat-1" {
		t.Fatalf("EnsureLinked() calls = %d, identity = %#v", ensureCalls, linkedIdentity)
	}
	if second := whc.ChatData(); second != chat || ensureCalls != 1 {
		t.Fatalf("second ChatData() = %#v, EnsureLinked() calls = %d", second, ensureCalls)
	}
	if err := whc.SetBotUserAccessGranted(false); err != nil {
		t.Fatalf("SetBotUserAccessGranted() error: %v", err)
	}
	if err := whc.SaveBotChat(); err != nil {
		t.Fatalf("SaveBotChat() error: %v", err)
	}
	if accessCalls != 1 || saveCalls != 1 {
		t.Fatalf("store calls = access:%d save:%d", accessCalls, saveCalls)
	}
}

func TestWebhookContextBase_AppUserIDLinksChatlessPlatformIdentity(t *testing.T) {
	var ensureCalls int
	store := &botsfwstoretest.FakeStateStore{}
	store.EnsureLinkedFunc = func(
		_ context.Context,
		request botsfwstore.LinkRequest,
	) (botsfwstore.LinkedIdentity, error) {
		ensureCalls++
		if request.Identity.ChatID != "" {
			t.Fatalf("chatless identity unexpectedly has chat ID %q", request.Identity.ChatID)
		}
		if request.NewChatData != nil {
			t.Fatal("chatless identity must not create chat data")
		}
		platformData, err := request.NewPlatformUserData("app-inline")
		if err != nil {
			return botsfwstore.LinkedIdentity{}, err
		}
		return botsfwstore.LinkedIdentity{
			AppUser:      botsfwstore.AppUser{ID: "app-inline"},
			PlatformUser: botsfwstore.PlatformUser{ID: request.Identity.BotUserID, Data: platformData},
		}, nil
	}

	request, err := http.NewRequest(http.MethodPost, "/webhook", nil)
	if err != nil {
		t.Fatal(err)
	}
	profile := newTestProfile("store-test-inline")
	settings := &BotSettings{
		Code: "bot-1", Token: "token", Locale: i18n.LocaleEnUS,
		Profile: profile, Store: store,
	}
	input := &moreTestInputMessage{
		inputType: botinput.TypeInlineQuery,
		senderID:  "user-inline",
		firstName: "Ada",
		language:  "ru",
	}
	whc, err := NewWebhookContextBase(
		CreateWebhookContextArgs{
			HttpRequest:  request,
			AppContext:   testAppContext{},
			BotContext:   BotContext{BotHost: testBotHost{}, BotSettings: settings},
			WebhookInput: input,
			Store:        store,
		},
		testBotPlatformMore{},
		storeTestFieldsSetter{},
		func() (bool, error) { return false, nil },
		func(context.Context) (string, string, error) { return "", "", nil },
	)
	if err != nil {
		t.Fatalf("NewWebhookContextBase() error: %v", err)
	}

	if got := whc.AppUserID(); got != "app-inline" {
		t.Fatalf("AppUserID() = %q, want app-inline", got)
	}
	if ensureCalls != 1 {
		t.Fatalf("EnsureLinked() calls = %d, want 1", ensureCalls)
	}
	if got := whc.AppUserID(); got != "app-inline" || ensureCalls != 1 {
		t.Fatalf("second AppUserID() = %q, EnsureLinked calls = %d", got, ensureCalls)
	}
	if whc.ChatData() != nil {
		t.Fatal("chatless inline query unexpectedly created chat data")
	}
}
