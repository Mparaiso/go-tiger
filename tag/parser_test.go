package tag_test

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/Mparaiso/go-tiger/logger"
	"github.com/Mparaiso/go-tiger/tag"
	"github.com/Mparaiso/go-tiger/test"
)

func ExampleParser() {
	// Let's parse the meta tag in Title field of type Book
	type Book struct {
		Title string `meta:"isbn:11010303030;published;author(firstname:john,lastname:doe);editor:apress"`
	}
	// let's get the tag
	titleField, _ := reflect.TypeOf(Book{}).FieldByName("Title")
	metaTag := titleField.Tag.Get("meta")
	// let's create a parser
	parser := tag.NewParser(strings.NewReader(metaTag))
	// do parse
	definitions, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(definitions)
	// Output:
	// [[ Name 'isbn' , Value '11010303030' , Params '[]' ] [ Name 'published' , Value '' , Params '[]' ] [ Name 'author' , Value '' , Params '[{Key:firstname Value:john} {Key:lastname Value:doe}]' ] [ Name 'editor' , Value 'apress' , Params '[]' ]]
}

func TestParser(t *testing.T) {
	testLogger := logger.NewTestLogger(t)

	for _, fixture := range []struct {
		String      string
		Length      int
		Definitions []*tag.Definition
	}{
		{`field:foo;complex_field(name:param,name2:param2,name_3:3);field:1;last_field`,
			4,
			[]*tag.Definition{
				{Name: "field", Value: "foo"},
				{Name: "complex_field", Parameters: []tag.Parameter{{Key: "name", Value: "param"}, {Key: "name2", Value: "param2"}, {Key: "name_3", Value: "3"}}},
				{Name: "field", Value: "1"},
				{Name: "last_field"},
			},
		},
	} {
		parser := tag.NewParser(strings.NewReader(fixture.String))
		parser.SetLogger(testLogger)
		definitions, err := parser.Parse()
		test.Fatal(t, err, nil)
		test.Fatal(t, len(definitions), fixture.Length)
		test.Fatal(t, reflect.DeepEqual(definitions, fixture.Definitions), true)
	}

}
