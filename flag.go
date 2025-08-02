package flag

import (
	"fmt"
	"os"
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

	checkForNameCollisions(allFlags)

	providedFlags := make([]flag, 0)

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
			flag, ok := getFlagByLongName(allFlags, long)
			if !ok {
				panic("unknown flag: --" + long)
			} else {
				providedFlags = append(providedFlags, flag)
			}
			i = parseNonPositional(flag, strct, i)
		} else if strings.HasPrefix(arg, "-") {
			short := arg[1:]
			flag, ok := getFlagByShortName(allFlags, short)
			if !ok {
				panic("unknown flag: -" + short)
			} else {
				providedFlags = append(providedFlags, flag)
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

	checkForConflicts(providedFlags)
	checkForMissingMandatoryFlags(allFlags, providedFlags)
	checkForMultipleUse(providedFlags)
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

func checkForMissingMandatoryFlags(flags []flag, providedFlags []flag) {
outer:
	for _, flag := range flags {
		if flag.mandatory {
			for _, providedFlag := range providedFlags {
				if providedFlag.name == flag.name {
					continue outer
				}
			}
			panic(fmt.Sprintf("missing mandatory flag: --%s (-%s)", flag.long, flag.short))
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
			sb.WriteByte(ch)
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
