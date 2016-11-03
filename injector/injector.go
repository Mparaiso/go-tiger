//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//      http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package injector

import (
	"fmt"
	"reflect"

	"github.com/Mparaiso/go-tiger/logger"
)

// Injector is an depencency injection container
type Injector struct {
	factories map[*FactoryData]reflect.Value
	Logger    logger.Logger
	parent    *Injector
}

// NewInjector return a new Injector
func NewInjector() *Injector {
	return &Injector{factories: map[*FactoryData]reflect.Value{}}
}

func (i *Injector) setParent(injector *Injector) {
	i.parent = injector
}

// CreateChild creates a child Injector
// that inherits from the current Injector
func (i *Injector) CreateChild() *Injector {
	child := NewInjector()
	child.setParent(i)
	return child
}

// SetLogger sets the logger for debugging purposes
func (i *Injector) SetLogger(Logger logger.Logger) {
	i.Logger = Logger
}

func (i Injector) logError(args ...interface{}) {
	if i.Logger != nil {
		i.Logger.Log(logger.Error, args...)
	}
}

func (i Injector) logDebug(args ...interface{}) {
	if i.Logger != nil {
		i.Logger.Log(logger.Debug, args...)
	}
}

func (i Injector) getFactoryReturnType(factoryType reflect.Type) (reflect.Type, error) {
	if !isFunction(factoryType) {
		i.logError(factoryType, ErrorNotAFunction)
		return nil, ErrorNotAFunction
	}
	numberOfReturnValues := factoryType.NumOut()
	if numberOfReturnValues != 2 {
		return nil, ErrorInvalidReturnValueNumber
	}
	if !factoryType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return nil, ErrorinvalidReturnType
	}
	return factoryType.Out(0), nil

}
func (Injector) isPointer(Type reflect.Type) bool {
	return Type.Kind() == reflect.Ptr
}

// RegisterFactory a service factory with the following signature
// func(dependency1 Type1,dependency2 Type2,...)(ServiceType,error)
func (i *Injector) RegisterFactory(factory Function, tag ...string) error {
	factoryType := reflect.TypeOf(factory)
	Type, err := i.getFactoryReturnType(factoryType)
	if err != nil {
		return err
	}

	factoryValue := reflect.ValueOf(factory)
	factoryData := &FactoryData{Type: Type}
	factoryData.Kind = Type.Kind()
	if len(tag) > 0 && tag[0] != "" {
		factoryData.Tag = tag[0]
	}
	// delete duplicate if exists
	if data := i.FindFactoryData(Type, tag...); data != nil {
		delete(i.factories, data)
	}
	i.factories[factoryData] = factoryValue
	return nil
}

// RegisterValue a value like a string, an integer, a struct ...
func (i *Injector) RegisterValue(value Value, tag ...string) {
	valueValue := reflect.ValueOf(value)
	valueType := valueValue.Type()
	factoryData := &FactoryData{Type: valueType, Kind: valueValue.Kind(), Resolved: true}
	if len(tag) > 0 && tag[0] != "" {
		factoryData.Tag = tag[0]
	}
	// delete duplicate if exists
	if data := i.FindFactoryData(valueType, tag...); data != nil {
		delete(i.factories, data)
	}
	i.factories[factoryData] = valueValue
}

// MustRegisterFactory panics on error
// otherwise It is the same as *Injector.Register
func (i *Injector) MustRegisterFactory(factory Function, tag ...string) *Injector {
	err := i.RegisterFactory(factory, tag...)
	if err != nil {
		panic(err)
	}
	return i
}

// HasService checks if a service was already registered
func (i *Injector) HasService(Type reflect.Type, tag ...string) bool {
	for data := range i.factories {
		if data.Type == Type && ((len(tag) > 0 && tag[0] != "" && data.Tag == tag[0]) || len(tag) == 0) {
			return true
		}
	}
	return false
}

// FindFactoryData finds FactoryData by Type and tag
func (i *Injector) FindFactoryData(Type reflect.Type, tag ...string) *FactoryData {
	for data := range i.factories {
		if data.Type == Type && ((len(tag) > 0 && tag[0] != "" && data.Tag == tag[0]) || (len(tag) == 0 && data.Tag == "")) {
			i.logDebug("Found factory data", fmt.Sprintf("%+v", data))
			return data
		}
	}
	return nil
}

// Resolve populates the pointer with a value if
// a corresponding service is found in the injector
// or return an error if needed
func (i *Injector) Resolve(pointer Pointer, tag ...string) error {
	pointerType := reflect.TypeOf(pointer)
	if !i.isPointer(pointerType) {
		i.logError(ErrorNotAPointer, pointerType)
		return ErrorNotAPointer
	}
	for factoryData, value := range i.factories {
		i.logDebug("Comparing argument type :", pointerType.Elem(), "with tag :", tag, " against injector service of type :", factoryData.Type, "with tag :", factoryData.Tag)
		// check if tags match
		if (len(tag) > 0 && tag[0] != "" && factoryData.Tag != tag[0]) || (len(tag) == 0 && factoryData.Tag != "") {
			continue
		}
		// check if pointer's element can be populated with the type of the service
		switch {
		case factoryData.Type == pointerType.Elem():
		case factoryData.Kind == reflect.Interface && factoryData.Type.AssignableTo(pointerType.Elem()):
		case pointerType.Elem().Kind() == reflect.Interface && factoryData.Type.AssignableTo(pointerType.Elem()):
		default:
			continue
		}
		// resolve factories if not resolved
		if !factoryData.Resolved {
			if factoryData.Visited {
				i.logError(ErrorCircularDependency, " resolved type : '", pointerType.Elem(), "' , visited type : '", factoryData.Type, "' ")
				return ErrorCircularDependency
			}
			factoryData.Visited = true
			result, err := i.doResolve(value)
			if err != nil {
				return err
			}

			i.factories[factoryData] = result
			factoryData.Resolved = true
		}
		// do assign
		pointerValue := reflect.ValueOf(pointer)
		if pointerValue == reflect.Zero(pointerType) {
			return fmt.Errorf("Error pointer is a zero value %s", pointerType)
		}
		pointerValue.Elem().Set(i.factories[factoryData])
		return nil
	}
	// Nothing found in this injector, let's try to parent injector
	if i.parent != nil {
		return i.parent.Resolve(pointer, tag...)
	}
	// Nothing found and no parent, return an error
	i.logError(ErrorServiceNotFound, "type :'", pointerType.Elem(), "' tag :", tag)
	return ErrorServiceNotFound
}

// Populate populates each field of a struct with the corresponding service.
// if a field has a tag `injector:"tagname"` , the injector will search for tagname
// as well
func (i *Injector) Populate(target Struct) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return ErrorNotAPointer
	}
	targetStructType := targetType.Elem()
	if targetStructType.Kind() != reflect.Struct {
		return ErrorNotAStruct
	}
	targetStructValue := reflect.ValueOf(target).Elem()
	for j := 0; j < targetStructValue.NumField(); j++ {
		fieldValue := targetStructValue.Field(j)
		tag := targetStructType.Field(j).Tag.Get("injector")
		pointer := fieldValue.Addr()
		err := i.Resolve(pointer.Interface(), tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Injector) doResolve(value reflect.Value) (reflect.Value, error) {

	results, err := i.doCall(value) //value.Call([]reflect.Value{})
	if err != nil {
		return reflect.Value{}, err
	}
	err1 := results[1].Interface()
	if err1 != nil {
		return reflect.Value{}, err1.(error)
	}
	return results[0], nil
}

// Call resolve each argument of a function to a service in the injector
// then call the function.
// Optional results are put in argumentPointer which is an array of pointers
func (i *Injector) Call(function Function, argumentPointer ...Pointer) error {
	functionType := reflect.TypeOf(function)
	if !isFunction(functionType) {
		return ErrorNotAFunction
	}
	functionValue := reflect.ValueOf(function)
	results, err := i.doCall(functionValue)
	if err != nil {
		return err
	}
	if len(argumentPointer) > len(results) {
		return fmt.Errorf("argumentPointer length and the number of outputs in the called function mismatch")
	}
	for j, pointer := range argumentPointer {
		pointerType := reflect.TypeOf(pointer)
		if pointerType.Kind() != reflect.Ptr {
			return ErrorNotAPointer
		}
		if !results[j].Type().AssignableTo(pointerType.Elem()) {
			return fmt.Errorf("Can't assign %s to %s ", results[j].Type(), pointerType.Elem())
		}
		reflect.ValueOf(pointer).Elem().Set(results[j])
	}
	return nil
}

func (i *Injector) doCall(function reflect.Value) ([]reflect.Value, error) {
	functionType := function.Type()
	argumentTypes := []reflect.Type{}
	for j := 0; j < functionType.NumIn(); j++ {
		argumentTypes = append(argumentTypes, functionType.In(j))
	}
	argumentValues := []reflect.Value{}
	for _, argumentType := range argumentTypes {
		argumentValue := reflect.New(argumentType)
		err := i.Resolve(argumentValue.Interface())
		if err != nil {
			return nil, err
		}
		argumentValues = append(argumentValues, argumentValue.Elem())
	}
	results := function.Call(argumentValues)
	return results, nil
}

// Pointer is a semantic type
// that represents a pointer
// such as '&value'
type Pointer interface{}

// Function is a semantic type that
// represents a function such as
// 'func(type1)(returnType1,error)'
type Function interface{}

// Value is a semantic type that represents any value
type Value interface{}

// Struct is a semantic type that represents a struct
type Struct interface{}

func isFunction(f reflect.Type) bool {
	return f.Kind() == reflect.Func
}

var (
	// ErrorNotAFunction is returned when an value is not of function kind
	ErrorNotAFunction = fmt.Errorf("Value is not a function.")
	// ErrorInvalidReturnValueNumber is returned when A factory doesn't return 2 values.
	ErrorInvalidReturnValueNumber = fmt.Errorf("A factory should return 2 values.")
	// ErrorinvalidReturnType is returned when the second return value of a factory is not of type error
	ErrorinvalidReturnType = fmt.Errorf("The second return value of a factory should be of type error.")
	// ErrorNotAPointer is returned when the value passed as an argument is not a pointer
	ErrorNotAPointer = fmt.Errorf("Value is not a pointer")
	// ErrorServiceNotFound is returned when a service was not found in the injector
	ErrorServiceNotFound = fmt.Errorf("Service not found in injector")
	// ErrorNotAStruct is returned when a struct was expected but not provided
	ErrorNotAStruct = fmt.Errorf("Value is not a struct")
	// ErrorCircularDependency is returned when the resolved type calls a non resolved type that has already been visited
	ErrorCircularDependency = fmt.Errorf("Circular dependeny detected")
)

// FactoryData represents the data of a factory in the injector
type FactoryData struct {
	Resolved bool
	Type     reflect.Type
	Tag      string
	Kind     reflect.Kind
	Visited  bool
}
