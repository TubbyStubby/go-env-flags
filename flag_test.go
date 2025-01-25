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
	"reflect"
	"testing"
	"time"
)

func TestFlagUnmarshal(t *testing.T) {
	t.Parallel()
	var (
		environ = map[string]string{}
		args    = []string{
			"-home=/home/test",
			"-workspace=/mnt/builds/slave/workspace/test",
			"-extra", "extra",
			"-int", "1",
			"-uint", "4294967295",
			"-float32", "2.3",
			"-float64=4.5",
			"-bool", "true",
			"-npm-config-cache", "first",
			"-npm-config-cache", "second",
			"-type-duration", "5s",
		}
		validStruct ValidStruct
	)
	flags, err := RegisterFlags(&validStruct)
	if err != nil {
		t.Errorf("Expected no error while register but got '%s'", err)
	}

	filteredArgs := filterUndefinedAndDups(flags, args)
	if err := flags.Parse(filteredArgs); err != nil {
		t.Errorf("Expected flag set to parse filtered args but got '%s'", err)
	}

	if err := Unmarshal(flags, environ, &validStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Home != "/home/test" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "/home/test", validStruct.Home)
	}

	if validStruct.Jenkins.Workspace != "/mnt/builds/slave/workspace/test" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "/mnt/builds/slave/workspace/test", validStruct.Jenkins.Workspace)
	}

	if validStruct.PointerString != nil {
		t.Errorf("Expected field value to be '%v' but got '%v'", nil, validStruct.PointerString)
	}

	if validStruct.Extra != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", validStruct.Extra)
	}

	if validStruct.Int != 1 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 1, validStruct.Int)
	}

	if validStruct.Uint != 4294967295 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 4294967295, validStruct.Uint)
	}

	if validStruct.Float32 != 2.3 {
		t.Errorf("Expected field value to be '%f' but got '%f'", 2.3, validStruct.Float32)
	}

	if validStruct.Float64 != 4.5 {
		t.Errorf("Expected field value to be '%f' but got '%f'", 4.5, validStruct.Float64)
	}

	if validStruct.Bool != true {
		t.Errorf("Expected field value to be '%t' but got '%t'", true, validStruct.Bool)
	}

	if validStruct.MultipleTags != "first" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "first", validStruct.MultipleTags)
	}

	if validStruct.Duration != 5*time.Second {
		t.Errorf("Expected field value to be '%s' but got '%s'", "5s", validStruct.Duration)
	}
}

func TestFlagPriority(t *testing.T) {
	t.Parallel()
	var (
		environ     = map[string]string{"HOME": "/home/bad"}
		args        = []string{"-home=/home/test"}
		validStruct ValidStruct
	)
	flags, err := RegisterFlags(&validStruct)
	if err != nil {
		t.Errorf("Expected no error while register but got '%s'", err)
	}

	filteredArgs := filterUndefinedAndDups(flags, args)
	if err := flags.Parse(filteredArgs); err != nil {
		t.Errorf("Expected flag set to parse filtered args but got '%s'", err)
	}

	if err := Unmarshal(flags, environ, &validStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Home != "/home/test" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "/home/test", validStruct.Home)
	}
}

func TestFlagUnmarshalPointer(t *testing.T) {
	t.Parallel()
	var (
		environ = map[string]string{}
		args    = []string{
			"-pointer-string", "",
			"-pointer-int", "1",
			"-pointer-uint", "4294967295",
			"-pointer-pointer-string", "",
		}
		validStruct ValidStruct
	)

	flags, err := RegisterFlags(&validStruct)
	if err != nil {
		t.Errorf("Expected no error while register but got '%s'", err)
	}

	filteredArgs := filterUndefinedAndDups(flags, args)
	if err := flags.Parse(filteredArgs); err != nil {
		t.Errorf("Expected flag set to parse filtered args but got '%s'", err)
	}

	if err := Unmarshal(flags, environ, &validStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.PointerString == nil {
		t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
	} else if *validStruct.PointerString != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", *validStruct.PointerString)
	}

	if validStruct.PointerInt == nil {
		t.Errorf("Expected field value to be '%d' but got '%v'", 1, nil)
	} else if *validStruct.PointerInt != 1 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 1, *validStruct.PointerInt)
	}

	if validStruct.PointerUint == nil {
		t.Errorf("Expected field value to be '%d' but got '%v'", 4294967295, nil)
	} else if *validStruct.PointerUint != 4294967295 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 4294967295, *validStruct.PointerUint)
	}

	if validStruct.PointerPointerString == nil {
		t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
	} else {
		if *validStruct.PointerPointerString == nil {
			t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
		} else if **validStruct.PointerPointerString != "" {
			t.Errorf("Expected field value to be '%s' but got '%s'", "", **validStruct.PointerPointerString)
		}
	}

	if validStruct.PointerMissing != nil {
		t.Errorf("Expected field value to be '%v' but got '%s'", nil, *validStruct.PointerMissing)
	}
}

// TODO: add support for custom unmarshal

func TestFlagUnmarshalSlice(t *testing.T) {
	t.Parallel()
	var (
		environ = map[string]string{}
		args    = []string{
			"-string", "separate|values",
			"-int", "1|2",
			"-int64", "3|4",
			"-duration", "60s|70h",
			"-bool", "true|false",
			"-kv=k1=v1|k2=v2",
			"-separator", "1&2", // struct has `separator=&`
		}
		iterValStruct IterValuesStruct
	)

	flags, err := RegisterFlags(&iterValStruct)
	if err != nil {
		t.Errorf("Expected no error while register but got '%s'", err)
	}

	filteredArgs := filterUndefinedAndDups(flags, args)
	if err := flags.Parse(filteredArgs); err != nil {
		t.Errorf("Expected flag set to parse filtered args but got '%s'", err)
	}

	if err := Unmarshal(flags, environ, &iterValStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	testCases := [][]interface{}{
		{iterValStruct.StringSlice, []string{"separate", "values"}},
		{iterValStruct.IntSlice, []int{1, 2}},
		{iterValStruct.Int64Slice, []int64{3, 4}},
		{iterValStruct.DurationSlice, []time.Duration{time.Second * 60, time.Hour * 70}},
		{iterValStruct.BoolSlice, []bool{true, false}},
		{iterValStruct.KVStringSlice, []string{"k1=v1", "k2=v2"}},
		{iterValStruct.WithSeparator, []int{1, 2}},
	}
	for _, testCase := range testCases {
		if !reflect.DeepEqual(testCase[0], testCase[1]) {
			t.Errorf("Expected field value to be '%v' but got '%v'", testCase[1], testCase[0])
		}
	}
}

func TestFlagUnmarshalDefaultValues(t *testing.T) {
	t.Parallel()
	var (
		environ            = map[string]string{}
		args               = []string{"-present", "youFoundMe"}
		defaultValueStruct DefaultValueStruct
	)

	flags, err := RegisterFlags(&defaultValueStruct)
	if err != nil {
		t.Errorf("Expected no error while register but got '%s'", err)
	}

	filteredArgs := filterUndefinedAndDups(flags, args)
	if err := flags.Parse(filteredArgs); err != nil {
		t.Errorf("Expected flag set to parse filtered args but got '%s'", err)
	}

	if err := Unmarshal(flags, environ, &defaultValueStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	testCases := [][]interface{}{
		{defaultValueStruct.DefaultInt, 7},
		{defaultValueStruct.DefaultUint, uint(4294967295)},
		{defaultValueStruct.DefaultFloat32, float32(8.9)},
		{defaultValueStruct.DefaultFloat64, 10.11},
		{defaultValueStruct.DefaultBool, true},
		{defaultValueStruct.DefaultString, "found"},
		{defaultValueStruct.DefaultKeyValueString, "key=value"},
		{defaultValueStruct.DefaultDuration, 5 * time.Second},
		{defaultValueStruct.DefaultStringSlice, []string{"separate", "values"}},
		{defaultValueStruct.DefaultSliceWithSeparator, []string{"separate", "values"}},
		{defaultValueStruct.DefaultRequiredSlice, []string{"other", "things"}},
		{defaultValueStruct.DefaultWithOptionsMissing, "present"},
		{defaultValueStruct.DefaultWithOptionsPresent, "youFoundMe"},
	}
	for _, testCase := range testCases {
		if !reflect.DeepEqual(testCase[0], testCase[1]) {
			t.Errorf("Expected field value to be '%v' but got '%v'", testCase[1], testCase[0])
		}
	}
}

func TestFlagUnmarshalRequiredValues(t *testing.T) {
	t.Parallel()
	var (
		environ              = make(map[string]string)
		requiredValuesStruct RequiredValueStruct
	)
	flags, err := RegisterFlags(&requiredValuesStruct)
	if err != nil {
		t.Errorf("Expected no error while register but got '%s'", err)
	}

	// Try missing REQUIRED_VAL and REQUIRED_VAL_MORE
	err = Unmarshal(flags, environ, &requiredValuesStruct)
	if err == nil {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}
	errMissing := ErrMissingRequiredValue{Value: "REQUIRED_VAL"}
	if err.Error() != errMissing.Error() {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}

	// Fill REQUIRED_VAL and retry REQUIRED_VAL_MORE
	args := []string{"-required-val", "required"}
	if err := flags.Parse(args); err != nil {
		t.Errorf("Expected flag set to parse filtered args but got '%s'", err)
	}
	err = Unmarshal(flags, environ, &requiredValuesStruct)
	if err == nil {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}
	errMissing = ErrMissingRequiredValue{Value: "REQUIRED_VAL_MORE"}
	if err.Error() != errMissing.Error() {
		t.Errorf("Expected error 'ErrMissingRequiredValue' but got '%s'", err)
	}

	args = []string{"-required-val-more", "required"}
	if err := flags.Parse(args); err != nil {
		t.Errorf("Expected flag set to parse filtered args but got '%s'", err)
	}
	if err = Unmarshal(flags, environ, &requiredValuesStruct); err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}
	if requiredValuesStruct.Required != "required" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "required", requiredValuesStruct.Required)
	}
	if requiredValuesStruct.RequiredWithDefault != "myValue" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "myValue", requiredValuesStruct.RequiredWithDefault)
	}
}

// TODO: do we need marshal for flags?
