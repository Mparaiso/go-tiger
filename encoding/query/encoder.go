//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package query

import (
	"fmt"
	"net/url"
	"reflect"
)

var (
	// ErrNotAStruct is returned when an variable is not a struct
	ErrNotAStruct = fmt.Errorf("Error not a struct.")
)

// ToValues turns a struct into url.Values
// than can be then encoded into a safe query string
// it only supports structs or pointers to struct
func ToValues(target interface{}) (url.Values, error) {
	Value := reflect.Indirect(reflect.ValueOf(target))
	Type := Value.Type()
	// if not struct ,return error
	if Type.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}
	values := url.Values{}
	// for each field in struct
	for i := 0; i < Type.NumField(); i++ {
		field := Type.Field(i)
		key := field.Name
		// if it has a schema struct tag , use it as key
		if tagValue, ok := field.Tag.Lookup("schema"); ok {
			if tagValue == "-" {
				continue
			}
			key = tagValue
		}
		fieldValue := reflect.Indirect(Value.FieldByName(field.Name))
		// if the filed is a slice or an array
		if fieldKind := fieldValue.Kind(); fieldKind == reflect.Slice || fieldKind == reflect.Array {
			value := Value.FieldByName(field.Name)
			for i := 0; i < value.Len(); i++ {
				// put each element of the array in the map , keyed by the same key
				values.Add(key, fmt.Sprint(value.Index(i).Interface()))
			}
		} else {
			// or put the value of the field in the map directly
			values.Set(key, fmt.Sprint(Value.FieldByName(field.Name).Interface()))

		}
	}
	return values, nil
}
