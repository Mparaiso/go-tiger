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
package funcs

import (
	"errors"
	"reflect"
)

var (
	// ErrNotAPointer : Error the value is not a pointer
	ErrNotAPointer = errors.New("ErrNotAPointer : Error the value is not a pointer")
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
// it returns an error conveniantly so it can be used in a
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
//
// Example:
//
//      var sumReduce func(ints []int{},reducer func(result int,element int)int,initial int)int
//      err := MakeReduce(&sumReduce)
//      // TODO: Handle error
//      result := sumReduce([]int{1,2,3},func(result,element int)int{
//          return result + element
//      })
//      fmt.Print(result)
//      // Output:
//      // 1
//
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
//
// Example:
//
//		type Person struct { Name string }
//		var mapPersonsToStrings func(persons []Person,mapper func(person Person)string)[]string
//		err := funcs.MakeMap(&mapPersonsToStrings)
//		// TODO: handle error
//		fmt.Print(mapPersonsToStrings([]Person{{"Joe"},{"David"}},func(person Person)string{
//			return person.Name
//		}))
//		// Output:
//		// [Joe Davis]
//
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
//
// Example:
//
//		type Person struct { Age int ; Name string }
//		var filterAdults func(persons []Person,predicate func(person Person)bool)[]Person
//		err := funcs.MakeMap(&filterAdults)
//		// TODO: handle error
//		people := []Person{{18,"Joe"},{26,"David"},{15,"Anna"}}
//		adults := filterAdults(people,func(person Person)bool{
//			return person.Age >= 18
//		})
//		fmt.Println(len(adults))
//		// Output:
//		// 2
//
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
// include returns true if element T exists in collection []T ,else returns false.
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
