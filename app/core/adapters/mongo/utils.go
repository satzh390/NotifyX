package mongo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/notifyx/core"
	"github.com/notifyx/core/domain"
	"go.mongodb.org/mongo-driver/bson"
)

func BuildUpdateMap[T any](doc T) (bson.M, error) {
	v := reflect.ValueOf(doc)
	t := reflect.TypeOf(doc)

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("BuildUpdateMap: expected struct but got %s", t.Kind())
	}

	result := bson.M{}

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		// Skip system fields
		if fieldType.Tag.Get("immutable") == "true" {
			continue
		}

		// Extract bson tag
		bsonTag := fieldType.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		// Extract the key (ignore other modifiers)
		key := strings.Split(bsonTag, ",")[0]
		// Omit zero values if omitempty is present
		if strings.Contains(bsonTag, "omitempty") && core.IsZero(fieldValue.Interface()) {
			continue
		}

		result[key] = fieldValue.Interface()
	}

	return result, nil
}

func BuildBsonFilter(opts domain.ListOptions) bson.M {
	filter := bson.M{}
	for key, value := range opts.Filter {
		if value != "" {
			filter[key] = value
		}
	}

	return filter
}

func BuildBsonSort(opts domain.ListOptions, defaultSort map[string]int) bson.D {
	sortMap := make(map[string]int)
	for k, v := range defaultSort {
		sortMap[k] = v
	}

	for _, s := range opts.SortBy {
		order := 1
		if s.Order == domain.SortDesc {
			order = -1
		}
		sortMap[s.Field] = order
	}

	// Build bson.D in correct order:
	// 1. User-specified fields (in order provided)
	finalSort := bson.D{}
	used := map[string]bool{}
	for _, s := range opts.SortBy {
		finalSort = append(finalSort, bson.E{
			Key:   s.Field,
			Value: sortMap[s.Field],
		})
		used[s.Field] = true
	}

	// 2. Default fields not specified by user
	for k, v := range defaultSort {
		if !used[k] {
			finalSort = append(finalSort, bson.E{Key: k, Value: v})
		}
	}

	return finalSort
}

func PageOrDefaultParam(opts domain.ListOptions) (page int, pageSize int) {
	// Set default pagination if not provided (0-based)
	page = max(0, opts.Pagination.Page)
	pageSize = opts.Pagination.PageSize
	if pageSize < 1 {
		pageSize = 20 // default page size
	}

	pageSize = min(100, pageSize)
	return
}

func ParseCommaSeparatedString(raw string) []string {
	rawValues := strings.Split(raw, ",")
	values := make([]string, 0, len(rawValues))
	for _, item := range rawValues {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
