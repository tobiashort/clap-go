package ansi

import "fmt"

type CursorControl = string

const (
	CursorHide               CursorControl = "\033[?25l"
	CursorShow               CursorControl = "\033[?25h"
	CursorMoveToHomePosition CursorControl = "\033[H"
)

func CursorMoveTo(line, column int) CursorControl {
	return fmt.Sprintf("\033[%d;%dH", line, column)
}

func CursorMoveUp(lines int) CursorControl {
	return fmt.Sprintf("\033[%dA", lines)
}

func CursorMoveDown(lines int) CursorControl {
	return fmt.Sprintf("\033[%dB", lines)
}

func CursorMoveRight(columns int) CursorControl {
	return fmt.Sprintf("\033[%dC", columns)
}

func CursorMoveLeft(columns int) CursorControl {
	return fmt.Sprintf("\033[%dD", columns)
}

func CursorMoveToColumn(column int) CursorControl {
	return fmt.Sprintf("\033[%dG", column)
}
