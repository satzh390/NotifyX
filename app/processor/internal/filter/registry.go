package filter

import (
	"log/slog"
	"reflect"
	"strings"
)

// RegisterCustomFilter registers a custom filter instance.
// The filter type name is derived from the struct type name.
// This is called automatically by custom filter implementations via init().
//
// Example:
//
//	func init() {
//	    filter.RegisterCustomFilter(&MyCustomFilter{})
//	}
func RegisterCustomFilter(filterInstance CustomFilter) {
	if filterInstance == nil {
		return
	}

	// Get the type name from the struct
	typeName := getTypeName(filterInstance)
	if typeName == "" {
		slog.Warn("filter: failed to get type name for custom filter, skipping registration")
		return
	}

	// Register the filter
	DefaultFilterRegistry.Register(typeName, filterInstance)
	slog.Info("filter: registered custom filter", slog.String("type", typeName))
}

// getTypeName extracts the type name from a filter instance.
// For a struct like ExampleCustomFilter, it returns "ExampleCustomFilter".
func getTypeName(filterInstance CustomFilter) string {
	if filterInstance == nil {
		return ""
	}

	// Get the type using reflection
	t := reflect.TypeOf(filterInstance)
	if t == nil {
		return ""
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Get the type name
	typeName := t.Name()
	if typeName == "" {
		return ""
	}

	// Remove any package prefix if present (shouldn't happen, but be safe)
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		typeName = typeName[idx+1:]
	}

	return typeName
}
