# Goconf

Lightweight configuration package

## Usage

Creating a config is pretty straightforward, it requires calling the `Load` function with a pointer to a struct value
and a reader that can be used to read a YAML configuration. A short and simple example:
```go
	grammar := struct {
		Test      string
		CustomKey string `yaml:"custom_key"`
        EnvOverwrite string `env:"OVERWRITE" yaml:"env_overwrite"`
	}{}

	configYaml := `test: "test_value"
custom_key: "test"
env_overwrite: "don't look"`
	source := strings.NewReader(configYaml)

	os.Setenv("CFG_OVERWRITE", "look")

	cfg, err := Load(&grammar, source)
	if err != nil {
		t.Fatalf("error: loading config: %v", err)
    }
    
    // Output: test_value
    fmt.Println(cfg.GetString("test"))

    // Output: test
    fmt.Println(cfg.GetString("custom_key"))

    // Output: look
    fmt.Println(cfg.GetString("env_overwrite"))
```

### Getting and setting values

Using the appropriate `Configuration.Get(key string) interface{}` and `Configuration.Set(key string, value interface{}) error` methods.

To get values with an expected type, there are some helpers (currently only these types are supported):  
`GetString(key string) string`  
`GetInt(key string) int`  
`GetFloat64(key string) float64`  
`GetBool(key string) bool`  

(`TODO: Set is not typesafe, and therefor almost useless currently.`)

## Tags

#### `yaml:"custom_key"`

This is a built-in for the `gopkg.in/yaml.v2` package but goconf uses it for custom keys as well.

#### `env:"key"`

This tag tells the parser to look for the value in the supplied environment variable. The key will be turned uppercase
and all non alphanumeric characters will be replaced to `_`.

#### `default:"value"`

This tag sets the value to the given one if the actual value after loading the YAML and setting from ENV is still the zero-value
of the field's type.