package env

import (
	"flag"
	"reflect"
	"strings"
)

const flagSetName = "env-flags"

func RegisterFlags(v interface{}) (*flag.FlagSet, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, ErrInvalidValue
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil, ErrInvalidValue
	}

	flags := flag.NewFlagSet(flagSetName, flag.PanicOnError)

	t := rv.Type()

	if err := registerStructFlags(flags, t, rv); err != nil {
		return nil, err
	}

	return flags, nil
}

func registerStructFlags(flags *flag.FlagSet, t reflect.Type, rv reflect.Value) error {
	for i := range rv.NumField() {
		valueField := rv.Field(i)
		typeField := t.Field(i)

		if valueField.Kind() == reflect.Struct {
			if !valueField.Addr().CanInterface() {
				continue
			}
			if err := registerStructFlags(flags, typeField.Type, valueField); err != nil {
				return err
			}
			continue
		}

		tag := typeField.Tag.Get("env")
		if tag == "" {
			continue
		}

		envTag := parseTag(tag)
		flagName := envTag.Flag
		if flagName == "" {
			flagName = toFlagName(envTag.Keys[0])
		}

		flags.String(flagName, envTag.Default, "")
	}
	return nil
}

func toFlagName(s string) string {
	nameSlice := make([]rune, 0, len(s)+3)

	var prev rune
	for _, r := range s {
		if r >= 'A' && r <= 'Z' &&
			prev >= 'a' && prev <= 'z' {
			nameSlice = append(nameSlice, '-', r)
		} else if !(r >= 'A' && r <= 'Z') &&
			!(r >= 'a' && r <= 'z') &&
			!(r >= '0' && r <= '9') {
			nameSlice = append(nameSlice, '-')
		} else {
			nameSlice = append(nameSlice, r)
		}
		prev = r
	}

	return strings.ToLower(string(nameSlice))
}

func filterUndefinedAndDups(flags *flag.FlagSet, args []string) []string {
	filteredArgs := make([]string, 0, len(args))
	seen := map[string]bool{}
	for i := 0; i < len(args); {
		arg := args[i]

		splitArg := strings.SplitN(arg, "=", 2)
		if splitArg[0][0] != '-' {
			i++
			continue
		}

		var flagName string
		if splitArg[0][1] == '-' {
			flagName = splitArg[0][2:]
		} else {
			flagName = splitArg[0][1:]
		}

		exists := flags.Lookup(flagName) != nil
		var nextArg string
		if i < len(args)-1 {
			nextArg = args[i+1]
		}

		if len(splitArg) == 1 {
			if nextArg != "" {
				if ok := seen[flagName]; exists && !ok {
					seen[flagName] = true
					filteredArgs = append(filteredArgs, arg, nextArg)
				}
				i += 2
			} else {
				i++
			}
		} else {
			if ok := seen[flagName]; exists && !ok {
				seen[flagName] = true
				filteredArgs = append(filteredArgs, arg)
			}
			i++
		}
	}
	return filteredArgs
}
