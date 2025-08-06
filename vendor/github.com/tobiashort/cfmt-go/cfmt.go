package cfmt

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/tobiashort/ansi-go"
	"github.com/tobiashort/isatty-go"
)

var regexps = map[*regexp.Regexp]ansi.Decor{
	makeRegexp("B"):  ansi.Bold,
	makeRegexp("U"):  ansi.Underline,
	makeRegexp("R"):  ansi.Reversed,
	makeRegexp("r"):  ansi.Red,
	makeRegexp("rB"): ansi.Red + ansi.Bold,
	makeRegexp("rU"): ansi.Red + ansi.Underline,
	makeRegexp("rR"): ansi.Red + ansi.Reversed,
	makeRegexp("g"):  ansi.Green,
	makeRegexp("gB"): ansi.Green + ansi.Bold,
	makeRegexp("gU"): ansi.Green + ansi.Underline,
	makeRegexp("gR"): ansi.Green + ansi.Reversed,
	makeRegexp("y"):  ansi.Yellow,
	makeRegexp("yB"): ansi.Yellow + ansi.Bold,
	makeRegexp("yU"): ansi.Yellow + ansi.Underline,
	makeRegexp("yR"): ansi.Yellow + ansi.Reversed,
	makeRegexp("b"):  ansi.Blue,
	makeRegexp("bB"): ansi.Blue + ansi.Bold,
	makeRegexp("bU"): ansi.Blue + ansi.Underline,
	makeRegexp("bR"): ansi.Blue + ansi.Reversed,
	makeRegexp("p"):  ansi.Purple,
	makeRegexp("pB"): ansi.Purple + ansi.Bold,
	makeRegexp("pU"): ansi.Purple + ansi.Underline,
	makeRegexp("pR"): ansi.Purple + ansi.Reversed,
	makeRegexp("c"):  ansi.Cyan,
	makeRegexp("cB"): ansi.Cyan + ansi.Bold,
	makeRegexp("cU"): ansi.Cyan + ansi.Underline,
	makeRegexp("cR"): ansi.Cyan + ansi.Reversed,
}

func makeRegexp(name string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("#%s\\{([^}]*)\\}", name))
}

func Print(a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.Reset)
	}
	fmt.Print(a...)
}

func Printf(format string, a ...any) {
	fmt.Printf(clr(format, ansi.Reset), a...)
}

func Println(a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.Reset)
	}
	fmt.Println(a...)
}

func Fprint(w io.Writer, a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.Reset)
	}
	fmt.Fprint(w, a...)
}

func Fprintf(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, clr(format, ansi.Reset), a...)
}

func Fprintln(w io.Writer, a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.Reset)
	}
	fmt.Fprintln(w, a...)
}

func Sprint(a ...any) string {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.Reset)
	}
	return fmt.Sprint(a...)
}

func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(clr(format, ansi.Reset), a...)
}

func Sprintln(a ...any) string {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.Reset)
	}
	return fmt.Sprintln(a...)
}

func stoc(s string) ansi.Decor {
	switch s {
	case "r":
		return ansi.Red
	case "g":
		return ansi.Green
	case "y":
		return ansi.Yellow
	case "b":
		return ansi.Blue
	case "p":
		return ansi.Purple
	case "c":
		return ansi.Cyan
	case "B":
		return ansi.Bold
	case "rB":
		return ansi.Red + ansi.Bold
	case "gB":
		return ansi.Green + ansi.Bold
	case "yB":
		return ansi.Yellow + ansi.Bold
	case "bB":
		return ansi.Blue + ansi.Bold
	case "pB":
		return ansi.Purple + ansi.Bold
	case "cB":
		return ansi.Cyan + ansi.Bold
	case "U":
		return ansi.Underline
	case "rU":
		return ansi.Red + ansi.Underline
	case "gU":
		return ansi.Green + ansi.Underline
	case "yU":
		return ansi.Yellow + ansi.Underline
	case "bU":
		return ansi.Blue + ansi.Underline
	case "pU":
		return ansi.Purple + ansi.Underline
	case "cU":
		return ansi.Cyan + ansi.Underline
	case "R":
		return ansi.Reversed
	case "rR":
		return ansi.Red + ansi.Reversed
	case "gR":
		return ansi.Green + ansi.Reversed
	case "yR":
		return ansi.Yellow + ansi.Reversed
	case "bR":
		return ansi.Blue + ansi.Reversed
	case "pR":
		return ansi.Purple + ansi.Reversed
	case "cR":
		return ansi.Cyan + ansi.Reversed
	default:
		panic(fmt.Errorf("cannot map string '%s' to ansi color", s))
	}
}

func CPrint(s string, a ...any) {
	c := stoc(s)
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), c)
	}
	if isatty.IsTerminal() {
		fmt.Print(c)
	}
	fmt.Print(a...)
	if isatty.IsTerminal() {
		fmt.Print(ansi.Reset)
	}
}

func CPrintf(s string, format string, a ...any) {
	c := stoc(s)
	if isatty.IsTerminal() {
		fmt.Print(c)
	}
	fmt.Printf(clr(format, c), a...)
	if isatty.IsTerminal() {
		fmt.Print(ansi.Reset)
	}
}

func CPrintln(s string, a ...any) {
	c := stoc(s)
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), c)
	}
	if isatty.IsTerminal() {
		fmt.Print(c)
	}
	fmt.Println(a...)
	if isatty.IsTerminal() {
		fmt.Print(ansi.Reset)
	}
}

func clr(str string, reset ansi.Decor) string {
	for regex, color := range regexps {
		matches := regex.FindAllStringSubmatch(str, -1)
		for _, match := range matches {
			if isatty.IsTerminal() {
				str = strings.Replace(str, match[0], color+match[1]+reset, 1)
			} else {
				str = strings.Replace(str, match[0], match[1], 1)
			}
		}
	}
	return str
}
