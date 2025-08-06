//go:build linux || darwin

package isatty

//#include <unistd.h>
import "C"

import (
	"os"
)

// Checks whether stdout is a terminal or not
func IsTerminal() bool {
	return int(C.isatty(C.int(os.Stdout.Fd()))) == 1
}
