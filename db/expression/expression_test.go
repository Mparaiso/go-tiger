package expression_test

import (
	"fmt"

	. "github.com/Mparaiso/go-tiger/db/expression"
)

func ExampleExpression() {
	exp := And(Like("Name", "%john doe%"), Eq("Age", 34))
	fmt.Println(exp)
	exp = And(Or(Eq("Name", "'john doe'"), Eq("Name", "'jane doe'")), Gt("Age", 18), NotIn("Location", "'Canada'", "'USA'", "'France'"))
	fmt.Println(exp)
	exp = Or(IsNull("CategoryId"), NotIn("Title", "'Book A'", "'Book B'"))
	fmt.Println(exp)
	// Output:
	// (Name LIKE %john doe%) AND (Age = 34)
	// ((Name = 'john doe') OR (Name = 'jane doe')) AND (Age > 18) AND (Location NOT IN ( 'Canada' , 'USA' , 'France' ))
	// (CategoryId IS NULL) OR (Title NOT IN ( 'Book A' , 'Book B' ))
}
