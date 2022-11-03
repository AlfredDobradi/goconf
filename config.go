package goconf

import (
	"fmt"
	"io"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Characters used when parsing the struct tags
// Currently no way of changing nor I see the point in implementing
const (
	AssignChar    string = `:`
	QuoteChar     string = `"`
	SeparatorChar string = ` `
)

// Configuration is the object that holds the root node
type Configuration struct {
	Node *Node
}

// Load creates a Configuration object from the supplied grammar structure and reader
// The loader only supports YAML so the reader should hold a decodeable YAML structure
func Load(grammar interface{}, source io.Reader) (*Configuration, error) {
	app := &Configuration{}

	decoder := yaml.NewDecoder(source)
	err := decoder.Decode(grammar)
	if err != nil {
		return app, err
	}

	v := reflect.ValueOf(grammar)
	iv := reflect.Indirect(v)
	if v.Kind() != reflect.Ptr || iv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a pointer to a struct but got %T", grammar)
	}

	rootNode := Node{
		Name:     "Application",
		Children: make([]*Node, 0),
	}

	buildNode(&rootNode, iv)

	app.Node = &rootNode

	return app, nil
}

// Set changes the value of a node.
// TODO: This is currently NOT typesafe and therefor not recommended to use.
func (c *Configuration) Set(key string, value interface{}) error {
	context := c.Node.FindNode(key)
	if context == nil {
		return NewErrKeyNotFound(key)
	}
	context.Value = value
	return nil
}

// Get retrieves the value of the node associated with the key
// Nested keys can be formatted using . (e.g. http.server.host)
// Since this returns an interface, using it in a typed context requires assertion
// TODO: Helper functions for typed retrieval (GetString, GetInt, etc.)
func (c *Configuration) Get(key string) interface{} {
	context := c.Node.FindNode(key)
	if context == nil {
		return nil
	}
	return context.Value
}

// GetString is a convenience function for retrieving a string value
func (c *Configuration) GetString(key string) string {
	val := c.Get(key)
	if tval, ok := val.(string); ok {
		return tval
	}
	return ""
}

// GetInt is a convenience function for retrieving an int value
func (c *Configuration) GetInt(key string) int {
	val := c.Get(key)
	if tval, ok := val.(int); ok {
		return tval
	}
	return 0
}

// GetFloat64 is a convenience function for retrieving a float64 value
func (c *Configuration) GetFloat64(key string) float64 {
	val := c.Get(key)
	if tval, ok := val.(float64); ok {
		return tval
	}
	return 0
}

// GetBool is a convenience function for retrieving a bool value
func (c *Configuration) GetBool(key string) bool {
	val := c.Get(key)
	if tval, ok := val.(bool); ok {
		return tval
	}
	return false
}

// getTypedValue converts string values (from env and default tags) to typed values
func getTypedValue(source string, kind reflect.Kind) (interface{}, error) {
	switch kind {
	case reflect.String:
		return source, nil
	case reflect.Int:
		if intVal, err := strconv.ParseInt(source, 10, 0); err != nil {
			return nil, err
		} else {
			return int(intVal), nil
		}
	case reflect.Float64:
		if floatVal, err := strconv.ParseFloat(source, 64); err != nil {
			return nil, err
		} else {
			return floatVal, nil
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(source); err != nil {
			return nil, err
		} else {
			return boolVal, nil
		}
	default:
		return nil, fmt.Errorf("no conversion available for this kind: %s", kind)
	}
}
