package util

import (
	"errors"
	"reflect"
	"strconv"
)

func StrArrayFromInterface(i interface{}) ([]string, error) {
	object := reflect.ValueOf(i)
	items := make([]string, 0)

	for i := 0; i < object.Len(); i++ {
		val, ok := object.Index(i).Interface().(string)
		if !ok {
			return items, errors.New("error getting item at index " + strconv.FormatInt(int64(i), 10))
		}
		items = append(items, val)
	}

	return items, nil
}
func InterfaceArrayFromInterface(i interface{}) ([]interface{}, error) {
	object := reflect.ValueOf(i)
	items := make([]interface{}, 0)

	for i := 0; i < object.Len(); i++ {
		items = append(items, object.Index(i).Interface())
	}

	return items, nil
}