TIGER
------

[![GoDoc](https://godoc.org/github.com/Mparaiso/go-tiger?status.png)](https://godoc.org/github.com/Mparaiso/go-tiger)

[![Build Status](https://travis-ci.org/Mparaiso/go-tiger.svg?branch=master)](https://travis-ci.org/Mparaiso/go-tiger)

author: mparaiso <mparaiso@online.fr>

license: Apache 2-0

copyright 2016

tiger is a minimal request router written in Go. 
It gives developers just enough to handle routing, thanks to regular expressions and 
a system of matchers that can be customized. It is easier to use and to learn than any 
other router package in Go.

#### Installation

    go get github.com/mparaiso/go-tiger

#### Features

    - [x] Path Variables
    - [x] Middleware queue
    - [x] Dependency injection (completely optional thus doesn't impact the framework performances!)
    - [x] Sub routes
    - [x] Modular architecture
    - [x] Custom HTTP verbs
    - [x] idiomatic http.HandlerFunc as handlers
    - [x] idiomatic Go middlewares ( with the following signature : func(http.HandlerFunc)http.HandlerFunc )

#### Basic Usage

```go
    package main

    import (
        "fmt"
        "log"
        "net/http"

        tiger "github.com/Mparaiso/go-tiger"
    )

    // Run this basic example with :
    // go run examples/basic/main.go

    func main() {
        addr := "127.0.0.1:8080"
        // Create a router
        router := tiger.NewRouter()
        // Use a tiger.Handler to read url variables
        router.Get("/greetings/:name", func(container tiger.Container) {
            // URL variables are merged with the Query, however the prefix 
            // can be modified to avoid collisions
            name := container.GetRequest().URL.Query().Get(":name")
            fmt.Fprintf(container.GetResponseWriter(), "Hello %s ! ", name)
        })
        // Use an idiomatic http.Handlerfunc as the app index
        router.Get("/", tiger.FromHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprint(w, `
            <html>
                <body>
                    <h3>Index</h3>
                    <p>Welcome to tiger !
                    <p><a href="/secure">Login</a>
                </body>
            </html>
            `)
        }))

        // Compile and use with http.Server
        log.Printf("Listening on %s \n", addr)
        log.Fatal(http.ListenAndServe(addr, router.Compile()))
    }
```
