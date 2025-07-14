package botinput

// EntryInputs provides information on parsed inputs from bot API request
type EntryInputs struct {
	Entry  Entry
	Inputs []InputMessage
}

// EntryInput provides information on parsed input from bot API request
type EntryInput struct {
	Entry Entry
	Input InputMessage
}
