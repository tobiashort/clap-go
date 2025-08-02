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
}

func Parse(strct any) {
	if !isStructPointer(strct) {
		panic("expected struct pointer")
	}

	strctType := reflect.TypeOf(strct).Elem()

	allFlags := make([]flag, 0)

	for i := range strctType.NumField() {
		field := strctType.Field(i)

		var (
			long          = strings.ToLower(field.Name)
			short         = string(strings.ToLower(field.Name)[0])
			conflictsWith = make([]string, 0)
			mandatory     = false
			positional    = false
			description   = ""
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
				} else if strings.HasPrefix(tagValue, "description=") {
					description = strings.Split(tagValue, "=")[1]
				} else if tagValue == "mandatory" {
					mandatory = true
				} else if tagValue == "positional" {
					positional = true
				} else {
					panic("unkown tag value: " + tagValue)
				}
			}
		}

		allFlags = append(allFlags, flag{
			name:          field.Name,
			type_:         field.Type,
			kind:          field.Type.Kind(),
			long:          long,
			short:         short,
			conflictsWith: conflictsWith,
			mandatory:     mandatory,
			positional:    positional,
			description:   description,
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

	allFlags = append(allFlags, implicitHelpFlag)

	checkForNameCollisions(allFlags)

	nonPositionalFlags := make([]flag, 0)

	positionalFlagIndex := 0
	positionalFlags := make([]flag, 0)
	for _, flag := range allFlags {
		if flag.positional {
			positionalFlags = append(positionalFlags, flag)
		}
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			long := arg[2:]
			if long == "help" {
				printHelp(allFlags, os.Stdout)
				os.Exit(0)
			}
			flag, ok := getFlagByLongName(allFlags, long)
			if !ok {
				panic("unknown flag: --" + long)
			} else {
				nonPositionalFlags = append(nonPositionalFlags, flag)
			}
			i = parseNonPositional(flag, strct, i)
		} else if strings.HasPrefix(arg, "-") {
			short := arg[1:]
			if short == "h" {
				printHelp(allFlags, os.Stdout)
				os.Exit(0)
			}
			flag, ok := getFlagByShortName(allFlags, short)
			if !ok {
				panic("unknown flag: -" + short)
			} else {
				nonPositionalFlags = append(nonPositionalFlags, flag)
			}
			i = parseNonPositional(flag, strct, i)
		} else {
			if positionalFlagIndex >= len(positionalFlags) {
				panic("too many arguments")
			} else {
				positionalFlag := positionalFlags[positionalFlagIndex]
				i = parsePositional(positionalFlag, strct, i)
				positionalFlagIndex++
			}
		}
	}

	checkForConflicts(nonPositionalFlags)
	checkForMissingMandatoryFlags(allFlags, nonPositionalFlags, positionalFlags)
	checkForMultipleUse(nonPositionalFlags)
}

func parseNonPositional(flag flag, strct any, index int) int {
	if flag.kind == reflect.Bool {
		setBool(strct, flag.name, true)
		return index
	} else if flag.kind == reflect.String {
		val := parseNonPositionalString(index, flag.name)
		setString(strct, flag.name, val)
		return index + 1
	} else if flag.kind == reflect.Int {
		val := parseNonPositionalInt(index, flag.name)
		setInt(strct, flag.name, val)
		return index + 1
	} else if flag.kind == reflect.Float64 {
		val := parseNonPositionalFloat(index, flag.name)
		setFloat(strct, flag.name, val)
		return index + 1
	} else if flag.kind == reflect.Slice {
		innerKind := flag.type_.Elem().Kind()
		var val any
		if innerKind == reflect.String {
			val = parseNonPositionalString(index, flag.name)
		} else if innerKind == reflect.Int {
			val = parseNonPositionalInt(index, flag.name)
		} else if innerKind == reflect.Float64 {
			val = parseNonPositionalFloat(index, flag.name)
		} else {
			panic("not implemented flag kind []" + innerKind.String())
		}
		addToSlice(strct, flag.name, val)
		return index + 1
	} else {
		panic(fmt.Sprintf("not implemented flag kind: %v", flag.kind))
	}
}

func parsePositional(flag flag, strct any, index int) int {
	if flag.kind == reflect.String {
		val := os.Args[index]
		setString(strct, flag.name, val)
	} else if flag.kind == reflect.Int {
		arg := os.Args[index]
		val, err := strconv.Atoi(arg)
		if err != nil {
			panic("value is not an int: " + arg)
		}
		setInt(strct, flag.name, val)
	} else if flag.kind == reflect.Float64 {
		arg := os.Args[index]
		val, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			panic("value is not a float: " + arg)
		}
		setFloat(strct, flag.name, val)
	} else {
		panic(fmt.Sprintf("not implemented flag kind: %v", flag.kind))
	}
	return index
}

func parseNonPositionalString(index int, name string) string {
	if index+1 >= len(os.Args) {
		panic("missing value for: " + name)
	}
	val := os.Args[index+1]
	return val
}

func parseNonPositionalInt(index int, name string) int {
	if index+1 >= len(os.Args) {
		panic("missing value for: " + name)
	}
	arg := os.Args[index+1]
	val, err := strconv.Atoi(arg)
	if err != nil {
		panic("value is not an int: " + arg)
	}
	return val
}

func parseNonPositionalFloat(index int, name string) float64 {
	if index+1 >= len(os.Args) {
		panic("missing value for: " + name)
	}
	arg := os.Args[index+1]
	val, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		panic("value is not a float: " + arg)
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
	seen := make(map[string]flag)
	for _, flag := range flags {
		if flag.positional {
			continue
		}
		existing, exists := seen[flag.long]
		if !exists {
			seen[flag.long] = flag
		} else {
			panic(fmt.Sprintf("flag name collision: %s (--%s) with %s (--%s)", flag.name, flag.long, existing.name, existing.long))
		}
		existing, exists = seen[flag.short]
		if !exists {
			seen[flag.short] = flag
		} else {
			panic(fmt.Sprintf("flag name collision: %s (-%s) with %s (-%s)", flag.name, flag.short, existing.name, existing.short))
		}
	}
}

func checkForConflicts(providedFlags []flag) {
	for _, outerFlag := range providedFlags {
		for _, inConflict := range outerFlag.conflictsWith {
			for _, innerFlag := range providedFlags {
				if innerFlag.name == inConflict {
					panic(fmt.Sprintf("conflicting flags: --%s (-%s), --%s (-%s)", outerFlag.long, outerFlag.short, innerFlag.long, innerFlag.short))
				}
			}
		}
	}
}

func checkForMissingMandatoryFlags(flags []flag, nonPositionalFlags []flag, positionalFlags []flag) {
	providedFlags := make([]flag, 0)
	for _, nonPositionalFlag := range nonPositionalFlags {
		providedFlags = append(providedFlags, nonPositionalFlag)
	}
	for _, positionalFlag := range positionalFlags {
		providedFlags = append(providedFlags, positionalFlag)
	}

outer:
	for _, flag := range flags {
		if flag.mandatory {
			for _, providedFlag := range providedFlags {
				if providedFlag.name == flag.name {
					continue outer
				}
			}
			if flag.positional {
				panic(fmt.Sprintf("missing mandatory positional flag: %s", flag.name))
			} else {
				panic(fmt.Sprintf("missing mandatory flag: --%s (-%s)", flag.long, flag.short))
			}
		}
	}
}

func checkForMultipleUse(providedFlags []flag) {
	seen := make(map[string]bool)
	for _, flag := range providedFlags {
		_, exists := seen[flag.name]
		if !exists {
			seen[flag.name] = true
		} else {
			if flag.kind != reflect.Slice {
				panic(fmt.Sprintf("multiple use of flag --%s (-%s)", flag.long, flag.short))
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

	// Add all options (required and optional)
	for _, f := range flags {
		if f.positional {
			continue
		}

		var flagSyntax string
		long := "--" + f.long
		if f.kind == reflect.Slice {
			flagSyntax = long + " ..."
		} else {
			flagSyntax = long
		}

		if f.mandatory {
			usageParts = append(usageParts, flagSyntax)
		} else {
			usageParts = append(usageParts, "["+flagSyntax+"]")
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
		if f.short != "" {
			parts = append(parts, "-"+f.short)
		}
		if f.long != "" {
			parts = append(parts, "--"+f.long)
		}
		label := strings.Join(parts, ", ")
		if label == "" {
			label = f.name
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
