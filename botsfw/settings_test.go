package botsfw

import (
	"context"
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strongo/i18n"
)

func newTestProfile(id string) BotProfile {
	return NewBotProfile(
		id,
		nil,
		func() botsfwmodels.BotChatData { return nil },
		func() botsfwmodels.PlatformUserData { return nil },
		func() botsfwmodels.AppUserData { return nil },
		nil,
		i18n.LocaleEnUS,
		nil,
		BotTranslations{},
	)
}

func newTestBotSettings(code, id, token string) BotSettings {
	return NewBotSettings(
		botsfwconst.Platform("test"),
		"local",
		newTestProfile(code+"-profile"),
		code,
		id,
		token,
		"",
		i18n.LocaleEnUS,
		nil,
		nil,
	)
}

func TestNewBotSettings(t *testing.T) {
	profile := newTestProfile("prof1")

	t.Run("panics_on_empty_platform", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewBotSettings: missing required parameter: platform", func() {
			NewBotSettings("", "local", profile, "code", "", "tok", "", i18n.LocaleEnUS, nil, nil)
		})
	})

	t.Run("panics_on_nil_profile", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewBotSettings: missing required parameter: profile", func() {
			NewBotSettings("telegram", "local", nil, "code", "", "tok", "", i18n.LocaleEnUS, nil, nil)
		})
	})

	t.Run("panics_on_empty_code", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewBotSettings: missing required parameter: code", func() {
			NewBotSettings("telegram", "local", profile, "", "", "tok", "", i18n.LocaleEnUS, nil, nil)
		})
	})

	t.Run("panics_on_empty_token", func(t *testing.T) {
		assert.Panics(t, func() {
			NewBotSettings("telegram", "local", profile, "mybot", "", "", "", i18n.LocaleEnUS, nil, nil)
		})
	})

	t.Run("panics_on_empty_locale", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewBotSettings: missing required parameter: Locale.Code5", func() {
			NewBotSettings("telegram", "local", profile, "mybot", "", "tok123", "", i18n.Locale{}, nil, nil)
		})
	})

	t.Run("valid_construction", func(t *testing.T) {
		bs := NewBotSettings("telegram", "prod", profile, "mybot", "bot123", "tok123", "ga-token", i18n.LocaleEnUS, nil, nil)
		assert.Equal(t, botsfwconst.Platform("telegram"), bs.Platform)
		assert.Equal(t, "prod", bs.Env)
		assert.Equal(t, profile, bs.Profile)
		assert.Equal(t, "mybot", bs.Code)
		assert.Equal(t, "bot123", bs.ID)
		assert.Equal(t, "tok123", bs.Token)
		assert.Equal(t, "ga-token", bs.GAToken)
		assert.Equal(t, i18n.LocaleEnUS, bs.Locale)
	})
}

func TestBotSettings_GetAppUserByID(t *testing.T) {
	var calledWith struct {
		botCode   string
		appUserID string
	}
	getter := func(ctx context.Context, tx dal.ReadSession, botID string, appUserID string) (record.DataWithID[string, botsfwmodels.AppUserData], error) {
		calledWith.botCode = botID
		calledWith.appUserID = appUserID
		return record.DataWithID[string, botsfwmodels.AppUserData]{}, nil
	}

	bs := BotSettings{
		Code:       "testbot",
		getAppUser: getter,
	}

	_, err := bs.GetAppUserByID(context.Background(), nil, "user42")
	require.NoError(t, err)
	assert.Equal(t, "testbot", calledWith.botCode)
	assert.Equal(t, "user42", calledWith.appUserID)
}

func TestNewBotSettingsBy(t *testing.T) {
	t.Run("panics_on_empty_bots", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewBotSettingsBy: missing required parameter: bots", func() {
			NewBotSettingsBy()
		})
	})

	t.Run("panics_on_empty_bot_code", func(t *testing.T) {
		assert.Panics(t, func() {
			NewBotSettingsBy(BotSettings{})
		})
	})

	t.Run("panics_on_duplicate_code", func(t *testing.T) {
		bot1 := newTestBotSettings("bot1", "id1", "tok1")
		bot2 := newTestBotSettings("bot1", "id2", "tok2")
		assert.Panics(t, func() {
			NewBotSettingsBy(bot1, bot2)
		})
	})

	t.Run("panics_on_duplicate_id", func(t *testing.T) {
		bot1 := newTestBotSettings("bot1", "sameid", "tok1")
		bot2 := newTestBotSettings("bot2", "sameid", "tok2")
		assert.Panics(t, func() {
			NewBotSettingsBy(bot1, bot2)
		})
	})

	t.Run("valid_single_bot", func(t *testing.T) {
		bot := newTestBotSettings("mybot", "id1", "tok1")
		sb := NewBotSettingsBy(bot)
		require.NotNil(t, sb.ByCode)
		require.NotNil(t, sb.ByID)
		assert.Contains(t, sb.ByCode, "mybot")
		assert.Contains(t, sb.ByID, "id1")
		assert.Equal(t, "mybot", sb.ByCode["mybot"].Code)
	})

	t.Run("valid_multiple_bots", func(t *testing.T) {
		bot1 := newTestBotSettings("bot_a", "idA", "tokA")
		bot2 := newTestBotSettings("bot_b", "idB", "tokB")
		sb := NewBotSettingsBy(bot1, bot2)
		assert.Len(t, sb.ByCode, 2)
		assert.Len(t, sb.ByID, 2)
		assert.Contains(t, sb.ByCode, "bot_a")
		assert.Contains(t, sb.ByCode, "bot_b")
	})

	t.Run("bot_with_empty_ID_no_ByID_entry", func(t *testing.T) {
		bot := newTestBotSettings("noIdBot", "", "tok1")
		sb := NewBotSettingsBy(bot)
		assert.Contains(t, sb.ByCode, "noIdBot")
		assert.NotContains(t, sb.ByID, "")
		assert.Empty(t, sb.ByID)
	})

	t.Run("ByProfile_mapping", func(t *testing.T) {
		bot1 := newTestBotSettings("botX", "idX", "tokX")
		bot2 := newTestBotSettings("botY", "idY", "tokY")
		sb := NewBotSettingsBy(bot1, bot2)
		require.NotNil(t, sb.ByProfile)
		// Each bot has a different profile (code+"-profile"), so 2 entries
		assert.Len(t, sb.ByProfile, 2)

		// Verify profile IDs map correctly
		profileX := bot1.Profile.ID()
		profileY := bot2.Profile.ID()
		assert.Contains(t, sb.ByProfile, profileX)
		assert.Contains(t, sb.ByProfile, profileY)
		assert.Len(t, sb.ByProfile[profileX], 1)
		assert.Len(t, sb.ByProfile[profileY], 1)
	})

	t.Run("same_profile_multiple_bots", func(t *testing.T) {
		sharedProfile := newTestProfile("shared")
		bot1 := NewBotSettings("test", "local", sharedProfile, "c1", "id1", "tok1", "", i18n.LocaleEnUS, nil, nil)
		bot2 := NewBotSettings("test", "local", sharedProfile, "c2", "id2", "tok2", "", i18n.LocaleEnUS, nil, nil)
		sb := NewBotSettingsBy(bot1, bot2)
		assert.Len(t, sb.ByProfile["shared"], 2)
	})
}
