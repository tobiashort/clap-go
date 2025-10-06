package ansi

type Decor = string

const (
	DecorRed    Decor = "\033[31m"
	DecorGreen  Decor = "\033[32m"
	DecorYellow Decor = "\033[33m"
	DecorBlue   Decor = "\033[34m"
	DecorPurple Decor = "\033[35m"
	DecorCyan   Decor = "\033[36m"

	DecorBold      Decor = "\033[1m"
	DecorUnderline Decor = "\033[4m"
	DecorReversed  Decor = "\033[7m"

	DecorReset Decor = "\033[0m"
)
