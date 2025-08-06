//go:build windows

package isatty

//#include <stdio.h>
//#include <io.h>
import "C"

// Checks whether stdout is a terminal or not
func IsTerminal() bool {
	return C._isatty(C.int(1)) != 0
}
