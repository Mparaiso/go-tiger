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

package injector_test

import (
	"fmt"
	"testing"

	"github.com/Mparaiso/go-tiger/injector"
	"github.com/Mparaiso/go-tiger/test"
)

func TestInjector(t *testing.T) {
	type Foo struct{ Value string }
	container := injector.NewInjector()
	test.Fatal(t, container != nil, true)
	err := container.RegisterFactory(func() (string, error) {
		return "Hello", nil
	})
	test.Fatal(t, err, nil)
	err = container.RegisterFactory(func() (*Foo, error) {
		return &Foo{"Foo"}, nil
	})
	test.Fatal(t, err, nil)
	var message string
	err = container.Resolve(&message)
	test.Fatal(t, err, nil)
	test.Fatal(t, message, "Hello")
	var foo *Foo
	err = container.Resolve(&foo)
	test.Fatal(t, err, nil)
	test.Fatal(t, foo.Value, "Foo")
}

type Interface interface {
	Do()
}

type Foo struct{}

func (Foo) Do() {}

func TestInjector_Resolve_Interface(t *testing.T) {
	container := injector.NewInjector()
	err := container.RegisterFactory(func() (Interface, error) {
		return &Foo{}, nil
	})
	test.Fatal(t, err, nil)
	var i Interface
	err = container.Resolve(&i)
	test.Fatal(t, err, nil)

	container = injector.NewInjector()
	err = container.RegisterFactory(func() (*Foo, error) {
		return &Foo{}, nil
	})
	test.Fatal(t, err, nil)
	err = container.Resolve(&i)
	test.Fatal(t, err, nil)
}

func TestInjector_RegisterValue(t *testing.T) {
	var s string
	type id string
	container := injector.NewInjector()
	container.RegisterValue("foo")
	err := container.Resolve(&s)
	test.Fatal(t, err, nil)
	test.Fatal(t, s, "foo")
	container = injector.NewInjector()
	container.RegisterValue("bar", "bar")
	err = container.Resolve(&s)
	test.Fatal(t, err, injector.ErrorServiceNotFound)
	container = injector.NewInjector()
	container.SetLogger(test.NewTestLogger(t))

	container.RegisterValue(id("some_id"))
	err = container.Resolve(&s)
	test.Fatal(t, err, injector.ErrorServiceNotFound)
	var i id
	err = container.Resolve(&i)
	test.Fatal(t, err, nil)
	test.Fatal(t, i, id("some_id"))
}

func TestInjector_Populate(t *testing.T) {
	type Target struct {
		ID   int    `injector:"id"`
		Name string `injector:"name"`
	}
	var target Target
	container := injector.NewInjector()
	container.SetLogger(test.NewTestLogger(t))
	container.RegisterValue(10, "id")
	container.RegisterValue(20)
	container.RegisterValue("john doe", "name")
	container.RegisterValue("jane doe")
	err := container.Populate(&target)
	test.Fatal(t, err, nil)
	test.Fatal(t, target.ID, 10)
	test.Fatal(t, target.Name, "john doe")
}

func TestInjector_Call(t *testing.T) {
	container := injector.NewInjector()
	container.SetLogger(test.NewTestLogger(t))
	container.RegisterValue(10)
	container.RegisterValue("foo")
	var result1 string
	err := container.Call(func(a int, b string) string {
		return fmt.Sprint(a, b)
	}, &result1)
	test.Fatal(t, err, nil)
	test.Fatal(t, result1, "10foo")
}

func TestInjector_Resolve_Factory(t *testing.T) {
	container := injector.NewInjector()
	container.SetLogger(test.NewTestLogger(t))
	err := container.RegisterFactory(func(a int, b string) (string, error) {
		return fmt.Sprint(a, b), nil
	}, "service")
	test.Fatal(t, err, nil)
	container.RegisterValue(5)
	container.RegisterValue("bar")
	var result string
	err = container.Resolve(&result, "service")
	test.Fatal(t, err, nil)
	test.Fatal(t, result, "5bar")

}

func TestInjector_Resolve_CircularDependency(t *testing.T) {
	container := injector.NewInjector()
	container.SetLogger(test.NewTestLogger(t))
	err := container.RegisterFactory(func(i int) (int, error) {
		return i, nil
	})
	test.Fatal(t, err, nil)
	var i int
	err = container.Resolve(&i)
	test.Fatal(t, err, injector.ErrorCircularDependency)
}

func TestInjector_Resolve_Parent(t *testing.T) {
	type Person struct{ Name string }
	container := injector.NewInjector()
	container.SetLogger(test.NewTestLogger(t))
	container.RegisterValue("foo")
	childContainer := container.CreateChild()
	err := childContainer.RegisterFactory(func(name string) (Person, error) {
		return Person{name}, nil
	})
	test.Fatal(t, err, nil)
	person := &Person{}
	err = childContainer.Resolve(person)
	test.Fatal(t, err, nil)
	test.Fatal(t, person.Name, "foo")
}
