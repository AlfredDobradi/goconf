package goconf

import (
	"os"
	"reflect"
	"strings"
)

// Node represents a single node in the Configuration
type Node struct {
	// Name is the name of the original field retrieved via reflection
	Name string

	// Key is the string to use in order to retrieve the node
	Key string

	// Parent holds a pointer to the parent node
	Parent *Node

	// Children holds pointers to the Child nodes for traversal
	Children []*Node

	// Typ stores original field's type
	Typ reflect.Type

	// Value stores the value if it's a leaf node (has no children)
	Value interface{}

	// Tag stores a pointer to the parsed Tag
	Tag *Tag
}

// GetChild is a helper function to get a node's child
func (n *Node) GetChild(key string) *Node {
	for _, child := range n.Children {
		if child.Key == key {
			return child
		}
	}

	return nil
}

// FindNode traverses the tree to find the node with the given key
// Nested keys use . to traverse deeper
func (n *Node) FindNode(path string) *Node {
	pathSegments := strings.Split(path, ".")
	context := n

	for i, segment := range pathSegments {
		child := context.GetChild(segment)
		if child == nil {
			return nil
		}

		context = child

		if i == len(pathSegments)-1 {
			return context
		}
	}
	return nil
}

// setValue takes the configuration value and checks if the field can be resolved from ENV
// If the value is empty after trying to resolve from ENV, it tries to set the default value
func (n *Node) setValue(ivalue interface{}) error {
	var err error

	// At this point, ivalue is the value found in the config
	// See if env tag is set and if the environment variable is set
	if n.Tag.Env != "" && os.Getenv(n.Tag.Env) != "" {
		ivalue, err = getTypedValue(os.Getenv(n.Tag.Env), n.Typ.Kind())
		if err != nil {
			// TODO: return better errors
			return err
		}
	}

	var needDefault = ivalue == nil
	if !needDefault {
		switch n.Typ.Kind() {
		case reflect.String:
			needDefault = ivalue.(string) == ""
		case reflect.Int:
			needDefault = ivalue.(int) == 0
		case reflect.Float64:
			needDefault = ivalue.(float64) == 0.0
		case reflect.Bool:
			needDefault = !ivalue.(bool)
		}
	}

	// If the value is still nil or an empty string, set it to default
	if needDefault && n.Tag.Default != "" {
		ivalue, err = getTypedValue(n.Tag.Default, n.Typ.Kind())
		if err != nil {
			// TODO: return better errors
			return err
		}
	}

	n.Value = ivalue
	return nil
}

// buildNode uses reflection to its possible children to the given parent node
func buildNode(parent *Node, iv reflect.Value) {
	for i := 0; i < iv.NumField(); i++ {
		node := &Node{}
		ft := iv.Type().Field(i)
		fv := iv.Field(i)

		// Set data useful for both further parents and leaf nodes
		node.Name = ft.Name
		node.Tag = parseTag(ft)
		node.Typ = ft.Type

		// If it's a struct node, we should build its children as well
		if ft.Type.Kind() == reflect.Struct {
			buildNode(node, fv)
		} else {
			// Set the node value using the original value, env or default
			_ = node.setValue(iv.FieldByName(ft.Name).Interface())
		}

		// if we used a custom key in YAML, we should use that in the node
		if node.Tag.Key != "" {
			node.Key = node.Tag.Key
		} else {
			node.Key = strings.ToLower(strings.Replace(ft.Name, "-", "_", -1))
		}

		// Add the parent in the node and the node as children in the parent
		node.Parent = parent
		parent.Children = append(parent.Children, node)
	}
}
