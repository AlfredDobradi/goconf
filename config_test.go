package goconf

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	grammar := struct {
		Test      string
		CustomKey string `yaml:"custom_key"`
		Nested    struct {
			NestedKey string
			Overwrite string `env:"CFG_NESTED_OVERWRITE"`
		}
		Typed struct {
			I    int
			DefI int `default:"1"`
			F    float64
			DefF float64 `default:"0.1"`
			B    bool
			DefB bool `default:"true"`
		}
		NestedCustom struct {
			Key string
		} `yaml:"nested_custom"`
	}{}

	configYaml := `test: "test_value"
custom_key: "test"
nested:
    nestedkey: "this is nested"
    overwrite: "don't look"
typed:
    i: 10
    f: 0.1
    b: true
nested_custom:
    key: "test"`
	source := strings.NewReader(configYaml)

	os.Setenv("CFG_NESTED_OVERWRITE", "look")

	config, err := Load(&grammar, source)
	if err != nil {
		t.Fatalf("error: loading config: %v", err)
	}

	assertions := []struct {
		label    string
		expected interface{}
		actual   interface{}
	}{
		{
			label:    "Plain root field",
			expected: "test_value",
			actual:   config.Get("test").(string),
		},
		{
			label:    "Nested field set in config",
			expected: "this is nested",
			actual:   config.Get("nested.nestedkey").(string),
		},
		{
			label:    "Nested field overwritten by env",
			expected: "look",
			actual:   config.Get("nested.overwrite").(string),
		},
		{
			label:    "Typed value: int",
			expected: 10,
			actual:   config.Get("typed.i").(int),
		},
		{
			label:    "Typed default value: int",
			expected: 1,
			actual:   config.Get("typed.defi").(int),
		},
		{
			label:    "Typed value: float64",
			expected: 0.1,
			actual:   config.Get("typed.f").(float64),
		},
		{
			label:    "Typed default value: float64",
			expected: 0.1,
			actual:   config.Get("typed.deff").(float64),
		},
		{
			label:    "Typed value: bool",
			expected: true,
			actual:   config.Get("typed.b").(bool),
		},
		{
			label:    "Typed default value: bool",
			expected: true,
			actual:   config.Get("typed.defb").(bool),
		},
		{
			label:    "Custom YAML key",
			expected: "test",
			actual:   config.Get("custom_key").(string),
		},
		{
			label:    "Custom YAML key on a non-leaf node",
			expected: "test",
			actual:   config.Get("nested_custom.key").(string),
		},
	}

	for i, a := range assertions {
		t.Logf("Assertion %d: %s", i+1, a.label)
		{
			requireEqual(t, a.expected, a.actual)
		}
	}
}

func TestLoadNotPointer(t *testing.T) {
	grammar := struct{}{}

	configYaml := `test: "test_value"`
	source := strings.NewReader(configYaml)

	expected := fmt.Sprintf("expected a pointer to a struct but got %T", grammar)

	t.Log("Given the need to fail when the passed parameter is not a pointer to the grammar")

	if _, err := Load(grammar, source); err == nil || err.Error() != expected {
		t.Fatalf("Failed: expected error %s, got %s", expected, err.Error())
	}

	t.Log("Passed")
}

func TestGet(t *testing.T) {
	grammar := struct {
		TestString string
		TestInt    int
		TestFloat  float64
		TestBool   bool
	}{}

	configYaml := `teststring: "val"
testint: 1
testfloat: 2.5
testbool: true`

	source := strings.NewReader(configYaml)

	config, err := Load(&grammar, source)
	if err != nil {
		t.Fatalf("error: loading config: %v", err)
	}

	// Test success scenarios
	t.Log("Successfully get string")
	requireEqual(t, "val", config.GetString("teststring"))
	t.Log("Successfully get int")
	requireEqual(t, 1, config.GetInt("testint"))
	t.Log("Successfully get float64")
	requireEqual(t, 2.5, config.GetFloat64("testfloat"))
	t.Log("Successfully get bool")
	requireEqual(t, true, config.GetBool("testbool"))

	// Test failure scenarios - expected to receive nil values
	t.Log("Get error trying to retrive non-existent key")
	requireEqual(t, nil, config.Get("nonexistent"))
	t.Log("Get error trying to convert get int as string")
	requireEqual(t, "", config.GetString("testint"))
	t.Log("Get error trying to convert get string as int")
	requireEqual(t, 0, config.GetInt("teststring"))
	t.Log("Get error trying to convert get string as float64")
	requireEqual(t, 0.0, config.GetFloat64("teststring"))
	t.Log("Get error trying to convert get string as bool")
	requireEqual(t, false, config.GetBool("teststring"))
}

func TestSet(t *testing.T) {
	grammar := struct {
		TestString string  `yaml:"test_string"`
		TestInt    int     `yaml:"test_int"`
		TestFloat  float64 `yaml:"test_float"`
		TestBool   bool    `yaml:"test_bool"`
	}{}

	configYaml := `test_string: "test_value"
test_int: 1
test_float: 1.5
test_bool: true`
	source := strings.NewReader(configYaml)

	config, err := Load(&grammar, source)
	if err != nil {
		t.Fatalf("error: loading config: %v", err)
	}

	tests := []struct {
		label    string
		key      string
		expected interface{}
		typ      reflect.Kind
		err      error
	}{
		{
			label:    "Setting string value",
			key:      "test_string",
			expected: "set",
			typ:      reflect.String,
		},
		{
			label:    "Setting int value",
			key:      "test_int",
			expected: 10,
			typ:      reflect.Int,
		},
		{
			label:    "Setting float value",
			key:      "test_float",
			expected: 4.5,
			typ:      reflect.Float64,
		},
		{
			label:    "Setting bool value",
			key:      "test_bool",
			expected: false,
			typ:      reflect.Bool,
		},
		{
			label: "Fail trying to set non-existent key",
			key:   "non_existent",
			typ:   reflect.Bool,
			err:   NewErrKeyNotFound("non_existent"),
		},
	}

	for i, tt := range tests {
		t.Logf("Assertion %d: %s", i+1, tt.label)

		err := config.Set(tt.key, tt.expected)
		if err != nil {
			if tt.err != nil {
				if errors.Is(tt.err, err) {
					t.Log("\tPassed")
					continue
				} else {
					t.Logf("\tExpected error %s, got %s", tt.err.Error(), err.Error())
				}
			}
			t.Errorf("error: %v", err)
			continue
		}

		switch tt.typ {
		case reflect.String:
			requireEqual(t, tt.expected.(string), config.Get(tt.key).(string))
		case reflect.Int:
			requireEqual(t, tt.expected.(int), config.Get(tt.key).(int))
		case reflect.Float64:
			requireEqual(t, tt.expected.(float64), config.Get(tt.key).(float64))
		case reflect.Bool:
			requireEqual(t, tt.expected.(bool), config.Get(tt.key).(bool))
		}
	}
}

func requireEqual(t *testing.T, expected interface{}, actual interface{}) {
	switch e := expected.(type) {
	case string:
		a := actual.(string)
		if e != a {
			// Error is log then fail, not very semantic naming
			t.Errorf("\tFailed: expected '%s', got '%s'", e, a)
			return
		}
	case int:
		a := actual.(int)
		if e != a {
			t.Errorf("\tFailed: expected %d, got %d", e, a)
			return
		}
	case float64:
		a := actual.(float64)
		if e != a {
			t.Errorf("\tFailed: expected %f, got %f", e, a)
			return
		}
	case bool:
		a := actual.(bool)
		if e != a {
			t.Errorf("\tFailed: expected %t, got %t", e, a)
			return
		}
	}

	t.Log("\tPassed")
}

func TestGetTypedValue(t *testing.T) {
	tests := []struct {
		label    string
		value    string
		kind     reflect.Kind
		expected interface{}
		err      error
	}{
		{
			label:    "String type - success",
			value:    "test",
			kind:     reflect.String,
			expected: "test",
		},
		{
			label:    "Int type - success",
			value:    "1",
			kind:     reflect.Int,
			expected: 1,
		},
		{
			label: "Int type - failure",
			value: "not_int",
			kind:  reflect.Int,
			err: &strconv.NumError{
				Func: "ParseInt",
				Num:  "not_int",
				Err:  strconv.ErrSyntax,
			},
		},
		{
			label:    "Float type - success",
			value:    "1.1",
			kind:     reflect.Float64,
			expected: 1.1,
		},
		{
			label: "Float type - failure",
			value: "not_float",
			kind:  reflect.Float64,
			err: &strconv.NumError{
				Func: "ParseFloat",
				Num:  "not_float",
				Err:  strconv.ErrSyntax,
			},
		},
		{
			label:    "Bool type - success",
			value:    "true",
			kind:     reflect.Bool,
			expected: true,
		},
		{
			label: "Bool type - failure",
			value: "not_bool",
			kind:  reflect.Bool,
			err: &strconv.NumError{
				Func: "ParseBool",
				Num:  "not_bool",
				Err:  strconv.ErrSyntax,
			},
		},
		{
			label: "Invalid type",
			value: "anything",
			kind:  reflect.Complex64,
			err:   fmt.Errorf("no conversion available for this kind: %s", reflect.Complex64),
		},
	}

	t.Log("Given the need to extract a typed value from a string")
	for i, tt := range tests {
		fn := func(t *testing.T) {
			t.Logf("TEST %d: %s", i+1, tt.label)
			tval, err := getTypedValue(tt.value, tt.kind)
			if err != nil {
				if tt.err != nil {
					if tt.err.Error() == err.Error() {
						t.Log("\tPassed")
						return
					} else {
						t.Fatalf("\tFailed: Expected error %s, got %s", tt.err.Error(), err.Error())
					}
				}
			}

			if tval != tt.expected {
				t.Fatalf("\tFailed: Expected %v, got %v", tt.expected, tval)
			}
			t.Log("\tPassed")
		}
		t.Run(tt.label, fn)
	}
}
