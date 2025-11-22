package mongo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/notifyx/core"
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
