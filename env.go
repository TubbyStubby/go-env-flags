// Copyright 2018 Netflix, Inc.
// Copyright 2025 TubbyStubby.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package env provides an `env` struct field tag to marshal and unmarshal
// environment variables.
package env

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	// tagKeyDefault is the key used in the struct field tag to specify a default
	tagKeyDefault = "default"
	// tagKeyRequired is the key used in the struct field tag to specify that the
	// field is required
	tagKeyRequired = "required"
	// tagKeySeparator is the key used in the struct field tag to specify a
	// separator for slice fields
	tagKeySeparator = "separator"
	// tagKeyFlag is the key used in the struct field tag to specify a different
	// name for the env flag
	tagKeyFlag = "flag"
	// tagKeyDesc is the key used in the struct field tag to specify a description
	// note: this only comes with flag help
	tagKeyDesc = "desc"
)

var (
	// ErrInvalidValue returned when the value passed to Unmarshal is nil or not a
	// pointer to a struct.
	ErrInvalidValue = errors.New("value must be a non-nil pointer to a struct")

	// ErrUnsupportedType returned when a field with tag "env" is unsupported.
	ErrUnsupportedType = errors.New("field is an unsupported type")

	// ErrUnexportedField returned when a field with tag "env" is not exported.
	ErrUnexportedField = errors.New("field must be exported")

	// unmarshalType is the reflect.Type element of the Unmarshaler interface
	unmarshalType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
)

// ErrMissingRequiredValue returned when a field with required=true contains no value or default
type ErrMissingRequiredValue struct {
	// Value is the type of value that is required to provide error context to
	// the caller
	Value string
}

func (e ErrMissingRequiredValue) Error() string {
	return fmt.Sprintf("value for this field is required [%s]", e.Value)
}

// Unmarshal parses an EnvSet and stores the result in the value pointed to by
// v. Fields that are matched in v will be deleted from EnvSet, resulting in
// an EnvSet with the remaining environment variables. If v is nil or not a
// pointer to a struct, Unmarshal returns an ErrInvalidValue.
//
// Fields tagged with "env" will have the unmarshalled EnvSet of the matching
// key from EnvSet. If the tagged field is not exported, Unmarshal returns
// ErrUnexportedField.
//
// If the field has a type that is unsupported, Unmarshal returns
// ErrUnsupportedType.
func Unmarshal(flags *flag.FlagSet, es EnvSet, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidValue
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrInvalidValue
	}

	t := rv.Type()
	for i := range rv.NumField() {
		valueField := rv.Field(i)
		if valueField.Kind() == reflect.Struct {
			if !valueField.Addr().CanInterface() {
				continue
			}
			if err := Unmarshal(flags, es, valueField.Addr().Interface()); err != nil {
				return err
			}
		}

		typeField := t.Field(i)
		tag := typeField.Tag.Get("env")
		if tag == "" {
			continue
		}

		if !valueField.CanSet() {
			return ErrUnexportedField
		}

		envTag := parseTag(tag)

		var envValue string
		var ok bool

		// check if any flags are set, either the flag tag or the key flags
		flagName := envTag.Flag
		if flagName != "" {
			ok = isFlagSet(flags, flagName)
			if ok {
				f := flags.Lookup(flagName)
				envValue = f.Value.String()
			}
		}
		if !ok {
			for _, envKey := range envTag.Keys {
				flagName = toFlagName(envKey)
				ok = isFlagSet(flags, flagName)
				if ok {
					f := flags.Lookup(flagName)
					envValue = f.Value.String()
					break
				}
			}
		}

		// if flag not set then check the env vars
		if !ok {
			for _, envKey := range envTag.Keys {
				envValue, ok = es[envKey]
				if ok {
					break
				}
			}
		}

		if !ok {
			if envTag.Default != "" {
				envValue = envTag.Default
			} else if envTag.Required {
				return &ErrMissingRequiredValue{Value: envTag.Keys[0]}
			} else {
				continue
			}
		}

		if err := set(typeField.Type, valueField, envValue, envTag.Separator); err != nil {
			return err
		}
		delete(es, tag)
	}

	return nil
}

func set(t reflect.Type, f reflect.Value, value, sliceSeparator string) error {
	// See if the type implements Unmarshaler and use that first,
	// otherwise, fallback to the previous logic
	var isUnmarshaler bool
	isPtr := t.Kind() == reflect.Ptr
	if isPtr {
		isUnmarshaler = t.Implements(unmarshalType) && f.CanInterface()
	} else if f.CanAddr() {
		isUnmarshaler = f.Addr().Type().Implements(unmarshalType) && f.Addr().CanInterface()
	}

	if isUnmarshaler {
		var ptr reflect.Value
		if isPtr {
			// In the pointer case, we need to create a new element to have an
			// address to point to
			ptr = reflect.New(t.Elem())
		} else {
			// And for scalars, we need the pointer to be able to modify the value
			ptr = f.Addr()
		}
		if u, ok := ptr.Interface().(Unmarshaler); ok {
			if err := u.UnmarshalEnvironmentValue(value); err != nil {
				return err
			}
			if isPtr {
				f.Set(ptr)
			}
			return nil
		}
	}

	switch t.Kind() {
	case reflect.Ptr:
		ptr := reflect.New(t.Elem())
		if err := set(t.Elem(), ptr.Elem(), value, sliceSeparator); err != nil {
			return err
		}
		f.Set(ptr)
	case reflect.String:
		f.SetString(value)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		f.SetBool(v)
	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t.PkgPath() == "time" && t.Name() == "Duration" {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}

			f.Set(reflect.ValueOf(duration))
			break
		}

		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		f.SetInt(int64(v))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		f.SetUint(v)
	case reflect.Slice:
		if sliceSeparator == "" {
			sliceSeparator = "|"
		}
		values := strings.Split(value, sliceSeparator)
		switch t.Elem().Kind() {
		case reflect.String:
			// already []string, just set directly
			f.Set(reflect.ValueOf(values))
		default:
			dest := reflect.MakeSlice(reflect.SliceOf(t.Elem()), len(values), len(values))
			for i, v := range values {
				if err := set(t.Elem(), dest.Index(i), v, sliceSeparator); err != nil {
					return err
				}
			}
			f.Set(dest)
		}
	default:
		return ErrUnsupportedType
	}

	return nil
}

// UnmarshalFromEnviron parses an EnvSet from os.Environ and stores the result
// in the value pointed to by v. Fields that weren't matched in v are returned
// in an EnvSet with the remaining environment variables. If v is nil or not a
// pointer to a struct, UnmarshalFromEnviron returns an ErrInvalidValue.
//
// Fields tagged with "env" will have the unmarshalled EnvSet of the matching
// key from EnvSet. If the tagged field is not exported, UnmarshalFromEnviron
// returns ErrUnexportedField.
//
// If the field has a type that is unsupported, UnmarshalFromEnviron returns
// ErrUnsupportedType.
func UnmarshalFromEnviron(v interface{}) (*flag.FlagSet, EnvSet, error) {
	flags, err := RegisterFlags(v)
	if err != nil {
		return nil, nil, err
	}

	filteredArgs := filterUndefinedAndDups(flags, os.Args[1:])
	err = flags.Parse(filteredArgs)
	if err != nil {
		return nil, nil, err
	}

	es, err := EnvironToEnvSet(os.Environ())
	if err != nil {
		return nil, nil, err
	}

	return flags, es, Unmarshal(flags, es, v)
}

// Marshal returns an EnvSet of v. If v is nil or not a pointer, Marshal returns
// an ErrInvalidValue.
//
// Marshal uses fmt.Sprintf to transform encountered values to its default
// string format. Values without the "env" field tag are ignored.
//
// Nested structs are traversed recursively.
func Marshal(v interface{}) (EnvSet, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, ErrInvalidValue
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil, ErrInvalidValue
	}

	es := make(EnvSet)
	t := rv.Type()
	for i := range rv.NumField() {
		valueField := rv.Field(i)
		if valueField.Kind() == reflect.Struct {
			if !valueField.Addr().CanInterface() {
				continue
			}

			nes, err := Marshal(valueField.Addr().Interface())
			if err != nil {
				return nil, err
			}

			for k, v := range nes {
				es[k] = v
			}
		}

		typeField := t.Field(i)
		tag := typeField.Tag.Get("env")
		if tag == "" {
			continue
		}

		envKeys := strings.Split(tag, ",")

		var el interface{}
		if typeField.Type.Kind() == reflect.Ptr {
			if valueField.IsNil() {
				continue
			}
			el = valueField.Elem().Interface()
		} else {
			el = valueField.Interface()
		}

		var (
			err      error
			envValue string
		)
		if m, ok := el.(Marshaler); ok {
			envValue, err = m.MarshalEnvironmentValue()
			if err != nil {
				return nil, err
			}
		} else {
			envValue = fmt.Sprintf("%v", el)
		}

		for _, envKey := range envKeys {
			// Skip keys with '=', as they represent tag options and not environment variable names.
			if strings.Contains(envKey, "=") {
				switch strings.ToLower(strings.SplitN(envKey, "=", 2)[0]) {
				case "separator", "required", "default", "flag", "desc":
					continue
				}
			}
			es[envKey] = envValue
		}
	}

	return es, nil
}

// tag is a struct used to store the parsed "env" field tag when unmarshalling.
type tag struct {
	// Keys is used to store the keys specified in the "env" field tag
	Keys []string
	// Default is used to specify a default value for the field
	Default string
	// Required is used to specify that the field is required
	Required bool
	// Separator is used to split the value of a slice field
	Separator string
	// Flag is used to provide alternative name for the env flag
	Flag string
	// Desc is used to provide a description for the field
	Desc string
}

// parseTag is used in the Unmarshal function to parse the "env" field tags
// into a tag struct for use in the set function.
func parseTag(tagString string) tag {
	var t tag
	envKeys := strings.Split(tagString, ",")
	for _, key := range envKeys {
		if !strings.Contains(key, "=") {
			t.Keys = append(t.Keys, key)
			continue
		}
		keyData := strings.SplitN(key, "=", 2)
		switch strings.ToLower(keyData[0]) {
		case tagKeyDefault:
			t.Default = keyData[1]
		case tagKeyRequired:
			t.Required = strings.ToLower(keyData[1]) == "true"
		case tagKeySeparator:
			t.Separator = keyData[1]
		case tagKeyFlag:
			t.Flag = keyData[1]
		case tagKeyDesc:
			t.Desc = keyData[1]
		default:
			// just ignoring unsupported keys
			continue
		}
	}
	return t
}
