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
	for i, r := range s {
		if i == 0 {
			nameSlice = append(nameSlice, r)
		} else if r >= 'A' && r <= 'Z' &&
			prev >= 'a' && prev <= 'z' {
			nameSlice = append(nameSlice, '-', r)
		} else if !(r >= 'A' && r <= 'Z') &&
			!(r >= 'a' && r <= 'z') &&
			!(r >= '0' && r <= '9') {
			nameSlice = append(nameSlice, '-')
		} else if r >= '0' && r <= '9' &&
			!(prev >= '0' && prev <= '9') {
			nameSlice = append(nameSlice, '-', r)
		} else {
			nameSlice = append(nameSlice, r)
		}
		prev = r
	}

	return strings.ToLower(string(nameSlice))
}

func filterUndefined(flags *flag.FlagSet, args []string) []string {
	filteredArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); {
		s := args[i]

		s2 := strings.SplitN(s, "=", 2)
		if s2[0][0] != '-' {
			continue
		}

		var f string
		if s2[0][1] == '-' {
			f = s2[0][2:]
		} else {
			f = s2[0][1:]
		}

		exists := flags.Lookup(f) != nil
		var nextS string
		if i < len(args)-1 {
			nextS = args[i+1]
		}

		if len(s2) == 1 {
			if nextS == "" || nextS[0] == '-' {
				i += 1
			} else if exists {
				filteredArgs = append(filteredArgs, s, nextS)
				i += 2
			}
		} else {
			if exists {
				filteredArgs = append(filteredArgs, s)
			}
			i += 1
		}
	}
	return filteredArgs
}
