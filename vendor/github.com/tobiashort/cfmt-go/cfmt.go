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
	makeRegexp("B"):  ansi.DecorBold,
	makeRegexp("U"):  ansi.DecorUnderline,
	makeRegexp("R"):  ansi.DecorReversed,
	makeRegexp("r"):  ansi.DecorRed,
	makeRegexp("rB"): ansi.DecorRed + ansi.DecorBold,
	makeRegexp("rU"): ansi.DecorRed + ansi.DecorUnderline,
	makeRegexp("rR"): ansi.DecorRed + ansi.DecorReversed,
	makeRegexp("g"):  ansi.DecorGreen,
	makeRegexp("gB"): ansi.DecorGreen + ansi.DecorBold,
	makeRegexp("gU"): ansi.DecorGreen + ansi.DecorUnderline,
	makeRegexp("gR"): ansi.DecorGreen + ansi.DecorReversed,
	makeRegexp("y"):  ansi.DecorYellow,
	makeRegexp("yB"): ansi.DecorYellow + ansi.DecorBold,
	makeRegexp("yU"): ansi.DecorYellow + ansi.DecorUnderline,
	makeRegexp("yR"): ansi.DecorYellow + ansi.DecorReversed,
	makeRegexp("b"):  ansi.DecorBlue,
	makeRegexp("bB"): ansi.DecorBlue + ansi.DecorBold,
	makeRegexp("bU"): ansi.DecorBlue + ansi.DecorUnderline,
	makeRegexp("bR"): ansi.DecorBlue + ansi.DecorReversed,
	makeRegexp("p"):  ansi.DecorPurple,
	makeRegexp("pB"): ansi.DecorPurple + ansi.DecorBold,
	makeRegexp("pU"): ansi.DecorPurple + ansi.DecorUnderline,
	makeRegexp("pR"): ansi.DecorPurple + ansi.DecorReversed,
	makeRegexp("c"):  ansi.DecorCyan,
	makeRegexp("cB"): ansi.DecorCyan + ansi.DecorBold,
	makeRegexp("cU"): ansi.DecorCyan + ansi.DecorUnderline,
	makeRegexp("cR"): ansi.DecorCyan + ansi.DecorReversed,
}

func makeRegexp(name string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("#%s\\{([^}]*)\\}", name))
}

func Print(a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.DecorReset)
	}
	fmt.Print(a...)
}

func Printf(format string, a ...any) {
	fmt.Printf(clr(format, ansi.DecorReset), a...)
}

func Println(a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.DecorReset)
	}
	fmt.Println(a...)
}

func Fprint(w io.Writer, a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.DecorReset)
	}
	fmt.Fprint(w, a...)
}

func Fprintf(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, clr(format, ansi.DecorReset), a...)
}

func Fprintln(w io.Writer, a ...any) {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.DecorReset)
	}
	fmt.Fprintln(w, a...)
}

func Sprint(a ...any) string {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.DecorReset)
	}
	return fmt.Sprint(a...)
}

func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(clr(format, ansi.DecorReset), a...)
}

func Sprintln(a ...any) string {
	for i := range a {
		a[i] = clr(fmt.Sprint(a[i]), ansi.DecorReset)
	}
	return fmt.Sprintln(a...)
}

func stoc(s string) ansi.Decor {
	switch s {
	case "r":
		return ansi.DecorRed
	case "g":
		return ansi.DecorGreen
	case "y":
		return ansi.DecorYellow
	case "b":
		return ansi.DecorBlue
	case "p":
		return ansi.DecorPurple
	case "c":
		return ansi.DecorCyan
	case "B":
		return ansi.DecorBold
	case "rB":
		return ansi.DecorRed + ansi.DecorBold
	case "gB":
		return ansi.DecorGreen + ansi.DecorBold
	case "yB":
		return ansi.DecorYellow + ansi.DecorBold
	case "bB":
		return ansi.DecorBlue + ansi.DecorBold
	case "pB":
		return ansi.DecorPurple + ansi.DecorBold
	case "cB":
		return ansi.DecorCyan + ansi.DecorBold
	case "U":
		return ansi.DecorUnderline
	case "rU":
		return ansi.DecorRed + ansi.DecorUnderline
	case "gU":
		return ansi.DecorGreen + ansi.DecorUnderline
	case "yU":
		return ansi.DecorYellow + ansi.DecorUnderline
	case "bU":
		return ansi.DecorBlue + ansi.DecorUnderline
	case "pU":
		return ansi.DecorPurple + ansi.DecorUnderline
	case "cU":
		return ansi.DecorCyan + ansi.DecorUnderline
	case "R":
		return ansi.DecorReversed
	case "rR":
		return ansi.DecorRed + ansi.DecorReversed
	case "gR":
		return ansi.DecorGreen + ansi.DecorReversed
	case "yR":
		return ansi.DecorYellow + ansi.DecorReversed
	case "bR":
		return ansi.DecorBlue + ansi.DecorReversed
	case "pR":
		return ansi.DecorPurple + ansi.DecorReversed
	case "cR":
		return ansi.DecorCyan + ansi.DecorReversed
	default:
		panic(fmt.Errorf("cannot map string '%s' to ansi Decorcolor", s))
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
		fmt.Print(ansi.DecorReset)
	}
}

func CPrintf(s string, format string, a ...any) {
	c := stoc(s)
	if isatty.IsTerminal() {
		fmt.Print(c)
	}
	fmt.Printf(clr(format, c), a...)
	if isatty.IsTerminal() {
		fmt.Print(ansi.DecorReset)
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
		fmt.Print(ansi.DecorReset)
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

func Begin(decor ansi.Decor) {
	fmt.Print(decor)
}

func End() {
	fmt.Print(ansi.DecorReset)
}
