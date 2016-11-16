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
// is ordered unlike Go's map[T]T
type OrderedMap struct {
	elements []element
	hashMap  map[interface{}]int
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{[]element{}, map[interface{}]int{}}
}

func (Map *OrderedMap) Set(key, value interface{}) *OrderedMap {
	if index, ok := Map.hashMap[key]; ok {
		Map.elements[index] = element{key, value}
	} else {
		Map.hashMap[key] = len(Map.elements)
		Map.elements = append(Map.elements, element{key, value})
	}
	return Map
}

func (Map *OrderedMap) Delete(key interface{}) bool {
	if index, ok := Map.hashMap[key]; ok {
		Map.elements = append(Map.elements[0:index], Map.elements[index+1:len(Map.elements)]...)
		delete(Map.hashMap, key)
		return true
	}
	return false
}

func (Map OrderedMap) Get(key interface{}) interface{} {
	for _, element := range Map.elements {
		if element.key == key {
			return element.value
		}
	}
	return nil
}

func (Map OrderedMap) Values() []interface{} {
	values := []interface{}{}
	for _, element := range Map.elements {
		values = append(values, element.value)
	}
	return values
}

func (Map OrderedMap) Keys() []interface{} {
	keys := []interface{}{}
	for _, element := range Map.elements {
		keys = append(keys, element.key)
	}
	return keys
}

func (Map OrderedMap) Has(key interface{}) bool {
	for _, element := range Map.elements {
		if element.key == key {
			return true
		}
	}
	return false
}

func (Map OrderedMap) Length() int {
	return len(Map.elements)
}

func (Map OrderedMap) KeyAt(index int) (key interface{}) {
	if index >= Map.Length() || index < 0 {
		return nil
	}
	return Map.elements[index].key
}
func (Map OrderedMap) ValueAt(index int) (key interface{}) {
	if index >= Map.Length() || index < 0 {
		return nil
	}
	return Map.elements[index].value
}
