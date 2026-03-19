package botsfw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCreateWebhookContextArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockInput := mock_botinput.NewMockInputMessage(ctrl)

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	appCtx := testAppContext{}
	botCtx := BotContext{
		BotSettings: &BotSettings{Code: "testbot"},
	}

	args := NewCreateWebhookContextArgs(req, appCtx, botCtx, mockInput, nil)

	assert.Equal(t, req, args.HttpRequest)
	assert.Equal(t, appCtx, args.AppContext)
	assert.Equal(t, botCtx, args.BotContext)
	assert.Equal(t, mockInput, args.WebhookInput)
	assert.Nil(t, args.Db)
}

func TestCreateWebhookContextArgs_zero_value(t *testing.T) {
	args := CreateWebhookContextArgs{}
	assert.Nil(t, args.HttpRequest)
	assert.Nil(t, args.AppContext)
	assert.Nil(t, args.WebhookInput)
	assert.Nil(t, args.Db)
}
