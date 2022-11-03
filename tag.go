package goconf

import (
	"reflect"
	"regexp"
	"strings"
)

// tagStore is a convenience type for searching elements
type tagStore map[string]string

// Tag represents a parsed struct tag
type Tag struct {
	// Default is the value we should use if we still have a zero-value after parsing
	Default string

	// Key is the custom key if it was used for decoding YAML
	Key string

	// Required tells us whether we should stop if the field is not set
	// TODO: this is not yet used at all
	Required bool

	// Env tells the parser what variable it should look for populating the node
	Env string

	// store is the underlying map that was used to fill the fields
	store tagStore
}

// Has is a convenience method to check whether the key exists in the store
func (ts tagStore) Has(key string) bool {
	if _, ok := ts[key]; ok {
		return true
	}
	return false
}

// Has is a convenience method to get a tag value from the store
func (ts tagStore) Get(key string) string {
	if val, ok := ts[key]; ok {
		return val
	}
	return ""
}

// parseTag parses a struct tag into a Tag value
func parseTag(ft reflect.StructField) *Tag {
	tag := &Tag{
		store: make(tagStore),
	}

	// The parsing is not very robust, we rely on having a relatively standard tag
	// There's an option to override the symbols but it makes little sense to do so
	// as it can break yaml decoding or the parser
	elements := strings.Split(string(ft.Tag), SeparatorChar)
	for _, element := range elements {
		if len(element) == 0 {
			continue
		}
		assignPos := strings.Index(element, AssignChar)
		if assignPos == -1 {
			tag.store[element] = ""
			continue
		}

		key := element[:assignPos]
		value := strings.Trim(element[assignPos+1:], QuoteChar)
		tag.store[key] = value
	}

	tag.Required = tag.store.Has("required")
	tag.Default = tag.store.Get("default")
	tag.Key = tag.store.Get("yaml")

	env := tag.store.Get("env")
	if env != "" {
		tag.Env = normalizeEnvKey(env)
	}

	return tag
}

func normalizeEnvKey(env string) string {
	env = strings.ToUpper(env)

	// Replace special characters with underscores
	reg, err := regexp.Compile("[^A-Z0-9]+")
	if err != nil {
		return ""
	}
	env = reg.ReplaceAllString(env, "_")

	// Replace repeating underscores with one
	reg, err = regexp.Compile("_+")
	if err != nil {
		return ""
	}
	env = reg.ReplaceAllString(env, "_")

	// Trim the underscores from either side
	env = strings.Trim(env, "_")

	return env
}
