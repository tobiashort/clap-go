package clap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

var (
	prog        string = filepath.Base(os.Args[0])
	description string = ""
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
			long          = strings.ToLower(field.Name)
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

	checkForNameCollisions(programArgs)

	programPositionalArgs := make([]arg, 0)
	for _, arg := range programArgs {
		if arg.positional {
			programPositionalArgs = append(programPositionalArgs, arg)
		}
	}

	givenNonPositionalArgs := make([]arg, 0)
	givenPositionalArgs := make([]arg, 0)
	positionalArgIndex := 0

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
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
		} else if strings.HasPrefix(arg, "-") {
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
			if positionalArgIndex >= len(programPositionalArgs) {
				userErr("too many arguments")
			} else {
				positionalArg := programPositionalArgs[positionalArgIndex]
				givenPositionalArgs = append(givenPositionalArgs, positionalArg)
				parsePositionalAtIndex(positionalArg, strct, i)
				positionalArgIndex++
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
		val := parseInt(value)
		setInt(strct, arg.name, val)
	} else if arg.kind == reflect.Float64 {
		val := parseFloat(value)
		setFloat(strct, arg.name, val)
	} else if arg.kind == reflect.Slice {
		innerKind := arg.type_.Elem().Kind()
		var val any
		if innerKind == reflect.String {
			val = value
		} else if innerKind == reflect.Int {
			val = parseInt(value)
		} else if innerKind == reflect.Float64 {
			val = parseFloat(value)
		} else {
			developerErr("not implemented argument kind []" + innerKind.String())
		}
		addToSlice(strct, arg.name, val)
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
		val, err := strconv.Atoi(value)
		if err != nil {
			developerErr("value is not an int: " + value)
		}
		setInt(strct, arg.name, val)
	} else if arg.kind == reflect.Float64 {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			developerErr("value is not a float: " + value)
		}
		setFloat(strct, arg.name, val)
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
		existing, exists := seenLong[arg.long]
		if !exists {
			seenLong[arg.long] = arg
		} else {
			developerErr(fmt.Sprintf("argument name collision: %s (--%s) with %s (--%s)", arg.name, arg.long, existing.name, existing.long))
		}
		existing, exists = seenShort[arg.short]
		if !exists {
			seenShort[arg.short] = arg
		} else {
			developerErr(fmt.Sprintf("argument name collision: %s (-%s) with %s (-%s)", arg.name, arg.short, existing.name, existing.short))
		}
	}
}

func checkForConflicts(givenNonPositionalArgs []arg) {
	for _, outerArg := range givenNonPositionalArgs {
		for _, inConflict := range outerArg.conflictsWith {
			for _, innerArg := range givenNonPositionalArgs {
				if innerArg.name == inConflict {
					developerErr(fmt.Sprintf("conflicting arguments: -%s|--%s, -%s|--%s", outerArg.short, outerArg.long, innerArg.short, innerArg.long))
				}
			}
		}
	}
}

func checkForMissingMandatoryArgs(programArgs []arg, givenNonPositionalArgs []arg, givenPositionalArgs []arg) {
	givenArgs := make([]arg, 0)
	for _, nonPositionalArg := range givenNonPositionalArgs {
		givenArgs = append(givenArgs, nonPositionalArg)
	}
	for _, positionalArg := range givenPositionalArgs {
		givenArgs = append(givenArgs, positionalArg)
	}

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

	for _, f := range args {
		if !f.mandatory {
			usageParts = append(usageParts, "[OPTIONS]")
			break
		}
	}

	for _, f := range args {
		if f.positional {
			continue
		}

		argSyntax := fmt.Sprintf("--%s <%s>", f.long, f.name)
		if f.kind == reflect.Slice {
			argSyntax = argSyntax + " ..."
		}

		if f.mandatory {
			usageParts = append(usageParts, argSyntax)
		}
	}

	// Add positional arguments
	for _, f := range args {
		if f.positional {
			if f.mandatory {
				usageParts = append(usageParts, "<"+f.name+">")
			} else {
				usageParts = append(usageParts, "["+f.name+"]")
			}
		}
	}

	fmt.Fprintf(&buf, "Usage:\n  %s\n\n", strings.Join(usageParts, " "))

	// --- Format help sections ---

	// Determine label width
	maxLabelLen := 0
	getLabel := func(f arg) string {
		var parts []string
		parts = append(parts, "-"+f.short)
		parts = append(parts, "--"+f.long)
		label := strings.Join(parts, ", ")
		if f.kind != reflect.Bool {
			label += fmt.Sprintf(" <%s>", f.name)
		}
		if len(label) > maxLabelLen {
			maxLabelLen = len(label)
		}
		return label
	}

	labels := make(map[string]string)
	for _, f := range args {
		if !f.positional {
			labels[f.name] = getLabel(f)
		}
	}

	// Required options
	hasRequired := false
	for _, f := range args {
		if !f.positional && f.mandatory {
			if !hasRequired {
				fmt.Fprintln(&buf, "Required options:")
				hasRequired = true
			}
			desc := f.description
			if f.kind == reflect.Slice {
				desc += " (can be specified multiple times)"
			}
			if f.defaultValue != "" {
				desc += fmt.Sprintf(" (default: %s)", f.defaultValue)
			}
			fmt.Fprintf(&buf, "  %-*s  %s\n", maxLabelLen, labels[f.name], desc)
		}
	}
	if hasRequired {
		fmt.Fprintln(&buf)
	}

	// Optional options
	hasOptional := false
	for _, f := range args {
		if !f.positional && !f.mandatory {
			if !hasOptional {
				fmt.Fprintln(&buf, "Options:")
				hasOptional = true
			}
			desc := f.description
			if f.kind == reflect.Slice {
				desc += " (can be specified multiple times)"
			}
			if f.defaultValue != "" {
				desc += fmt.Sprintf(" (default: %s)", f.defaultValue)
			}
			fmt.Fprintf(&buf, "  %-*s  %s\n", maxLabelLen, labels[f.name], desc)
		}
	}
	if hasOptional {
		fmt.Fprintln(&buf)
	}

	// Positional arguments
	hasPositional := false
	for _, f := range args {
		if f.positional {
			if !hasPositional {
				fmt.Fprintln(&buf, "Positional arguments:")
				hasPositional = true
			}
			desc := f.description
			if f.mandatory {
				desc += " (required)"
			}
			if f.defaultValue != "" {
				desc += fmt.Sprintf(" (default: %s)", f.defaultValue)
			}
			fmt.Fprintf(&buf, "  %-*s  %s\n", maxLabelLen, f.name, f.description)
		}
	}

	fmt.Fprint(w, strings.TrimSpace(buf.String()))
}

func developerErr(msg string) {
	panic(developerError{msg})
}

func userErr(msg string) {
	panic(userError{msg})
}
