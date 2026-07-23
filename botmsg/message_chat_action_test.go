package botmsg

import "testing"

func TestChatAction_BotMessageType(t *testing.T) {
	if got := (ChatAction{}).BotMessageType(); got != TypeChatAction {
		t.Errorf("BotMessageType() = %v, want %v", got, TypeChatAction)
	}
}

func TestChatAction_Validate(t *testing.T) {
	if err := (ChatAction{Action: ChatActionTyping}).Validate(); err != nil {
		t.Errorf("Validate() with a typing action returned %v, want nil", err)
	}
	if err := (ChatAction{Action: ChatActionTyping, ChatID: "123"}).Validate(); err != nil {
		t.Errorf("Validate() with a chat id returned %v, want nil", err)
	}
	if err := (ChatAction{}).Validate(); err == nil {
		t.Error("Validate() with no action returned nil, want an error")
	}
}

// TypeChatAction must be distinct from the other message types (guards against an
// accidental duplicate iota value if the enum is reordered).
func TestTypeChatAction_IsDistinct(t *testing.T) {
	others := []Type{
		TypeUndefined, TypeCallbackAnswer, BotMessageTypeInlineResults, TypeText,
		TypeEditMessage, TypeLeaveChat, TypeExportChatInviteLink, TypeSendPhoto,
		TypeSendInvoice, TypeCreateInvoiceLink, TypeAnswerPreCheckoutQuery,
		TypeSetDescription, TypeSetShortDescription, TypeSetCommands,
	}
	for _, o := range others {
		if TypeChatAction == o {
			t.Fatalf("TypeChatAction (%d) collides with another message type", TypeChatAction)
		}
	}
}
