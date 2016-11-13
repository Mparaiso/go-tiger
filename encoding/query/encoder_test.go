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
package query_test

import (
	"fmt"

	"github.com/Mparaiso/go-tiger/encoding/query"
)

func ExampleToValues() {
	// This example demonstrates how to use query.ToValues()
	// to convert a struct into url.Values

	// Given a type Person
	type Person struct {
		Age       uint
		FirstName string
		LastName  string
		// we can use a struct tag to explicitely set the key of the query parameter
		Geolocation [2]float32 `schema:"location"`
		// we can also ignore fields in a struct with -
		Children []Person `schema:"-"`
	}
	p := Person{28, "John", "Doe", [2]float32{100.02, 50.34}, []Person{}}
	// let's convert it into url.Values
	values, err := query.ToValues(p)
	fmt.Println(err)
	fmt.Println(values.Get("Age"))
	fmt.Println(len(values["location"]))
	// The final result should look like the following query string
	// values.Encode() => "Age=28&FirstName=John&LastName=Doe&location=100.02&location=50.34"

	// Output:
	// <nil>
	// 28
	// 2
}
