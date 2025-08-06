package ansi

import "fmt"

type CursorControl = string

const (
	MoveCursorToHomePosition CursorControl = "\033[H"
)

func MoveCursorTo(line, column int) CursorControl {
	return fmt.Sprintf("\033[%d;%dH", line, column)
}

func MoveCursorUp(lines int) CursorControl {
	return fmt.Sprintf("\033[%dA", lines)
}

func MoveCursorDown(lines int) CursorControl {
	return fmt.Sprintf("\033[%dB", lines)
}

func MoveCursorRight(columns int) CursorControl {
	return fmt.Sprintf("\033[%dC", columns)
}

func MoveCursorLeft(columns int) CursorControl {
	return fmt.Sprintf("\033[%dD", columns)
}

func MoveCursorToColumn(column int) CursorControl {
	return fmt.Sprintf("\033[%dG", column)
}
