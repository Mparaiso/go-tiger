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

// Package funcs provides utilities to enable functional programming with Go.
// The main goal here is to provide type safety at runtime while being able to
// use generic algorithms and patterns. Performance, while important, is not a priority.
// Code reuse is. A reduce function shouldn't be written for every possible type combination,
// that's why a function is generated, at runtime, with the proper type signature so its use
// is completely type safe.
package funcs

import (
	"errors"
	"reflect"
)

var (
	// ErrNotAPointer : Error the value is not a pointer
	ErrNotAPointer = errors.New("ErrNotAPointer : Error the value is not a pointer")
	// ErrNotAMap : Error the value is not a map
	ErrNotAMap = errors.New("ErrNotAMap : Error the value is not a map")
	// ErrNotAFunction : the value is not a function
	ErrNotAFunction = errors.New("ErrNotAFunction : the value is not a function")
	// ErrNotEnoughArguments : Error the signature of the function doesn't have enough arguments to be set
	ErrNotEnoughArguments = errors.New("ErrNotEnoughArguments : Error the signature of the function doesn't have enough arguments to be set")
	// ErrReduceIncompatibleSignature : Error the signature of the function is not compatible with the operation
	ErrReduceIncompatibleSignature = errors.New("ErrReduceIncompatibleSignature : Error the signature of the function is not compatible with the operation")
	// ErrNotIterable : Error the value is not a slice or an array
	ErrNotIterable = errors.New("ErrNotIterable : Error the value is not a slice or an array")
	// ErrInvalidNumberOfReturnValues : Error the number of return values in the function is not valid
	ErrInvalidNumberOfReturnValues = errors.New("ErrInvalidNumberOfReturnValues : Error the number of return values in the function is not valid")
	// ErrInvalidNumberOfInputValues : Error the number of arguments in the function is not valid
	ErrInvalidNumberOfInputValues = errors.New("ErrInvalidNumberOfInputValues : Error the number of arguments in the function is not valid")
	// ErrUnexpectedType : Error a type was expected and a different type was foun
	ErrUnexpectedType = errors.New("ErrUnexpectedType : Error a type was expected and a different type was found")
	// ErrNoComparableType : Error a type was expected to be comparable
	ErrNoComparableType = errors.New("ErrNoComparableType : Error a type was expected to be comparable")
)

// Must panics on error
// it returns an error conveniently so it can be used in a
// declaration statement outside a body
// example:
//
//      var _ := funcs.Must(ShouldNotReturnAnError())
func Must(err error) error {
	if err != nil {
		panic(err)
	}
	return err
}

// MakeForEach creates a forEach function from a pointer to function with the following signatures :
//
//		forEach(collection []A, callback func(A) )
//		forEach(collection []A, callback func(A,int))
//		forEach(collection []A, callback func(A,int,[]A))
//
// or return an error if types do not match.
// forEach calls callback for every element of collection.
func MakeForEach(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Function := Value.Elem()
	// expect a pointer to function
	if Function.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FunctionType := Function.Type()
	// expect a function with 2 arguments
	if FunctionType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect a function with a single return value
	if FunctionType.NumOut() != 0 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FunctionType.In(0)
	CallbackType := FunctionType.In(1)
	if CallbackType.NumOut() != 0 {
		return ErrInvalidNumberOfReturnValues
	}
	if numIn := CallbackType.NumIn(); numIn < 1 || numIn > 3 {
		return ErrInvalidNumberOfInputValues
	}
	if CollectionType.Elem() != CallbackType.In(0) {
		return ErrUnexpectedType
	}
	if CallbackType.NumIn() > 1 {
		if CallbackType.In(1) != reflect.TypeOf(1) {
			return ErrUnexpectedType
		}
	}
	if CallbackType.NumIn() > 2 {
		if CallbackType.In(2) != CollectionType {
			return ErrUnexpectedType
		}
	}
	forEachFunction := reflect.MakeFunc(FunctionType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		mapper := args[1]
		for i := 0; i < collection.Len(); i++ {
			switch mapper.Type().NumIn() {
			case 1:
				mapper.Call([]reflect.Value{collection.Index(i)})
			case 2:
				mapper.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})
			case 3:
				mapper.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})
			}
		}
		return
	})
	Value.Elem().Set(forEachFunction)
	return nil

}

// MakeGroupBy creates a  groupBy function froma a pointer function with the following signatures :
//
//		groupBy(collection []A, selector func(element A)(key B))map[B][]A
//		groupBy(collection []A, selector func(element A,index int)(key B))map[B][]A
//		groupBy(collection []A, selector func(element A,index int,collection []A)(key B))map[B][]A
//
// or returns an error if types do not match.
//
// groupBy returns a map of slices so that every element from collection is grouped
// by the result of the selector function, which serves as the key for the map.
func MakeGroupBy(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	FuncValue := Value.Elem()
	if FuncValue.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FuncType := FuncValue.Type()
	if FuncType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	if FuncType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FuncType.In(0)
	SelectorType := FuncType.In(1)
	if kind := CollectionType.Kind(); kind != reflect.Array && kind != reflect.Slice {
		return ErrUnexpectedType
	}
	if SelectorType.Kind() != reflect.Func {
		return ErrUnexpectedType
	}
	if SelectorType.NumIn() < 1 {
		return ErrInvalidNumberOfInputValues
	}
	if SelectorType.NumOut() > 3 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionElemType := CollectionType.Elem()
	if SelectorType.In(0) != CollectionElemType {
		return ErrUnexpectedType
	}
	if SelectorType.NumIn() > 1 && SelectorType.In(1) != reflect.TypeOf(1) {
		return ErrUnexpectedType
	}
	if SelectorType.NumIn() > 2 && SelectorType.In(2) != CollectionType {
		return ErrUnexpectedType
	}
	if FuncType.Out(0).Kind() != reflect.Map {
		return ErrUnexpectedType
	}
	if FuncType.Out(0).Key() != SelectorType.Out(0) {
		return ErrUnexpectedType
	}
	if FuncType.Out(0).Elem() != CollectionType {
		return ErrUnexpectedType
	}
	groupByFunc := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		collectionType := collection.Type()
		selector := args[1]
		Map := reflect.MakeMap(reflect.MapOf(selector.Type().Out(0), collection.Type()))
		results = []reflect.Value{Map}
		numIn := selector.Type().NumIn()
		for i := 0; i < collection.Len(); i++ {
			var keys []reflect.Value
			switch numIn {
			case 1:
				keys = selector.Call([]reflect.Value{collection.Index(i)})
			case 2:
				keys = selector.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})
			case 3:
				keys = selector.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})
			}
			if len(Map.MapKeys()) == 0 || !func() bool {
				for _, key := range Map.MapKeys() {
					if keys[0].Interface() == key.Interface() {
						return true
					}
				}
				return false
			}() {
				Map.SetMapIndex(keys[0], reflect.New(collectionType).Elem())
			}
			Map.SetMapIndex(keys[0], reflect.Append(Map.MapIndex(keys[0]), collection.Index(i)))
		}
		return
	})
	Value.Elem().Set(groupByFunc)
	return nil
}

// MakeGetValues assigns a function to pointerToFunction with the following signature :
//
//		getValues(map[K]V)[]V
//
// or returns an error if types do not match. getValues extract the values of a map into an array
func MakeGetValues(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	FuncValue := Value.Elem()
	if FuncValue.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FuncType := FuncValue.Type()
	if FuncType.NumIn() != 1 {
		return ErrInvalidNumberOfInputValues
	}
	if FuncType.In(0).Kind() != reflect.Map {
		return ErrNotAMap
	}
	if FuncType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	MapType := FuncType.In(0)

	if FuncType.Out(0).Kind() != reflect.Slice && FuncType.Out(0).Kind() != reflect.Array {
		return ErrUnexpectedType
	}

	if FuncType.Out(0).Elem() != MapType.Elem() {
		return ErrUnexpectedType
	}
	getValues := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		Map := args[0]
		Out := reflect.MakeSlice(reflect.SliceOf(Map.Type().Elem()), 0, 0)
		for _, key := range Map.MapKeys() {
			Out = reflect.Append(Out, Map.MapIndex(key))
		}
		results = append(results, Out)
		return
	})
	Value.Elem().Set(getValues)
	return nil
}

// MakeGetKeys assigns a getKey to the pointerToFunction with the following signature :
//
//		getKeys(map[K]V)[]K
//
// or returns an error if types do not match. getKeys extract the keys of a map into an array
func MakeGetKeys(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	FuncValue := Value.Elem()

	if FuncValue.Kind() != reflect.Func {
		return ErrNotAMap
	}
	FuncType := FuncValue.Type()

	if FuncType.NumIn() != 1 {
		return ErrInvalidNumberOfInputValues
	}

	if FuncType.In(0).Kind() != reflect.Map {
		return ErrNotAMap
	}

	if FuncType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}

	if FuncType.Out(0).Kind() != reflect.Slice && FuncType.Out(0).Kind() != reflect.Array {
		return ErrUnexpectedType
	}

	MapType := FuncType.In(0)

	if FuncType.Out(0).Elem() != MapType.Key() {
		return ErrUnexpectedType
	}
	getKeys := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		Map := args[0]
		Out := reflect.MakeSlice(reflect.SliceOf(Map.Type().Key()), 0, 0)
		for _, key := range Map.MapKeys() {
			Out = reflect.Append(Out, key)
		}
		results = append(results, Out)
		return
	})
	Value.Elem().Set(getKeys)
	return nil
}

// MakeKeyBy creates a  keyBy function from a pointer function with the following possible signatures :
//
//		keyBy(collection []A, selector func(element A)B)map[B]A
//		keyBy(collection []A, selector func(element A,index int)B)map[B]A
//		keyBy(collection []A, selector func(element A,index int,collection []A)B)map[B]A
//
// or returns an error if types do not match.
//
// keyBy returns a map so that elements from collection are keyed by the return value of
// the selector function.
func MakeKeyBy(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	FuncValue := Value.Elem()
	if FuncValue.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FuncType := FuncValue.Type()
	if FuncType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	if FuncType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FuncType.In(0)
	SelectorType := FuncType.In(1)
	if kind := CollectionType.Kind(); kind != reflect.Array && kind != reflect.Slice {
		return ErrUnexpectedType
	}
	if SelectorType.Kind() != reflect.Func {
		return ErrUnexpectedType
	}
	if SelectorType.NumIn() < 1 {
		return ErrInvalidNumberOfInputValues
	}
	if SelectorType.NumOut() > 3 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionElemType := CollectionType.Elem()
	if SelectorType.In(0) != CollectionElemType {
		return ErrUnexpectedType
	}
	if SelectorType.NumIn() > 1 && SelectorType.In(1) != reflect.TypeOf(1) {
		return ErrUnexpectedType
	}
	if SelectorType.NumIn() > 2 && SelectorType.In(2) != CollectionType {
		return ErrUnexpectedType
	}
	if FuncType.Out(0).Kind() != reflect.Map {
		return ErrUnexpectedType
	}
	if FuncType.Out(0).Key() != SelectorType.Out(0) {
		return ErrUnexpectedType
	}
	if FuncType.Out(0).Elem() != CollectionElemType {
		return ErrUnexpectedType
	}
	keyByFunc := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		selector := args[1]
		Map := reflect.MakeMap(reflect.MapOf(selector.Type().Out(0), collection.Type().Elem()))
		results = []reflect.Value{Map}
		numIn := selector.Type().NumIn()
		for i := 0; i < collection.Len(); i++ {
			var keys []reflect.Value
			switch numIn {
			case 1:
				keys = selector.Call([]reflect.Value{collection.Index(i)})
			case 2:
				keys = selector.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})
			case 3:
				keys = selector.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})
			}
			Map.SetMapIndex(keys[0], collection.Index(i))
		}
		return
	})
	Value.Elem().Set(keyByFunc)
	return nil
}

// MakeReduce creates a reduce function from a pointer function with the following signatures :
//
// 		reduce(collection []A, reducer func(result B,element A)B , initial B )B
//		reduce(collection []A, reducer func(result B,element A,index int)B , initial B)B
//		reduce(collection []A, reducer func(result B,element A,index int, collection []A) , initial B)B
//
// or returns an error if types do not match
//
// MakeReduce allow developers to quickly create reduce functions
// without starting from scratch each time they need basic functional
// programming capabilities. Result also allows type safety.
func MakeReduce(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Function := Value.Elem()
	// expect a pointer to function
	if Function.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FunctionType := Function.Type()
	// expect a function with 3 arguments
	if FunctionType.NumIn() != 3 {
		return ErrInvalidNumberOfInputValues
	}
	// expect a function with a single return value
	if FunctionType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FunctionType.In(0)
	ReducerType := FunctionType.In(1)
	InitialType := FunctionType.In(2)
	// expect initial type to match the function return value's type
	if FunctionType.Out(0) != InitialType {
		return ErrUnexpectedType
	}
	// expect CollectionType to be a collection
	if kind := CollectionType.Kind(); kind != reflect.Slice && kind != reflect.Array {
		return ErrNotIterable
	}
	// expect the reducer to be a function
	if ReducerType.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	// expect the reducer to have 1 return value
	if ReducerType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	// expect the return type of the reducer to equal the type of the initial value
	if ReducerType.Out(0) != InitialType {
		return ErrUnexpectedType
	}
	// expect the reducer to take 2 to 4 arguments
	if ReducerType.NumIn() < 2 || ReducerType.NumIn() > 4 {
		return ErrInvalidNumberOfInputValues
	}
	// expect the first argument of the reducer to match the type of initial value
	if ReducerType.In(0) != InitialType {
		return ErrUnexpectedType
	}
	// expect the second argument to match the type of element of the collection
	if ReducerType.In(1) != CollectionType.Elem() {
		return ErrUnexpectedType
	}
	// if more than 2 arguments
	if ReducerType.NumIn() > 2 {
		// the third argument should be a hint
		if ReducerType.In(2) != reflect.TypeOf(1) {
			return ErrUnexpectedType
		}
	}
	// if more than 3 arguments
	if ReducerType.NumIn() > 3 {
		// the fourth argument should be match the collection type
		if ReducerType.In(3) != CollectionType {
			return ErrUnexpectedType
		}
	}
	reduceFunction := reflect.MakeFunc(FunctionType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		reducer := args[1]
		initial := args[2]
		results = []reflect.Value{initial}
		for i := 0; i < collection.Len(); i++ {
			switch reducer.Type().NumIn() {
			case 2:
				results = reducer.Call([]reflect.Value{results[0], collection.Index(i)})
			case 3:
				results = reducer.Call([]reflect.Value{results[0], collection.Index(i), reflect.ValueOf(i)})
			case 4:
				results = reducer.Call([]reflect.Value{results[0], collection.Index(i), reflect.ValueOf(i), collection})
			}
		}
		return
	})
	Value.Elem().Set(reduceFunction)
	return nil

}

// MakeMap creates a map function from a pointer to function with the following signatures :
//
//		map(collection []A, mapper func(A)B )[]B
//		map(collection []A, mapper func(A,int)B )[]B
//		map(collection []A, mapper func(A,int,[]A)B )[]B
//
// or return an error if types do not match.
func MakeMap(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Function := Value.Elem()
	// expect a pointer to function
	if Function.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FunctionType := Function.Type()
	// expect a function with 2 arguments
	if FunctionType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect a function with a single return value
	if FunctionType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FunctionType.In(0)
	MapperType := FunctionType.In(1)
	if MapperType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	if numIn := MapperType.NumIn(); numIn < 1 || numIn > 3 {
		return ErrInvalidNumberOfInputValues
	}
	if MapperType.Out(0) != FunctionType.Out(0).Elem() {
		return ErrUnexpectedType
	}
	if CollectionType.Elem() != MapperType.In(0) {
		return ErrUnexpectedType
	}
	if MapperType.NumIn() > 1 {
		if MapperType.In(1) != reflect.TypeOf(1) {
			return ErrUnexpectedType
		}
	}
	if MapperType.NumIn() > 2 {
		if MapperType.In(2) != CollectionType {
			return ErrUnexpectedType
		}
	}
	mapFunction := reflect.MakeFunc(FunctionType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		mapper := args[1]
		results = []reflect.Value{reflect.New(reflect.SliceOf(mapper.Type().Out(0))).Elem()}
		for i := 0; i < collection.Len(); i++ {
			switch mapper.Type().NumIn() {
			case 1:
				results[0] = reflect.Append(results[0], mapper.Call([]reflect.Value{collection.Index(i)})...)
			case 2:
				results[0] = reflect.Append(results[0], mapper.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})...)
			case 3:
				results[0] = reflect.Append(results[0], mapper.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})...)
			}
		}
		return
	})
	Value.Elem().Set(mapFunction)
	return nil

}

// MakeFilter creates a filter function from a pointer to function with the following signatures :
//
//		filter(collection []A, predicate func(element A)bool )[]A
//		filter(collection []A, predicate func(element A, index int)bool )[]A
//		filter(collection []A, predicate func(element A, index int, collection []A)bool )[]A
//
// or return an error if types do not match.
// filter returns a collection of every element for witch
// predicate returns true.
func MakeFilter(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Function := Value.Elem()
	// expect a pointer to function
	if Function.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FunctionType := Function.Type()
	// expect a function with 2 arguments
	if FunctionType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect a function with a single return value
	if FunctionType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FunctionType.In(0)

	if FunctionType.Out(0) != CollectionType {
		return ErrUnexpectedType
	}

	PredicateType := FunctionType.In(1)

	if PredicateType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	if numIn := PredicateType.NumIn(); numIn < 1 || numIn > 3 {
		return ErrInvalidNumberOfInputValues
	}
	if PredicateType.Out(0) != reflect.TypeOf(bool(true)) {
		return ErrUnexpectedType
	}
	if CollectionType.Elem() != PredicateType.In(0) {
		return ErrUnexpectedType
	}
	if PredicateType.NumIn() > 1 {
		if PredicateType.In(1) != reflect.TypeOf(1) {
			return ErrUnexpectedType
		}
	}
	if PredicateType.NumIn() > 2 {
		if PredicateType.In(2) != CollectionType {
			return ErrUnexpectedType
		}
	}
	filterFunction := reflect.MakeFunc(FunctionType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		predicate := args[1]
		results = []reflect.Value{reflect.New(collection.Type()).Elem()}
		for i := 0; i < collection.Len(); i++ {
			var res = false
			switch predicate.Type().NumIn() {
			case 1:
				res = predicate.Call([]reflect.Value{collection.Index(i)})[0].Bool()
			case 2:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})[0].Bool()
			case 3:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})[0].Bool()
			}
			if res {
				results[0] = reflect.Append(results[0], collection.Index(i))
			}
		}
		return
	})
	Value.Elem().Set(filterFunction)
	return nil

}

// MakeEvery creates an every function from a pointer to function with the following signatures :
//
//		every(collection []A, predicate func(element A)bool )bool
//		every(collection []A, predicate func(element A, index int)bool )bool
//		every(collection []A, predicate func(element A, index int, collection []A)bool )bool
//
// or return an error if types do not match.
//
// every returns true if for all elements of collection, predicate returns true, otherwise it returns false
//
func MakeEvery(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Function := Value.Elem()
	// expect a pointer to function
	if Function.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FunctionType := Function.Type()
	// expect a function with 2 arguments
	if FunctionType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect a function with a single return value
	if FunctionType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FunctionType.In(0)

	if FunctionType.Out(0) != reflect.TypeOf(true) {
		return ErrUnexpectedType
	}

	PredicateType := FunctionType.In(1)

	if PredicateType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	if numIn := PredicateType.NumIn(); numIn < 1 || numIn > 3 {
		return ErrInvalidNumberOfInputValues
	}
	if PredicateType.Out(0) != reflect.TypeOf(bool(true)) {
		return ErrUnexpectedType
	}
	if CollectionType.Elem() != PredicateType.In(0) {
		return ErrUnexpectedType
	}
	if PredicateType.NumIn() > 1 {
		if PredicateType.In(1) != reflect.TypeOf(1) {
			return ErrUnexpectedType
		}
	}
	if PredicateType.NumIn() > 2 {
		if PredicateType.In(2) != CollectionType {
			return ErrUnexpectedType
		}
	}
	filterFunction := reflect.MakeFunc(FunctionType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		predicate := args[1]
		results = []reflect.Value{reflect.ValueOf(true)}
		for i := 0; i < collection.Len(); i++ {
			var res bool
			switch predicate.Type().NumIn() {
			case 1:
				res = predicate.Call([]reflect.Value{collection.Index(i)})[0].Bool()
			case 2:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})[0].Bool()
			case 3:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})[0].Bool()
			}
			if !res {
				results[0] = reflect.ValueOf(false)
				return
			}
		}
		return
	})
	Value.Elem().Set(filterFunction)
	return nil

}

// MakeSome creates a some function from a pointer to function with the following signatures :
//
//		some(collection []A, predicate func(element A)bool )bool
//		some(collection []A, predicate func(element A, index int)bool )bool
//		some(collection []A, predicate func(element A, index int, collection []A)bool )bool
//
// or return an error if types do not match.
//
// some returns true if for one element of collection, predicate returns true, otherwise it returns false
//
func MakeSome(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Function := Value.Elem()
	// expect a pointer to function
	if Function.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FunctionType := Function.Type()
	// expect a function with 2 arguments
	if FunctionType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect a function with a single return value
	if FunctionType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FunctionType.In(0)

	if FunctionType.Out(0) != reflect.TypeOf(true) {
		return ErrUnexpectedType
	}

	PredicateType := FunctionType.In(1)

	if PredicateType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	if numIn := PredicateType.NumIn(); numIn < 1 || numIn > 3 {
		return ErrInvalidNumberOfInputValues
	}
	if PredicateType.Out(0) != reflect.TypeOf(bool(true)) {
		return ErrUnexpectedType
	}
	if CollectionType.Elem() != PredicateType.In(0) {
		return ErrUnexpectedType
	}
	if PredicateType.NumIn() > 1 {
		if PredicateType.In(1) != reflect.TypeOf(1) {
			return ErrUnexpectedType
		}
	}
	if PredicateType.NumIn() > 2 {
		if PredicateType.In(2) != CollectionType {
			return ErrUnexpectedType
		}
	}
	filterFunction := reflect.MakeFunc(FunctionType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		predicate := args[1]
		results = []reflect.Value{reflect.ValueOf(false)}
		for i := 0; i < collection.Len(); i++ {
			var res bool
			switch predicate.Type().NumIn() {
			case 1:
				res = predicate.Call([]reflect.Value{collection.Index(i)})[0].Bool()
			case 2:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})[0].Bool()
			case 3:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})[0].Bool()
			}
			if res {
				results[0] = reflect.ValueOf(true)
				return
			}
		}
		return
	})
	Value.Elem().Set(filterFunction)
	return nil

}

// MakeFind creates a find function from a pointer to function with the following signatures :
//
//		find(collection []A, predicate func(element A)bool )(result A,index int)
//		find(collection []A, predicate func(element A, index int)bool )(result A, index int)
//		find(collection []A, predicate func(element A, index int, collection []A)bool )(result A, index int)
//
// or return an error if types do not match.
//
// find returns an element of the collection and its index if for that element, predicate returns true,
// otherwise the zero value of the collection's element type and -1 .
//
func MakeFind(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	Function := Value.Elem()
	// expect a pointer to function
	if Function.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FunctionType := Function.Type()
	// expect a function with 2 arguments
	if FunctionType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect a function with a single return value
	if FunctionType.NumOut() != 2 {
		return ErrInvalidNumberOfReturnValues
	}
	CollectionType := FunctionType.In(0)

	if FunctionType.Out(0) != CollectionType.Elem() {
		return ErrUnexpectedType
	}
	if FunctionType.Out(1) != reflect.TypeOf(1) {
		return ErrUnexpectedType
	}

	PredicateType := FunctionType.In(1)

	if PredicateType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	if numIn := PredicateType.NumIn(); numIn < 1 || numIn > 3 {
		return ErrInvalidNumberOfInputValues
	}
	if PredicateType.Out(0) != reflect.TypeOf(bool(true)) {
		return ErrUnexpectedType
	}
	if CollectionType.Elem() != PredicateType.In(0) {
		return ErrUnexpectedType
	}
	if PredicateType.NumIn() > 1 {
		if PredicateType.In(1) != reflect.TypeOf(1) {
			return ErrUnexpectedType
		}
	}
	if PredicateType.NumIn() > 2 {
		if PredicateType.In(2) != CollectionType {
			return ErrUnexpectedType
		}
	}
	filterFunction := reflect.MakeFunc(FunctionType, func(args []reflect.Value) (results []reflect.Value) {
		collection := args[0]
		predicate := args[1]
		results = []reflect.Value{reflect.Zero(collection.Type().Elem()), reflect.ValueOf(-1)}
		for i := 0; i < collection.Len(); i++ {
			var res bool
			switch predicate.Type().NumIn() {
			case 1:
				res = predicate.Call([]reflect.Value{collection.Index(i)})[0].Bool()
			case 2:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})[0].Bool()
			case 3:
				res = predicate.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})[0].Bool()
			}
			if res {
				results[0] = collection.Index(i)
				results[1] = reflect.ValueOf(i)
				return
			}
		}
		return
	})
	Value.Elem().Set(filterFunction)
	return nil

}

// MakeIndexOf creates an indexOf function from a pointer to function using the following signature :
//
// 		indexOf([]T,T)int
//
// or returns an error if types do not match
func MakeIndexOf(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	// expect the pointer to be a function
	if Value.Elem().Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FuncValue := Value.Elem()
	FuncType := FuncValue.Type()
	// expect the function to have 2 arguments
	if FuncType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect the function to have 1 return value
	if FuncType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	// expect the return value to be an integer
	if FuncType.Out(0) != reflect.TypeOf(int(0)) {
		return ErrUnexpectedType
	}
	FirstArgumentType := FuncType.In(0)

	// expect the first argument to be an array
	if kind := FirstArgumentType.Kind(); kind != reflect.Array && kind != reflect.Slice {
		return ErrUnexpectedType
	}
	FirstArgumentElementType := FirstArgumentType.Elem()
	if !FirstArgumentElementType.Comparable() {
		return ErrNoComparableType
	}
	SecondArgumentType := FuncType.In(1)
	// expect the element of the array of the first argument and the second argument to
	// have matching types
	if SecondArgumentType != FirstArgumentElementType {
		return ErrUnexpectedType
	}
	Result := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		ArrayValue := args[0]
		NeedleValue := args[1]
		for i := 0; i < ArrayValue.Len(); i++ {
			if ArrayValue.Index(i).Interface() == NeedleValue.Interface() {
				results = []reflect.Value{reflect.ValueOf(i)}
				return
			}
		}
		results = []reflect.Value{reflect.ValueOf(-1)}
		return
	})
	Value.Elem().Set(Result)
	return nil
}

// MakeInclude creates an include function from a pointer to function using the following signature :
//
//		include([]T,T)bool
//
// or returns an error if types do not match.
//
// include returns true if element T exists in collection []T ,else returns false.
// T must be a comparable type.
func MakeInclude(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	// expect a pointer
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	// expect the pointer to be a function
	if Value.Elem().Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FuncValue := Value.Elem()
	FuncType := FuncValue.Type()
	// expect the function to have 2 arguments
	if FuncType.NumIn() != 2 {
		return ErrInvalidNumberOfInputValues
	}
	// expect the function to have 1 return value
	if FuncType.NumOut() != 1 {
		return ErrInvalidNumberOfReturnValues
	}
	// expect the return value to be an boolean
	if FuncType.Out(0) != reflect.TypeOf(bool(true)) {
		return ErrUnexpectedType
	}
	FirstArgumentType := FuncType.In(0)

	// expect the first argument to be an array
	if kind := FirstArgumentType.Kind(); kind != reflect.Array && kind != reflect.Slice {
		return ErrUnexpectedType
	}
	FirstArgumentElementType := FirstArgumentType.Elem()
	if !FirstArgumentElementType.Comparable() {
		return ErrNoComparableType
	}
	SecondArgumentType := FuncType.In(1)
	// expect the element of the array of the first argument and the second argument to
	// have matching types
	if SecondArgumentType != FirstArgumentElementType {
		return ErrUnexpectedType
	}
	Result := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		ArrayValue := args[0]
		NeedleValue := args[1]
		for i := 0; i < ArrayValue.Len(); i++ {
			if ArrayValue.Index(i).Interface() == NeedleValue.Interface() {
				results = []reflect.Value{reflect.ValueOf(true)}
				return
			}
		}
		results = []reflect.Value{reflect.ValueOf(false)}
		return
	})
	Value.Elem().Set(Result)
	return nil
}

// MakeMapToArray assigns a mapToArray function with the following signature :
//
//		mapToArray(Map map[K]V, mapper func(V)T)[]T
//		mapToArray(Map map[K]V, mapper func(V,K)T)[]T
//		mapToArray(Map map[K]V, mapper func(V,K,map[K]V)T)[]T
//
// or returns an error if types do not match.
//
// mapToArray applies mapper to each value of a map and returns an array of the return values of the mapper.
func MakeMapToArray(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	FuncValue := Value.Elem()
	if FuncValue.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FuncType := FuncValue.Type()
	mapToArray := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		Map := args[0]
		mapper := args[1]
		mapperInLength := mapper.Type().NumIn()
		returnCollection := reflect.MakeSlice(reflect.SliceOf(mapper.Type().Out(0)), 0, 0)
		for _, key := range Map.MapKeys() {
			var returnValues []reflect.Value
			switch mapperInLength {
			case 1:
				returnValues = mapper.Call([]reflect.Value{Map.MapIndex(key)})
			case 2:
				returnValues = mapper.Call([]reflect.Value{Map.MapIndex(key), key})
			case 3:
				returnValues = mapper.Call([]reflect.Value{Map.MapIndex(key), key, Map})
			}
			returnCollection = reflect.Append(returnCollection, returnValues[0])
		}
		results = []reflect.Value{returnCollection}
		return
	})
	Value.Elem().Set(mapToArray)
	return nil
}

// MakeFlatten assigns a flatten function to pointerToFunction with the following signature :
//
// 		flaten([][]T)[]T
//
// or return an error if types do not match.
// flatten flattens an array or arrays into an array
func MakeFlatten(pointerToFunction interface{}) error {
	Value := reflect.ValueOf(pointerToFunction)
	if Value.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	FuncValue := Value.Elem()
	if FuncValue.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	FuncType := FuncValue.Type()
	flatten := reflect.MakeFunc(FuncType, func(args []reflect.Value) (results []reflect.Value) {
		collectionOfCollection := args[0]
		collection := collectionOfCollection.Elem()
		returnCollection := reflect.MakeSlice(collection.Elem().Type(), 0, 0)
		for i := 0; i < collectionOfCollection.Len(); i++ {
			collection := collectionOfCollection.Index(i)
			for j := 0; j < collection.Len(); j++ {
				returnCollection = reflect.Append(returnCollection, collection.Index(j))
			}
		}
		results = []reflect.Value{returnCollection}
		return
	})
	Value.Elem().Set(flatten)
	return nil
}
