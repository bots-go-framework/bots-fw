package botkb

// KeyboardType defines MessageKeyboard type
type KeyboardType int

//goland:noinspection GoUnusedConst
const (
	// KeyboardTypeNone for no MessageKeyboard
	// Used by: Telegram
	KeyboardTypeNone KeyboardType = iota

	// KeyboardTypeHide commands to hide MessageKeyboard
	// Used by: Telegram
	KeyboardTypeHide

	// KeyboardTypeInline for inline MessageKeyboard
	// Used by: Telegram
	KeyboardTypeInline

	// KeyboardTypeBottom for bottom MessageKeyboard
	// Used by: Telegram
	KeyboardTypeBottom

	// KeyboardTypeForceReply to force reply from a user
	// Used by: Telegram
	KeyboardTypeForceReply
)

// Keyboard defines MessageKeyboard
type Keyboard interface {
	// KeyboardType defines MessageKeyboard type
	KeyboardType() KeyboardType
}

var _ Keyboard = (*MessageKeyboard)(nil)

type MessageKeyboard struct {
	kbType  KeyboardType
	Buttons [][]Button
}

func (k MessageKeyboard) KeyboardType() KeyboardType {
	return k.kbType
}

func NewMessageKeyboard(kbType KeyboardType, buttons ...[]Button) *MessageKeyboard {
	return &MessageKeyboard{
		kbType:  kbType,
		Buttons: buttons,
	}
}
