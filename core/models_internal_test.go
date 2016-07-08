package bots

import (
	"testing"
)

func TestBotChatEntity_PopStepsFromAwaitingReplyToUpTo(t *testing.T) {
	logger := &MockLogger{T: t}
	chatEntity := BotChatEntity{}

	chatEntity.AwaitingReplyTo = "step1/step2/step3"
	chatEntity.PopStepsFromAwaitingReplyToUpTo("step2", logger)
	if chatEntity.AwaitingReplyTo != "step1/step2" {
		t.Errorf("Failed to remove last step3. AwaitingReplyTo: " + chatEntity.AwaitingReplyTo)
	}

	chatEntity.AwaitingReplyTo = "step1/step2"
	chatEntity.PopStepsFromAwaitingReplyToUpTo("step2", logger)
	if chatEntity.AwaitingReplyTo != "step1/step2" {
		t.Errorf("Failed to remove last step3. AwaitingReplyTo: " + chatEntity.AwaitingReplyTo)
	}
}
