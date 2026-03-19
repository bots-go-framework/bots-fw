package botsfw

import (
	"testing"

	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWebhookNewContext(t *testing.T) {
	t.Run("zero_value", func(t *testing.T) {
		wnc := WebhookNewContext{}
		assert.Nil(t, wnc.BotSettings)
		assert.Nil(t, wnc.InputMessage)
	})

	t.Run("with_values", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockInput := mock_botinput.NewMockInputMessage(ctrl)

		wnc := WebhookNewContext{
			BotContext: BotContext{
				BotSettings: &BotSettings{Code: "testbot"},
			},
			InputMessage: mockInput,
		}
		assert.NotNil(t, wnc.BotSettings)
		assert.Equal(t, "testbot", wnc.BotSettings.Code)
		assert.NotNil(t, wnc.InputMessage)
	})
}
