TIGER
------

author: mparaiso <mparaiso@online.fr>

license: GPL-3.0

copyright 2016

tiger is a minimal request router written in Go. 
It gives developers just enough to handle routing, thanks to regular expressions and 
a system of matchers that can be customized. It is easier to use and to learn than any 
other router package in Go.

#### Installation

    go get github.com/mparaiso/tiger-go-framework

#### Basic Usage

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