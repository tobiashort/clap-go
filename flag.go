package flag

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type flag struct {
	name  string
	type_ reflect.Type
	kind  reflect.Kind
	short string
	long  string
}

func Parse(v any) {
	if !isStructPointer(v) {
		panic("expected struct pointer")
	}

	typeOf := reflect.TypeOf(v)
	typeOfDeref := typeOf.Elem()

	flags := make([]flag, 0)

	for i := range typeOfDeref.NumField() {
		field := typeOfDeref.Field(i)

		long := strings.ToLower(field.Name)
		short := string(strings.ToLower(field.Name)[0])

		tag := field.Tag.Get("flag")
		if tag != "" {
			for tagValue := range strings.SplitSeq(tag, ",") {
				if strings.HasPrefix(tagValue, "short=") {
					short = strings.Split(tagValue, "=")[1]
				} else if strings.HasPrefix(tagValue, "long=") {
					long = strings.Split(tagValue, "=")[1]
				} else {
					panic("unkown tag value: " + tagValue)
				}
			}
		}

		flags = append(flags, flag{
			name:  field.Name,
			type_: field.Type,
			kind:  field.Type.Kind(),
			long:  long,
			short: short,
		})
	}

	checkForNameCollisions(flags)

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			// long
			name := arg[2:]
			flag, ok := getFlagByLongName(flags, name)
			if !ok {
				panic("unknown flag: " + name)
			}
			if flag.kind == reflect.Bool {
				setBool(v, flag.name, true)
			} else if flag.kind == reflect.String {
				parseString(i, v, flag.name)
				i++
			} else if flag.kind == reflect.Int {
				parseInt(i, v, flag.name)
				i++
			} else {
				panic(fmt.Sprintf("not implemented flag kind: %+v", flag))
			}
		} else if strings.HasPrefix(arg, "-") {
			// short
			name := arg[1:]
			flag, ok := getFlagByShortName(flags, name)
			if !ok {
				panic("unknown flag: " + name)
			}
			if flag.kind == reflect.Bool {
				setBool(v, flag.name, true)
			} else if flag.kind == reflect.String {
				parseString(i, v, flag.name)
				i++
			} else if flag.kind == reflect.Int {
				parseInt(i, v, flag.name)
				i++
			} else {
				panic(fmt.Sprintf("not implemented flag kind: %+v", flag))
			}
		} else {
			// positional
			panic("not implemented")
		}
	}
}

func parseString(i int, v any, name string) {
	if i+1 > len(os.Args) {
		panic("missing value for: " + name)
	}
	value := os.Args[i+1]
	setString(v, name, value)
}

func parseInt(i int, v any, name string) {
	if i+1 > len(os.Args) {
		panic("missing value for: " + name)
	}
	value := os.Args[i+1]
	setInt(v, name, value)
}

func isStructPointer(v any) bool {
	t := reflect.TypeOf(v)
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

func setInt(obj any, name string, val string) {
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	reflect.ValueOf(obj).Elem().FieldByName(name).SetInt(int64(i))
}

func setBool(obj any, name string, val bool) {
	reflect.ValueOf(obj).Elem().FieldByName(name).SetBool(val)
}

func setString(obj any, name string, val string) {
	reflect.ValueOf(obj).Elem().FieldByName(name).SetString(val)
}

func checkForNameCollisions(flags []flag) {
	seen := make(map[string]bool)
	for _, flag := range flags {
		_, exists := seen[flag.long]
		if !exists {
			seen[flag.long] = true
		} else {
			panic(fmt.Sprintf("flag name collision: %s: %s", flag.name, flag.long))
		}
		_, exists = seen[flag.short]
		if !exists {
			seen[flag.short] = true
		} else {
			panic(fmt.Sprintf("flag name collision: %s: %s", flag.name, flag.short))
		}
	}
}
