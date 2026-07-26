package botmsg

// Type defines a type of output message from bot to user
type Type int

const (
	// TypeUndefined unknown type
	TypeUndefined Type = iota
	// TypeCallbackAnswer sends a callback answer
	TypeCallbackAnswer
	// BotMessageTypeInlineResults sends inline results
	BotMessageTypeInlineResults
	// TypeText sends text reply
	TypeText
	// TypeEditMessage edit previously sent message
	TypeEditMessage
	// TypeLeaveChat commands messenger to kick off bot from a botChat
	TypeLeaveChat
	// TypeExportChatInviteLink sends invite link
	TypeExportChatInviteLink

	TypeSendPhoto

	TypeSendInvoice
	TypeCreateInvoiceLink
	TypeAnswerPreCheckoutQuery

	TypeSetDescription
	TypeSetShortDescription
	TypeSetCommands

	// TypeChatAction shows a chat action indicator (e.g. "typing") in the target chat
	TypeChatAction

	// TypeSendRichMessage sends a native Telegram rich message.
	TypeSendRichMessage

	// TypeSendRichMessageDraft streams a temporary native Telegram rich-message draft.
	TypeSendRichMessageDraft

	// TypeEditRichMessage edits an existing native Telegram rich message.
	TypeEditRichMessage
)
