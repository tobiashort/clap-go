package ansi

type EraseFunction = string

const (
	EraseFromCursorToEndOfScreen   EraseFunction = "\033[0J"
	EraseFromCursorToStartOfScreen EraseFunction = "\033[1J"
	EraseEntireScreen              EraseFunction = "\033[2J"
	EraseFromCursorToEndOfLine     EraseFunction = "\033[0K"
	EraseFromCursorToStartOfLine   EraseFunction = "\033[1K"
	EraseEntireLine                EraseFunction = "\033[2K"
)
