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

package container

type element struct {
	key   interface{}
	value interface{}
}

// OrderedMap is a map that
// which elements are ordered
type OrderedMap struct {
	elements []element
	hashMap  map[interface{}]int
}

// NewOrderedMap creates a new OrderedMap
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{[]element{}, map[interface{}]int{}}
}

// Set sets a keyed value
func (Map *OrderedMap) Set(key, value interface{}) *OrderedMap {
	if index, ok := Map.hashMap[key]; ok {
		Map.elements[index] = element{key, value}
	} else {
		Map.hashMap[key] = len(Map.elements)
		Map.elements = append(Map.elements, element{key, value})
	}
	return Map
}

// Delete deletes a value by key, return false
// if the value was not found
func (Map *OrderedMap) Delete(key interface{}) bool {
	if index, ok := Map.hashMap[key]; ok {
		Map.elements = append(Map.elements[0:index], Map.elements[index+1:len(Map.elements)]...)
		delete(Map.hashMap, key)
		return true
	}
	return false
}

// Get returns a value by key
func (Map OrderedMap) Get(key interface{}) interface{} {
	for _, element := range Map.elements {
		if element.key == key {
			return element.value
		}
	}
	return nil
}

// Values returns all values in the map
func (Map OrderedMap) Values() []interface{} {
	values := []interface{}{}
	for _, element := range Map.elements {
		values = append(values, element.value)
	}
	return values
}

// Keys returns all keys in a map
func (Map OrderedMap) Keys() []interface{} {
	keys := []interface{}{}
	for _, element := range Map.elements {
		keys = append(keys, element.key)
	}
	return keys
}

// Has returns true if a key was found in the map
func (Map OrderedMap) Has(key interface{}) bool {
	for _, element := range Map.elements {
		if element.key == key {
			return true
		}
	}
	return false
}

// Length returns the length of a map
func (Map OrderedMap) Length() int {
	return len(Map.elements)
}

// KeyAt returns a key at index
func (Map OrderedMap) KeyAt(index int) (key interface{}) {
	if index >= Map.Length() || index < 0 {
		return nil
	}
	return Map.elements[index].key
}

// ValueAt returns a value at index, or nil
func (Map OrderedMap) ValueAt(index int) (key interface{}) {
	if index >= Map.Length() || index < 0 {
		return nil
	}
	return Map.elements[index].value
}
