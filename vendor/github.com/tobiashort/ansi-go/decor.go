package ansi

type Decor = string

const (
	Red    Decor = "\033[31m"
	Green  Decor = "\033[32m"
	Yellow Decor = "\033[33m"
	Blue   Decor = "\033[34m"
	Purple Decor = "\033[35m"
	Cyan   Decor = "\033[36m"

	Bold      Decor = "\033[1m"
	Underline Decor = "\033[4m"
	Reversed  Decor = "\033[7m"

	Reset Decor = "\033[0m"
)
