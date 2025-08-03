package flag

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type flag struct {
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

	programFlags := make([]flag, 0)

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

		tag := field.Tag.Get("flag")
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

		programFlags = append(programFlags, flag{
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

	implicitHelpFlag := flag{
		name:        "Help",
		type_:       reflect.TypeOf(true),
		kind:        reflect.Bool,
		long:        "help",
		short:       "h",
		description: "Show this help message and exit",
	}

	programFlags = append(programFlags, implicitHelpFlag)

	checkForNameCollisions(programFlags)

	programPositionalFlags := make([]flag, 0)
	for _, flag := range programFlags {
		if flag.positional {
			programPositionalFlags = append(programPositionalFlags, flag)
		}
	}

	givenNonPositionalFlags := make([]flag, 0)
	givenPositionalFlags := make([]flag, 0)
	positionalFlagIndex := 0

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			long := arg[2:]
			if long == "help" {
				printHelp(programFlags, os.Stdout)
				os.Exit(0)
			}
			flag, ok := getFlagByLongName(programFlags, long)
			if !ok {
				userErr("unknown flag: --" + long)
			} else {
				givenNonPositionalFlags = append(givenNonPositionalFlags, flag)
			}
			i = parseNonPositionalAtIndex(flag, strct, i)
		} else if strings.HasPrefix(arg, "-") {
			shortGrouped := arg[1:]
			for _, rune := range shortGrouped {
				short := string(rune)
				if short == "h" {
					printHelp(programFlags, os.Stdout)
					os.Exit(0)
				}
				flag, ok := getFlagByShortName(programFlags, short)
				if !ok {
					userErr("unknown flag: -" + short)
				} else {
					givenNonPositionalFlags = append(givenNonPositionalFlags, flag)
				}
				i = parseNonPositionalAtIndex(flag, strct, i)
			}
		} else {
			if positionalFlagIndex >= len(programPositionalFlags) {
				userErr("too many arguments")
			} else {
				positionalFlag := programPositionalFlags[positionalFlagIndex]
				givenPositionalFlags = append(givenPositionalFlags, positionalFlag)
				parsePositionalAtIndex(positionalFlag, strct, i)
				positionalFlagIndex++
			}
		}
	}

	checkForConflicts(givenNonPositionalFlags)
	checkForMissingMandatoryFlags(programFlags, givenNonPositionalFlags, givenPositionalFlags)
	checkForMultipleUse(givenNonPositionalFlags)

outer:
	for _, flag := range programFlags {
		if flag.defaultValue == "" {
			continue
		}
		for _, givenFlag := range givenNonPositionalFlags {
			if flag.name == givenFlag.name {
				continue outer
			}
		}
		for _, givenFlag := range givenPositionalFlags {
			if flag.name == givenFlag.name {
				continue outer
			}
		}
		if flag.positional {
			parsePositional(flag, strct, flag.defaultValue)
		} else {
			parseNonPositional(flag, strct, flag.defaultValue)
		}
	}
}

func parseNonPositionalAtIndex(flag flag, strct any, index int) int {
	if flag.kind == reflect.Bool {
		parseNonPositional(flag, strct, "")
		return index
	} else {
		if index+1 >= len(os.Args) {
			userErr(fmt.Sprintf("missing value for: -%s|--%s", flag.short, flag.long))
		}
		arg := os.Args[index+1]
		parseNonPositional(flag, strct, arg)
		return index + 1
	}
}

func parseNonPositional(flag flag, strct any, arg string) {
	if flag.kind == reflect.Bool {
		setBool(strct, flag.name, true)
	} else if flag.kind == reflect.String {
		setString(strct, flag.name, arg)
	} else if flag.kind == reflect.Int {
		val := parseInt(arg)
		setInt(strct, flag.name, val)
	} else if flag.kind == reflect.Float64 {
		val := parseFloat(arg)
		setFloat(strct, flag.name, val)
	} else if flag.kind == reflect.Slice {
		innerKind := flag.type_.Elem().Kind()
		var val any
		if innerKind == reflect.String {
			val = arg
		} else if innerKind == reflect.Int {
			val = parseInt(arg)
		} else if innerKind == reflect.Float64 {
			val = parseFloat(arg)
		} else {
			developerErr("not implemented flag kind []" + innerKind.String())
		}
		addToSlice(strct, flag.name, val)
	} else {
		developerErr(fmt.Sprintf("not implemented flag kind: %v", flag.kind))
		panic("unreachable")
	}
}

func parsePositionalAtIndex(flag flag, strct any, index int) {
	arg := os.Args[index]
	parsePositional(flag, strct, arg)
}

func parsePositional(flag flag, strct any, arg string) {
	if flag.kind == reflect.String {
		setString(strct, flag.name, arg)
	} else if flag.kind == reflect.Int {
		val, err := strconv.Atoi(arg)
		if err != nil {
			developerErr("value is not an int: " + arg)
		}
		setInt(strct, flag.name, val)
	} else if flag.kind == reflect.Float64 {
		val, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			developerErr("value is not a float: " + arg)
		}
		setFloat(strct, flag.name, val)
	} else {
		developerErr(fmt.Sprintf("not implemented flag kind: %v", flag.kind))
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

func getFlagByLongName(flags []flag, name string) (flag, bool) {
	for _, flag := range flags {
		if flag.long == name {
			return flag, true
		}
	}
	return flag{}, false
}

func getFlagByShortName(flags []flag, name string) (flag, bool) {
	for _, flag := range flags {
		if flag.short == name {
			return flag, true
		}
	}
	return flag{}, false
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

func checkForNameCollisions(flags []flag) {
	seenLong := make(map[string]flag)
	seenShort := make(map[string]flag)
	for _, flag := range flags {
		if flag.positional {
			continue
		}
		existing, exists := seenLong[flag.long]
		if !exists {
			seenLong[flag.long] = flag
		} else {
			developerErr(fmt.Sprintf("flag name collision: %s (--%s) with %s (--%s)", flag.name, flag.long, existing.name, existing.long))
		}
		existing, exists = seenShort[flag.short]
		if !exists {
			seenShort[flag.short] = flag
		} else {
			developerErr(fmt.Sprintf("flag name collision: %s (-%s) with %s (-%s)", flag.name, flag.short, existing.name, existing.short))
		}
	}
}

func checkForConflicts(givenNonPositionalFlags []flag) {
	for _, outerFlag := range givenNonPositionalFlags {
		for _, inConflict := range outerFlag.conflictsWith {
			for _, innerFlag := range givenNonPositionalFlags {
				if innerFlag.name == inConflict {
					developerErr(fmt.Sprintf("conflicting flags: -%s|--%s, -%s|--%s", outerFlag.short, outerFlag.long, innerFlag.short, innerFlag.long))
				}
			}
		}
	}
}

func checkForMissingMandatoryFlags(programFlags []flag, givenNonPositionalFlags []flag, givenPositionalFlags []flag) {
	givenFlags := make([]flag, 0)
	for _, nonPositionalFlag := range givenNonPositionalFlags {
		givenFlags = append(givenFlags, nonPositionalFlag)
	}
	for _, positionalFlag := range givenPositionalFlags {
		givenFlags = append(givenFlags, positionalFlag)
	}

outer:
	for _, flag := range programFlags {
		if flag.mandatory {
			for _, givenFlag := range givenFlags {
				if givenFlag.name == flag.name {
					continue outer
				}
			}
			if flag.positional {
				userErr(fmt.Sprintf("missing mandatory positional flag: %s", flag.name))
			} else {
				userErr(fmt.Sprintf("missing mandatory flag: -%s|--%s", flag.short, flag.long))
			}
		}
	}
}

func checkForMultipleUse(givenNonPositionalFlags []flag) {
	seen := make(map[string]bool)
	for _, flag := range givenNonPositionalFlags {
		_, exists := seen[flag.name]
		if !exists {
			seen[flag.name] = true
		} else {
			if flag.kind != reflect.Slice {
				userErr(fmt.Sprintf("multiple use of flag -%s|--%s", flag.short, flag.long))
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

func printHelp(flags []flag, w io.Writer) {
	var usageParts []string
	usageParts = append(usageParts, filepath.Base(os.Args[0]))

	for _, f := range flags {
		if !f.mandatory {
			usageParts = append(usageParts, "[OPTIONS]")
			break
		}
	}

	for _, f := range flags {
		if f.positional {
			continue
		}

		flagSyntax := fmt.Sprintf("--%s <%s>", f.long, f.name)
		if f.kind == reflect.Slice {
			flagSyntax = flagSyntax + " ..."
		}

		if f.mandatory {
			usageParts = append(usageParts, flagSyntax)
		}
	}

	// Add positional arguments
	for _, f := range flags {
		if f.positional {
			if f.mandatory {
				usageParts = append(usageParts, "<"+f.name+">")
			} else {
				usageParts = append(usageParts, "["+f.name+"]")
			}
		}
	}

	fmt.Fprintf(w, "Usage:\n  %s\n\n", strings.Join(usageParts, " "))

	// --- Format help sections ---

	// Determine label width
	maxLabelLen := 0
	getLabel := func(f flag) string {
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
	for _, f := range flags {
		if !f.positional {
			labels[f.name] = getLabel(f)
		}
	}

	// Required options
	hasRequired := false
	for _, f := range flags {
		if !f.positional && f.mandatory {
			if !hasRequired {
				fmt.Fprintln(w, "Required options:")
				hasRequired = true
			}
			desc := f.description
			if f.kind == reflect.Slice {
				desc += " (can be specified multiple times)"
			}
			fmt.Fprintf(w, "  %-*s  %s\n", maxLabelLen, labels[f.name], desc)
		}
	}
	if hasRequired {
		fmt.Fprintln(w)
	}

	// Optional options
	hasOptional := false
	for _, f := range flags {
		if !f.positional && !f.mandatory {
			if !hasOptional {
				fmt.Fprintln(w, "Options:")
				hasOptional = true
			}
			desc := f.description
			if f.kind == reflect.Slice {
				desc += " (can be specified multiple times)"
			}
			fmt.Fprintf(w, "  %-*s  %s\n", maxLabelLen, labels[f.name], desc)
		}
	}
	if hasOptional {
		fmt.Fprintln(w)
	}

	// Positional arguments
	hasPositional := false
	for _, f := range flags {
		if f.positional {
			if !hasPositional {
				fmt.Fprintln(w, "Positional arguments:")
				hasPositional = true
			}
			reqMark := ""
			if f.mandatory {
				reqMark = " (required)"
			}
			fmt.Fprintf(w, "  %-*s  %s%s\n", maxLabelLen, f.name, f.description, reqMark)
		}
	}
}

func developerErr(msg string) {
	panic(developerError{msg})
}

func userErr(msg string) {
	panic(userError{msg})
}
