package clap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/tobiashort/cfmt-go"
)

var (
	prog        string = filepath.Base(os.Args[0])
	description string = ""
	example     string = ""
)

type arg struct {
	name          string
	type_         reflect.Type
	kind          reflect.Kind
	short         string
	long          string
	conflictsWith []string
	mandatory     bool
	positional    bool
	description   string
	defaultValue  string
}

type userError struct {
	msg string
}

func (err userError) Error() string {
	return err.msg
}

type developerError struct {
	msg string
}

func (err developerError) Error() string {
	return err.msg
}

func Prog(s string) {
	prog = s
}

func Description(s string) {
	description = s
}

func Example(s string) {
	example = s
}

func Parse(strct any) {
	defer func() {
		r := recover()
		if r != nil {
			switch err := r.(type) {
			case userError:
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			default:
				panic(r)
			}
		}
	}()
	parse(strct)
}

func parse(strct any) {
	if !isStructPointer(strct) {
		developerErr("expected struct pointer")
	}

	strctType := reflect.TypeOf(strct).Elem()

	programArgs := make([]arg, 0)

	for i := range strctType.NumField() {
		field := strctType.Field(i)

		var (
			long          = toKebabCase(field.Name)
			short         = string(strings.ToLower(field.Name)[0])
			conflictsWith = make([]string, 0)
			mandatory     = false
			positional    = false
			description   = ""
			defaultValue  = ""
		)

		tag := field.Tag.Get("clap")
		if tag != "" {
			tagValues := parseTagValues(tag)

			for _, tagValue := range tagValues {
				if strings.HasPrefix(tagValue, "short=") {
					short = strings.Split(tagValue, "=")[1]
				} else if strings.HasPrefix(tagValue, "long=") {
					long = strings.Split(tagValue, "=")[1]
				} else if strings.HasPrefix(tagValue, "conflicts-with=") {
					conflictsWith = strings.Split(strings.Split(tagValue, "=")[1], ",")
				} else if strings.HasPrefix(tagValue, "default-value=") {
					defaultValue = strings.Split(tagValue, "=")[1]
				} else if strings.HasPrefix(tagValue, "description=") {
					description = strings.Split(tagValue, "=")[1]
				} else if tagValue == "mandatory" {
					mandatory = true
				} else if tagValue == "positional" {
					positional = true
				} else {
					developerErr("unknown tag value: " + tagValue)
				}
			}
		}

		programArgs = append(programArgs, arg{
			name:          field.Name,
			type_:         field.Type,
			kind:          field.Type.Kind(),
			long:          long,
			short:         short,
			conflictsWith: conflictsWith,
			mandatory:     mandatory,
			positional:    positional,
			description:   description,
			defaultValue:  defaultValue,
		})
	}

	implicitHelpArg := arg{
		name:        "Help",
		type_:       reflect.TypeOf(true),
		kind:        reflect.Bool,
		long:        "help",
		short:       "h",
		description: "Show this help message and exit",
	}

	programArgs = append(programArgs, implicitHelpArg)
	programNonPositionalArgs := filterArgs(programArgs, func(arg arg) bool { return !arg.positional })
	programPositionalArgs := filterArgs(programArgs, func(arg arg) bool { return arg.positional })

	checkForNameCollisions(programArgs)
	checkForMandatoryArgsWithDefaultValue(programArgs)
	checkForSlicesWithDefaultValue(programArgs)
	checkForEitherLongOrShortGiven(programNonPositionalArgs)
	checkForInvalidPositionalArguments(programPositionalArgs)

	givenNonPositionalArgs := make([]arg, 0)
	givenPositionalArgs := make([]arg, 0)

	positionalArgIndex := 0
	doubleDashSeen := false
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--" {
			doubleDashSeen = true
			continue
		}
		if !doubleDashSeen && strings.HasPrefix(arg, "--") {
			long := arg[2:]
			if long == "help" {
				printHelp(programArgs, os.Stdout)
				os.Exit(0)
			}
			arg, ok := getArgByLongName(programArgs, long)
			if !ok {
				userErr("unknown argument: --" + long)
			} else {
				givenNonPositionalArgs = append(givenNonPositionalArgs, arg)
			}
			i = parseNonPositionalAtIndex(arg, strct, i)
		} else if !doubleDashSeen && strings.HasPrefix(arg, "-") {
			shortGrouped := arg[1:]
			for _, rune := range shortGrouped {
				short := string(rune)
				if short == "h" {
					printHelp(programArgs, os.Stdout)
					os.Exit(0)
				}
				arg, ok := getArgByShortName(programArgs, short)
				if !ok {
					userErr("unknown argument: -" + short)
				} else {
					givenNonPositionalArgs = append(givenNonPositionalArgs, arg)
				}
				i = parseNonPositionalAtIndex(arg, strct, i)
			}
		} else {
			if len(programPositionalArgs) == 0 {
				userErr("too many arguments")
			} else if positionalArgIndex >= len(programPositionalArgs) && programPositionalArgs[len(programPositionalArgs)-1].kind != reflect.Slice {
				userErr("too many arguments")
			} else {
				positionalArg := programPositionalArgs[positionalArgIndex]
				givenPositionalArgs = append(givenPositionalArgs, positionalArg)
				parsePositionalAtIndex(positionalArg, strct, i)
				if positionalArgIndex+1 < len(programPositionalArgs) {
					positionalArgIndex++
				}
			}
		}
	}

	checkForConflicts(givenNonPositionalArgs)
	checkForMissingMandatoryArgs(programArgs, givenNonPositionalArgs, givenPositionalArgs)
	checkForMultipleUse(givenNonPositionalArgs)

outer:
	for _, arg := range programArgs {
		if arg.defaultValue == "" {
			continue
		}
		for _, givenArg := range givenNonPositionalArgs {
			if arg.name == givenArg.name {
				continue outer
			}
		}
		for _, givenArg := range givenPositionalArgs {
			if arg.name == givenArg.name {
				continue outer
			}
		}
		if arg.positional {
			parsePositional(arg, strct, arg.defaultValue)
		} else {
			parseNonPositional(arg, strct, arg.defaultValue)
		}
	}
}

func parseNonPositionalAtIndex(arg arg, strct any, index int) int {
	if arg.kind == reflect.Bool {
		parseNonPositional(arg, strct, "")
		return index
	} else {
		if index+1 >= len(os.Args) {
			userErr(fmt.Sprintf("missing value for: -%s|--%s", arg.short, arg.long))
		}
		value := os.Args[index+1]
		parseNonPositional(arg, strct, value)
		return index + 1
	}
}

func parseNonPositional(arg arg, strct any, value string) {
	if arg.kind == reflect.Bool {
		setBool(strct, arg.name, true)
	} else if arg.kind == reflect.String {
		setString(strct, arg.name, value)
	} else if arg.kind == reflect.Int {
		setInt(strct, arg.name, parseInt(value))
	} else if arg.kind == reflect.Float64 {
		setFloat(strct, arg.name, parseFloat(value))
	} else if arg.kind == reflect.Slice {
		innerKind := arg.type_.Elem().Kind()
		var parsed any
		if innerKind == reflect.String {
			parsed = value
		} else if innerKind == reflect.Int {
			parsed = parseInt(value)
		} else if innerKind == reflect.Float64 {
			parsed = parseFloat(value)
		} else {
			developerErr("not implemented argument kind []" + innerKind.String())
		}
		addToSlice(strct, arg.name, parsed)
	} else if arg.type_ == reflect.TypeOf(time.Duration(0)) {
		setDuration(strct, arg.name, parseDuration(value))
	} else {
		developerErr(fmt.Sprintf("not implemented argument kind: %v", arg.kind))
		panic("unreachable")
	}
}

func parsePositionalAtIndex(arg arg, strct any, index int) {
	value := os.Args[index]
	parsePositional(arg, strct, value)
}

func parsePositional(arg arg, strct any, value string) {
	if arg.kind == reflect.String {
		setString(strct, arg.name, value)
	} else if arg.kind == reflect.Int {
		setInt(strct, arg.name, parseInt(value))
	} else if arg.kind == reflect.Float64 {
		setFloat(strct, arg.name, parseFloat(value))
	} else if arg.kind == reflect.Slice {
		innerKind := arg.type_.Elem().Kind()
		var parsed any
		if innerKind == reflect.String {
			parsed = value
		} else if innerKind == reflect.Int {
			parsed = parseInt(value)
		} else if innerKind == reflect.Float64 {
			parsed = parseFloat(value)
		} else {
			developerErr("not implemented argument kind []" + innerKind.String())
		}
		addToSlice(strct, arg.name, parsed)
	} else if arg.type_ == reflect.TypeOf(time.Duration(0)) {
		setDuration(strct, arg.name, parseDuration(value))
	} else {
		developerErr(fmt.Sprintf("not implemented argument kind: %v", arg.kind))
	}
}

func parseInt(arg string) int {
	val, err := strconv.Atoi(arg)
	if err != nil {
		userErr("value is not an int: " + arg)
	}
	return val
}

func parseFloat(arg string) float64 {
	val, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		userErr("value is not a float: " + arg)
	}
	return val
}

func parseDuration(arg string) time.Duration {
	val, err := time.ParseDuration(arg)
	if err != nil {
		userErr("value is not a duration: " + arg)
	}
	return val
}

func isStructPointer(strct any) bool {
	t := reflect.TypeOf(strct)
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func getArgByLongName(args []arg, name string) (arg, bool) {
	for _, arg := range args {
		if arg.long == name {
			return arg, true
		}
	}
	return arg{}, false
}

func getArgByShortName(args []arg, name string) (arg, bool) {
	for _, arg := range args {
		if arg.short == name {
			return arg, true
		}
	}
	return arg{}, false
}

func setInt(strct any, name string, val int) {
	reflect.ValueOf(strct).Elem().FieldByName(name).SetInt(int64(val))
}

func setFloat(strct any, name string, val float64) {
	reflect.ValueOf(strct).Elem().FieldByName(name).SetFloat(val)
}

func setBool(strct any, name string, val bool) {
	reflect.ValueOf(strct).Elem().FieldByName(name).SetBool(val)
}

func setString(strct any, name string, val string) {
	reflect.ValueOf(strct).Elem().FieldByName(name).SetString(val)
}

func setDuration(strct any, name string, val time.Duration) {
	reflect.ValueOf(strct).Elem().FieldByName(name).Set(reflect.ValueOf(val))
}

func addToSlice(strct any, name string, val any) {
	field := reflect.ValueOf(strct).Elem().FieldByName(name)
	if field.IsNil() {
		field.Set(reflect.MakeSlice(field.Type(), 0, 1))
	}
	updatedSlice := reflect.Append(field, reflect.ValueOf(val))
	field.Set(updatedSlice)
}

func checkForNameCollisions(args []arg) {
	seenLong := make(map[string]arg)
	seenShort := make(map[string]arg)
	for _, arg := range args {
		if arg.positional {
			continue
		}
		if arg.long != "" {
			existing, exists := seenLong[arg.long]
			if !exists {
				seenLong[arg.long] = arg
			} else {
				developerErr(fmt.Sprintf("argument name collision: %s (--%s) with %s (--%s)", arg.name, arg.long, existing.name, existing.long))
			}
		}
		if arg.short != "" {
			existing, exists := seenShort[arg.short]
			if !exists {
				seenShort[arg.short] = arg
			} else {
				developerErr(fmt.Sprintf("argument name collision: %s (-%s) with %s (-%s)", arg.name, arg.short, existing.name, existing.short))
			}
		}
	}
}

func checkForSlicesWithDefaultValue(programArgs []arg) {
	for _, arg := range programArgs {
		if arg.kind == reflect.Slice && arg.defaultValue != "" {
			developerErr("slice arguments cannot have default values: " + arg.name)
		}
	}
}

func checkForMandatoryArgsWithDefaultValue(programArgs []arg) {
	for _, arg := range programArgs {
		if arg.mandatory && arg.defaultValue != "" {
			developerErr("mandatory arguments cannot have default values: " + arg.name)
		}
	}
}

func checkForEitherLongOrShortGiven(programNonPositionalArgs []arg) {
	for _, arg := range programNonPositionalArgs {
		if arg.long == "" && arg.short == "" {
			developerErr("Either long or short name must be specified: " + arg.name)
		}
	}
}

func checkForInvalidPositionalArguments(programPositionalArgs []arg) {
	sliceSeen := false
	optionalSeen := false

	for _, arg := range programPositionalArgs {
		if arg.kind != reflect.Slice && sliceSeen {
			developerErr("positional arguments of slices can only be located at the end: " + arg.name)
		}
		if arg.kind == reflect.Slice && optionalSeen {
			developerErr("when slice as a positional argument is used, all preceding positional arguments must be mandatory: " + arg.name)
		}
		if arg.mandatory && optionalSeen {
			developerErr("you cannot have mandatory positional arguments after optional ones: " + arg.name)
		}
		if arg.kind == reflect.Slice {
			sliceSeen = true
		}
		if !arg.mandatory {
			optionalSeen = true
		}
	}
}

func checkForConflicts(givenNonPositionalArgs []arg) {
	for _, outerArg := range givenNonPositionalArgs {
		for _, inConflict := range outerArg.conflictsWith {
			for _, innerArg := range givenNonPositionalArgs {
				if innerArg.name == inConflict {
					userErr(fmt.Sprintf("conflicting arguments: -%s|--%s, -%s|--%s", outerArg.short, outerArg.long, innerArg.short, innerArg.long))
				}
			}
		}
	}
}

func checkForMissingMandatoryArgs(programArgs []arg, givenNonPositionalArgs []arg, givenPositionalArgs []arg) {
	givenArgs := make([]arg, 0)
	givenArgs = append(givenArgs, givenNonPositionalArgs...)
	givenArgs = append(givenArgs, givenPositionalArgs...)

outer:
	for _, arg := range programArgs {
		if arg.mandatory {
			for _, givenArg := range givenArgs {
				if givenArg.name == arg.name {
					continue outer
				}
			}
			if arg.positional {
				userErr(fmt.Sprintf("missing mandatory positional argument: %s", arg.name))
			} else {
				userErr(fmt.Sprintf("missing mandatory argument: -%s|--%s", arg.short, arg.long))
			}
		}
	}
}

func checkForMultipleUse(givenNonPositionalArgs []arg) {
	seen := make(map[string]bool)
	for _, arg := range givenNonPositionalArgs {
		_, exists := seen[arg.name]
		if !exists {
			seen[arg.name] = true
		} else {
			if arg.kind != reflect.Slice {
				userErr(fmt.Sprintf("multiple use of argument -%s|--%s", arg.short, arg.long))
			}
		}
	}
}

func parseTagValues(tag string) []string {
	var tagValues []string

	var sb strings.Builder
	inQuotes := false
	escapeNext := false

	for i := range len(tag) {
		ch := tag[i]

		if escapeNext {
			sb.WriteByte(ch)
			escapeNext = false
			continue
		}

		switch ch {
		case '\\':
			escapeNext = true
		case '\'':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				sb.WriteByte(ch)
			} else {
				tagValues = append(tagValues, sb.String())
				sb.Reset()
			}
		default:
			sb.WriteByte(ch)
		}
	}

	if sb.Len() > 0 {
		tagValues = append(tagValues, sb.String())
	}

	return tagValues
}

func printHelp(args []arg, w io.Writer) {
	buf := bytes.Buffer{}

	if description != "" {
		fmt.Fprintf(&buf, "%s\n\n", description)
	}

	var usageParts []string
	usageParts = append(usageParts, prog)

	for _, arg := range args {
		if !arg.mandatory {
			usageParts = append(usageParts, "[OPTIONS]")
			break
		}
	}

	for _, arg := range args {
		if arg.positional {
			continue
		}

		var argSyntax string
		if arg.long != "" {
			argSyntax = fmt.Sprintf("--%s <%s>", arg.long, arg.name)
		} else if arg.short != "" {
			argSyntax = fmt.Sprintf("-%s <%s>", arg.short, arg.name)
		} else {
			developerErr("Either long or short name must be specified: " + arg.name)
		}

		if arg.kind == reflect.Slice {
			argSyntax = argSyntax + "..."
		}

		if arg.mandatory {
			usageParts = append(usageParts, argSyntax)
		}
	}

	// Add positional arguments
	for _, arg := range args {
		if arg.positional {
			usagePart := ""
			if arg.mandatory {
				usagePart = "<" + arg.name + ">"
			} else {
				usagePart = "[" + arg.name + "]"
			}
			if arg.kind == reflect.Slice {
				usagePart += "..."
			}
			usageParts = append(usageParts, usagePart)
		}
	}

	cfmt.Fprintf(&buf, "#B{Usage:}\n  %s\n\n", strings.Join(usageParts, " "))

	// --- Format help sections ---

	// Determine label width
	maxLabelLen := 0
	getLabel := func(arg arg) string {
		var parts []string
		if arg.short != "" {
			parts = append(parts, "-"+arg.short)
		}
		if arg.long != "" {
			parts = append(parts, "--"+arg.long)
		}
		label := strings.Join(parts, ", ")
		if arg.kind != reflect.Bool {
			label += fmt.Sprintf(" <%s>", arg.name)
		}
		if len(label) > maxLabelLen {
			maxLabelLen = len(label)
		}
		return label
	}

	labels := make(map[string]string)
	for _, arg := range args {
		if !arg.positional {
			labels[arg.name] = getLabel(arg)
		}
	}

	// Required options
	hasRequired := false
	for _, arg := range args {
		if !arg.positional && arg.mandatory {
			if !hasRequired {
				cfmt.Fprintln(&buf, "#B{Required options:}")
				hasRequired = true
			}
			desc := arg.description
			if arg.kind == reflect.Slice {
				desc += " (can be specified multiple times)"
			}
			if arg.defaultValue != "" {
				desc += fmt.Sprintf(" (default: %s)", arg.defaultValue)
			}
			fmt.Fprintf(&buf, "  %-*s  %s\n", maxLabelLen, labels[arg.name], desc)
		}
	}
	if hasRequired {
		fmt.Fprintln(&buf)
	}

	// Optional options
	hasOptional := false
	for _, arg := range args {
		if !arg.positional && !arg.mandatory {
			if !hasOptional {
				cfmt.Fprintln(&buf, "#B{Options:}")
				hasOptional = true
			}
			additionalDesciptions := make([]string, 0)
			if arg.kind == reflect.Slice {
				additionalDesciptions = append(additionalDesciptions, "can be specified multiple times")
			}
			if arg.defaultValue != "" {
				additionalDesciptions = append(additionalDesciptions, "default: "+arg.defaultValue)
			}
			var description string
			if len(additionalDesciptions) > 0 {
				description = fmt.Sprintf("%s (%s)", arg.description, strings.Join(additionalDesciptions, ", "))
			} else {
				description = arg.description
			}
			fmt.Fprintf(&buf, "  %-*s  %s\n", maxLabelLen, labels[arg.name], description)
		}
	}
	if hasOptional {
		fmt.Fprintln(&buf)
	}

	// Positional arguments
	hasPositional := false
	for _, arg := range args {
		if arg.positional {
			if !hasPositional {
				cfmt.Fprintln(&buf, "#B{Positional arguments:}")
				hasPositional = true
			}
			additionalDesciptions := make([]string, 0)
			if arg.mandatory {
				additionalDesciptions = append(additionalDesciptions, "required")
			}
			if arg.kind == reflect.Slice {
				additionalDesciptions = append(additionalDesciptions, "can be specified multiple times")
			}
			if arg.defaultValue != "" {
				additionalDesciptions = append(additionalDesciptions, "default: "+arg.defaultValue)
			}
			var description string
			if len(additionalDesciptions) > 0 {
				description = fmt.Sprintf("%s (%s)", arg.description, strings.Join(additionalDesciptions, ", "))
			} else {
				description = arg.description
			}
			fmt.Fprintf(&buf, "  %-*s  %s\n", maxLabelLen, arg.name, description)
		}
	}
	if hasPositional {
		fmt.Fprintln(&buf)
	}

	if example != "" {
		cfmt.Fprint(&buf, "#B{Example:}\n")
		lines := strings.Split(example, "\n")
		for _, line := range lines {
			fmt.Fprintf(&buf, "  %s\n", line)
		}
	}

	fmt.Fprint(w, strings.TrimSpace(buf.String())+"\n")
}

func developerErr(msg string) {
	panic(developerError{msg})
}

func userErr(msg string) {
	panic(userError{msg})
}

func toKebabCase(s string) string {
	allUpper := true
	for _, r := range s {
		if unicode.IsLower(r) {
			allUpper = false
			break
		}
	}
	if allUpper {
		return strings.ToLower(s)
	}

	re := regexp.MustCompile(`([A-Z][a-z]*)`)
	words := re.FindAllString(s, -1)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "-")
}

func filterArgs(args []arg, predicate func(arg arg) bool) []arg {
	filtered := make([]arg, 0)
	for _, arg := range args {
		if predicate(arg) {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}
