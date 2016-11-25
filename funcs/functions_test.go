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

package funcs_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/Mparaiso/go-tiger/funcs"
	"github.com/Mparaiso/go-tiger/test"
)

func BenchmarkMap_Without_Reflection(b *testing.B) {
	type Person struct {
		Name string
		Age  int
	}
	var (
		persons = []Person{{"Joe", 18}, {"Kim", 29}, {"Jane", 38}, {"Jack", 24}, {"David", 79}}
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		names := []string{}
		for j := 0; j < len(persons); j++ {
			names = append(names, persons[j].Name)
		}

	}
}

func BenchmarkMap_MakeMap(b *testing.B) {
	type Person struct {
		Name string
		Age  int
	}
	var (
		mapPersonsToNames func(persons []Person, mapper func(person Person) string) []string
		persons           = []Person{{"Joe", 18}, {"Kim", 29}, {"Jane", 38}, {"Jack", 24}, {"David", 79}}
	)
	if err := funcs.MakeMap(&mapPersonsToNames); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mapPersonsToNames(persons, func(person Person) string {
			return person.Name
		})
	}
}
func TestErrNotApointer(t *testing.T) {
	var reduce func([]int, func(int, int) int, int)
	err := funcs.MakeReduce(reduce)
	test.Fatal(t, err, funcs.ErrNotAPointer)
}

func TestErrNotAFunction(t *testing.T) {
	var reduce interface{}
	err := funcs.MakeReduce(&reduce)
	test.Fatal(t, err, funcs.ErrNotAFunction)
}

func TestErrInvalidNumberOfInputValues(t *testing.T) {
	var reduce func(collection []string, reducer func(result string, element string) string)
	err := funcs.MakeReduce(&reduce)
	test.Fatal(t, err, funcs.ErrInvalidNumberOfInputValues)
}

func ExampleMakeGroupBy() {
	// let's group people by sex
	type Sex int
	const (
		male Sex = iota
		female
	)
	type Person struct {
		Name string
		Sex  Sex
	}
	var groupPeopleBySex func(people []Person, selector func(person Person) Sex) map[Sex][]Person
	if err := funcs.MakeGroupBy(&groupPeopleBySex); err != nil {
		log.Fatal(err)
	}
	people := []Person{{"Alex", female}, {"John", male}, {"David", male}, {"Doris", female}, {"Jack", male}}
	peopleBySex := groupPeopleBySex(people, func(person Person) Sex {
		return person.Sex
	})
	fmt.Println(len(peopleBySex))
	fmt.Println(len(peopleBySex[male]))
	fmt.Println(len(peopleBySex[female]))
	// Output:
	// 2
	// 3
	// 2
}
func ExampleMakeReduce() {
	// Let's use MakeReduce to create a map function.
	type Person struct {
		Name string
		Age  int
	}
	var MapPersonsToNames func(persons []Person, reducer func(names []string, person Person) []string, initial []string) []string
	err := funcs.MakeReduce(&MapPersonsToNames)
	if err != nil {
		log.Fatal(err)
	}
	result := MapPersonsToNames([]Person{{"Frank", 30}, {"John", 43}, {"Jane", 26}}, func(names []string, person Person) []string {
		return append(names, person.Name)
	}, []string{})
	fmt.Println(result)
	// Output:
	// [Frank John Jane]
}
func ExampleMakeReduce_Second() {
	// The reducer can take up to 4 arguments.
	var reduceStringsToString func(strings []string, reducer func(result string, element string, index int, strings []string) string, initial string) string
	fmt.Println(funcs.MakeReduce(&reduceStringsToString))
	fmt.Println(reduceStringsToString([]string{"foo", "bar"}, func(result, element string, index int, strings []string) string {
		return result + fmt.Sprint(index) + element
	}, ""))
	// Output:
	// <nil>
	// 0foo1bar
}

func ExampleMakeMap() {
	// Let's map book titles
	type Book struct {
		Title string
		ID    int
	}
	var mapBooksToStrings func([]Book, func(Book) string) []string
	if err := funcs.MakeMap(&mapBooksToStrings); err != nil {
		log.Fatal(err)
	}
	books := []Book{{"Les Misérables", 34043}, {"Germinal", 349439}}
	fmt.Println(mapBooksToStrings(books, func(book Book) string {
		return book.Title
	}))
	// Output:
	// [Les Misérables Germinal]
}

func ExampleMakeMap_Second() {
	// Let's map the person fullnames to an array of names
	type Person struct{ FirstName, LastName string }
	var mapPersonsToStrings func([]Person, func(Person, int) string) []string
	if err := funcs.MakeMap(&mapPersonsToStrings); err != nil {
		log.Fatal(err)
	}
	people := []Person{{"John", "Doe"}, {"Jane", "Doe"}}
	fmt.Println(mapPersonsToStrings(people, func(p Person, index int) string {
		return fmt.Sprintf("#%d %s %s", index, p.FirstName, p.LastName)
	}))
	// Output:
	// [#0 John Doe #1 Jane Doe]
}

func TestMakeIndexOfErrors(t *testing.T) {
	var IndexOfInts func([]int, int) int
	err := funcs.MakeIndexOf(IndexOfInts)
	test.Fatal(t, err, funcs.ErrNotAPointer)
	err = funcs.MakeIndexOf(&struct{}{})
	test.Fatal(t, err, funcs.ErrNotAFunction)
	var IndexOfString func([]string, int) int
	err = funcs.MakeIndexOf(&IndexOfString)
	test.Fatal(t, err, funcs.ErrUnexpectedType)
	var IndexOfByte func([]byte, byte) string
	err = funcs.MakeIndexOf(&IndexOfByte)
	test.Fatal(t, err, funcs.ErrUnexpectedType)
	var IndexOfArray func([][]string, []string) int
	err = funcs.MakeIndexOf(&IndexOfArray)
	test.Fatal(t, err, funcs.ErrNoComparableType)
}

func ExampleMakeIndexOf() {
	// Let's make an indexOf function
	var IndexOfInts func([]int, int) int
	fmt.Println(funcs.MakeIndexOf(&IndexOfInts))
	fmt.Println(IndexOfInts([]int{1, 2, 4}, 2))
	fmt.Println(IndexOfInts([]int{2, 6, 8}, 1))

	// Output:
	// <nil>
	// 1
	// -1
}

func ExampleMakeSome() {
	var someStrings func(collection []string, predicate func(s string) bool) bool
	if err := funcs.MakeSome(&someStrings); err != nil {
		log.Fatal(err)
	}
	fmt.Println(someStrings([]string{"foo", "bar", "baz"}, func(s string) bool {
		return s[0] == 'f'
	}))
	fmt.Println(someStrings([]string{"foo", "bar", "baz"}, func(s string) bool {
		return s[0] == 'a'
	}))

	// Output:
	// true
	// false
}

func ExampleMakeEvery() {
	var everyInts func(collection []int, predicate func(i int) bool) bool

	if err := funcs.MakeEvery(&everyInts); err != nil {
		log.Fatal(err)
	}

	fmt.Println(everyInts([]int{1, 2, 3}, func(element int) bool {
		return element%2 == 0
	}))

	fmt.Println(everyInts([]int{1, 3, 5}, func(element int) bool {
		return element%2 != 0
	}))

	// Output:
	// false
	// true

}

func ExampleMakeFilter() {
	type Person struct {
		Age  int
		Name string
	}
	var filterAdults func(persons []Person, predicate func(person Person, i int) bool) []Person
	fmt.Println(funcs.MakeFilter(&filterAdults))
	// TODO: handle error
	people := []Person{{18, "Joe"}, {26, "David"}, {15, "Anna"}}
	adults := filterAdults(people, func(person Person, index int) bool {
		return person.Age >= 18
	})
	fmt.Println(len(adults))
	fmt.Println(adults[0].Age, adults[0].Name)

	// Output:
	// <nil>
	// 2
	// 18 Joe
}

func ExampleMakeInclude() {
	// Let's create an include function
	var StringsInclude func([]string, string) bool
	fmt.Println(funcs.MakeInclude(&StringsInclude))

	fmt.Println(StringsInclude([]string{"a", "b", "d", "e"}, "e"))
	fmt.Println(StringsInclude([]string{"e", "k", "f"}, "g"))

	// Output:
	// <nil>
	// true
	// false
}

func ExampleMakeForEach() {
	// let's create a for each function
	var forEachString func([]string, func(string, int, []string))
	if err := funcs.MakeForEach(&forEachString); err != nil {
		log.Fatal(err)
	}
	forEachString([]string{"a", "b", "c", "d"}, func(s string, i int, strings []string) {
		fmt.Print(s, ":", i)
		if i < len(strings)-1 {
			fmt.Print(",")
		}
	})
	// Output:
	// a:0,b:1,c:2,d:3
}

func ExampleMakeFind() {
	// Let's create a find function
	type City struct {
		Name string
		Code int
	}
	var findCity func([]City, func(city City) bool) (city City, index int)

	if err := funcs.MakeFind(&findCity); err != nil {
		log.Fatal(err)
	}

	cities := []City{{"Albany", 349}, {"Portland", 556}, {"Boston", 494}}
	// if found , it returns the found element and its index
	city, index := findCity(cities, func(city City) bool {
		return city.Code == 494
	})
	fmt.Println(city)
	fmt.Println(index)
	// if not found, it returns the zero value of a city element and -1
	city, index = findCity(cities, func(city City) bool {
		return city.Code == 690
	})
	fmt.Println(city)
	fmt.Println(index)

	// Output:
	// {Boston 494}
	// 2
	// { 0}
	// -1
}
