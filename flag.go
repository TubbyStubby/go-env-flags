// Copyright 2025 TubbyStubby.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package env

import (
	"flag"
	"fmt"
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

	flags := flag.NewFlagSet(flagSetName, flag.ExitOnError)

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
		description := generateDescription(envTag)

		if flagName != "" {
			flags.String(flagName, envTag.Default, description)
		}

		for _, envKeyNames := range envTag.Keys {
			flagName = toFlagName(envKeyNames)
			if flags.Lookup(flagName) == nil {
				flags.String(flagName, envTag.Default, description)
			}
		}
	}
	return nil
}

func generateDescription(t tag) string {
	var parts []string

	if t.Desc != "" {
		parts = append(parts, t.Desc)
	}

	if len(t.Keys) > 0 {
		parts = append(parts, fmt.Sprintf("Environment: %s", strings.Join(t.Keys, ", ")))
	}

	if t.Default != "" {
		parts = append(parts, fmt.Sprintf("Default: %s", t.Default))
	}

	if t.Required {
		parts = append(parts, "Required: true")
	}

	return strings.Join(parts, ". ")
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

		if len(arg) == 0 || arg[0] != '-' {
			i++
			continue
		}

		splitArg := strings.SplitN(arg, "=", 2)

		var flagName string
		if strings.HasPrefix(splitArg[0], "--") {
			flagName = splitArg[0][2:]
		} else {
			flagName = splitArg[0][1:]
		}

		exists := flags.Lookup(flagName) != nil || flagName == "help" || flagName == "h"
		var nextArg string
		if i < len(args)-1 {
			nextArg = args[i+1]
		}

		if len(splitArg) == 1 {
			if ok := seen[flagName]; exists && !ok {
				seen[flagName] = true
				filteredArgs = append(filteredArgs, arg, nextArg)
			}
			i += 2
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

func isFlagSet(flags *flag.FlagSet, name string) bool {
	fSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			fSet = true
		}
	})
	return fSet
}
