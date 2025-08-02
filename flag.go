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
}

func Parse(strct any) {
	if !isStructPointer(strct) {
		panic("expected struct pointer")
	}

	strctType := reflect.TypeOf(strct).Elem()

	flags := make([]flag, 0)

	for i := range strctType.NumField() {
		field := strctType.Field(i)

		long := strings.ToLower(field.Name)
		short := string(strings.ToLower(field.Name)[0])
		conflictsWith := make([]string, 0)
		mandatory := false

		tag := field.Tag.Get("flag")
		if tag != "" {
			for tagValue := range strings.SplitSeq(tag, ",") {
				if strings.HasPrefix(tagValue, "short=") {
					short = strings.Split(tagValue, "=")[1]
				} else if strings.HasPrefix(tagValue, "long=") {
					long = strings.Split(tagValue, "=")[1]
				} else if strings.HasPrefix(tagValue, "conflicts-with=") {
					conflictsWith = strings.Split(strings.Split(tagValue, "=")[1], ",")
				} else if tagValue == "mandatory" {
					mandatory = true
				} else {
					panic("unkown tag value: " + tagValue)
				}
			}
		}

		flags = append(flags, flag{
			name:          field.Name,
			type_:         field.Type,
			kind:          field.Type.Kind(),
			long:          long,
			short:         short,
			conflictsWith: conflictsWith,
			mandatory:     mandatory,
		})
	}

	checkForNameCollisions(flags)

	providedFlags := make([]flag, 0)

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			long := arg[2:]
			flag, ok := getFlagByLongName(flags, long)
			if !ok {
				panic("unknown flag: --" + long)
			} else {
				providedFlags = append(providedFlags, flag)
			}
			i = parseNonPositional(flag, strct, i)
		} else if strings.HasPrefix(arg, "-") {
			short := arg[1:]
			flag, ok := getFlagByShortName(flags, short)
			if !ok {
				panic("unknown flag: -" + short)
			} else {
				providedFlags = append(providedFlags, flag)
			}
			i = parseNonPositional(flag, strct, i)
		} else {
			// positional
			panic("not implemented")
		}
	}

	checkForConflicts(providedFlags)
	checkForMissingMandatoryFlags(flags, providedFlags)
	checkForMultipleUse(providedFlags)
}

func parseNonPositional(flag flag, strct any, index int) int {
	if flag.kind == reflect.Bool {
		setBool(strct, flag.name, true)
		return index
	} else if flag.kind == reflect.String {
		val := parseString(index, flag.name)
		setString(strct, flag.name, val)
		return index + 1
	} else if flag.kind == reflect.Int {
		val := parseInt(index, flag.name)
		setInt(strct, flag.name, val)
		return index + 1
	} else if flag.kind == reflect.Float64 {
		val := parseFloat(index, flag.name)
		setFloat(strct, flag.name, val)
		return index + 1
	} else if flag.kind == reflect.Slice {
		innerKind := flag.type_.Elem().Kind()
		var val any
		if innerKind == reflect.String {
			val = parseString(index, flag.name)
		} else if innerKind == reflect.Int {
			val = parseInt(index, flag.name)
		} else if innerKind == reflect.Float64 {
			val = parseFloat(index, flag.name)
		} else {
			panic("not implemented flag kind []" + innerKind.String())
		}
		addToSlice(strct, flag.name, val)
		return index + 1
	} else {
		panic(fmt.Sprintf("not implemented flag kind: %+v", flag))
	}
}

func parseString(i int, name string) string {
	if i+1 > len(os.Args) {
		panic("missing value for: " + name)
	}
	value := os.Args[i+1]
	return value
}

func parseInt(i int, name string) int {
	if i+1 > len(os.Args) {
		panic("missing value for: " + name)
	}
	value := os.Args[i+1]
	i, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return i
}

func parseFloat(i int, name string) float64 {
	if i+1 > len(os.Args) {
		panic("missing value for: " + name)
	}
	value := os.Args[i+1]
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(err)
	}
	return f
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
	seen := make(map[string]bool)
	for _, flag := range flags {
		_, exists := seen[flag.long]
		if !exists {
			seen[flag.long] = true
		} else {
			panic(fmt.Sprintf("flag name collision: %s: --%s", flag.name, flag.long))
		}
		_, exists = seen[flag.short]
		if !exists {
			seen[flag.short] = true
		} else {
			panic(fmt.Sprintf("flag name collision: %s: -%s", flag.name, flag.short))
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
