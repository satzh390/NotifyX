package core

import "reflect"

func IsZero(v interface{}) bool {
	return reflect.ValueOf(v).IsZero()
}
