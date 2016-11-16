package container_test

import (
	"fmt"

	"github.com/Mparaiso/go-tiger/container"
)

func ExampleOrderedMap() {
	/*
	   The map is ordered so the sequence of results
	   when iterating over that map is guaranteed to be
	   deterministic, unlike the default Go map
	*/
	Map := container.NewOrderedMap()
	Map.Set("a", 1)
	Map.Set("b", 2)
	Map.Set("d", 3)
	Map.Set("e", 4)
	Map.Set("a", 10)
	Map.Delete("b")
	// iterate through the map with an index
	for i := 0; i < Map.Length(); i++ {
		fmt.Printf("%v %v\n", Map.KeyAt(i), Map.ValueAt(i))
	}
	// iterates through with an array of keys
	for _, key := range Map.Keys() {
		fmt.Printf("%v %v\n", key, Map.Get(key))
	}
	// Output:
	// a 10
	// d 3
	// e 4
	// a 10
	// d 3
	// e 4

}
