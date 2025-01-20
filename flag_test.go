package env

import (
	"testing"
	"time"
)

func TestFlagUnmarshal(t *testing.T) {
	t.Parallel()
	var (
		environ = map[string]string{}
		args    = []string{
			"-home", "/home/test",
			"-workspace", "/mnt/builds/slave/workspace/test",
			"-extra", "extra",
			"-int", "1",
			"-uint", "4294967295",
			"-float32", "2.3",
			"-float64", "4.5",
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
