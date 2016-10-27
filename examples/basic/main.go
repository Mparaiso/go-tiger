package main

import (
	"fmt"
	"net/http"

	"github.com/Mparaiso/tiger-go-framework"
)

// Run this basic example with :
// go run examples/basic/main.go

func main() {
	router := tiger.NewRouter()
	router.Get("/:name", func(container tiger.Container) {
		name := container.GetRequest().URL.Query().Get(":name")
		fmt.Fprintf(container.GetResponseWriter(), "Hello %s ! ", name)
	})
	http.ListenAndServe(":8080", router.Compile())
}
