package funcs_test

import (
	"fmt"
	"testing"

	"github.com/Mparaiso/go-tiger/funcs"
	"github.com/Mparaiso/go-tiger/test"
)

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
func ExmapleMakeReduce() {
	// Let's use MakeReduce to create a map function.
	type Person struct {
		Name string
		Age  int
	}
	var MapPersonsToNames func(persons []Person, reducer func(names []string, person Person) []string, initial []string) string
	err := funcs.MakeReduce(&MapPersonsToNames)
	fmt.Println(err)
	result := MapPersonsToNames([]Person{{"Frank", 30}, {"John", 43}, {"Jane", 26}}, func(names []string, person Person) []string {
		return append(names, person.Name)
	}, []string{})
	fmt.Println(result)
	// Output:
	// <nil>
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
	fmt.Println(funcs.MakeMap(&mapBooksToStrings))
	fmt.Println(mapBooksToStrings([]Book{{"Les Misérables", 34043}, {"Germinal", 349439}}, func(book Book) string {
		return book.Title
	}))
	// Output:
	// <nil>
	// [Les Misérables Germinal]
}

func ExampleMakeMap_Second() {
	// Let's map the person fullnames to an array of names
	type Person struct{ FirstName, LastName string }
	var mapPersonsToStrings func([]Person, func(Person, int) string) []string
	fmt.Println(funcs.MakeMap(&mapPersonsToStrings))
	fmt.Println(mapPersonsToStrings([]Person{{"John", "Doe"}, {"Jane", "Doe"}}, func(p Person, index int) string {
		return fmt.Sprintf("#%d %s %s", index, p.FirstName, p.LastName)
	}))
	// Output:
	// <nil>
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
