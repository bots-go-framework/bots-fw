package bots

import (
	"testing"
)

func TestBotChatEntity_PopStepsFromAwaitingReplyUpToSpecificParent(t *testing.T) {
	logger := &MockLogger{T: t}
	chatEntity := BotChatEntity{}

	chatEntity.AwaitingReplyTo = "step1/step2/step3"
	chatEntity.PopStepsFromAwaitingReplyUpToSpecificParent("step2", logger)
	if chatEntity.AwaitingReplyTo != "step1/step2" {
		t.Errorf("Failed to remove last step3. AwaitingReplyTo: " + chatEntity.AwaitingReplyTo)
	}

	chatEntity.AwaitingReplyTo = "step1/step2"
	chatEntity.PopStepsFromAwaitingReplyUpToSpecificParent("step2", logger)
	if chatEntity.AwaitingReplyTo != "step1/step2" {
		t.Errorf("Failed to remove last step3. AwaitingReplyTo: " + chatEntity.AwaitingReplyTo)
	}
}
