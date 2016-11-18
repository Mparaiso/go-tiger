TIGER
------

[![GoDoc](https://godoc.org/github.com/Mparaiso/go-tiger?status.png)](https://godoc.org/github.com/Mparaiso/go-tiger)

[![Build Status](https://travis-ci.org/Mparaiso/go-tiger.svg?branch=master)](https://travis-ci.org/Mparaiso/go-tiger)


[![codebeat badge](https://codebeat.co/badges/bff186bc-1b39-4d22-9c07-159844cc1c87)](https://codebeat.co/projects/github-com-mparaiso-go-tiger)

author: mparaiso <mparaiso@online.fr>

license: Apache 2-0

copyright 2016

tiger is a minimal request router written in Go. 

It gives developers just enough to handle routing, thanks to regular expressions and 
a system of matchers that can be customized. It is easier to use and to learn than any 
other router package in Go.

tiger also contains a toolkit providing various utilities to make creating web apps a breeze.


#### Installation

    go get github.com/mparaiso/go-tiger

#### Router Features

    - [x] Path Variables
    - [x] Middleware queue
    - [x] Sub routes
    - [x] Modular architecture
    - [x] Custom HTTP verbs
    - [x] support for idiomatic http.HandlerFunc as handlers
    - [x] support idiomatic Go middlewares with the following signature : func(http.HandlerFunc)http.HandlerFunc

#### Tiger toolkit Features

    - [x] Access control lists
    - [x] generic datastructures
        - [x] Ordered map
    - [x] Databases :
        - [x] Mysql support
        - [x] SQlite support
        - [x] PostgreSQL support
        - [x] Query builder
        - [x] DB row to struct value mapper
    - [ ] ORM (WIP)
    - [x] Dependency injection (completely optional thus doesn't impact the performances of the router!)
    - [x] Optional Dependendy injection
    - [x] Signals
    - [x] Validation
    - [x] logging
    - [x] light weight test helpers

#### Basic Usage

```go
    package main

    import (
        "fmt"
        "net/http"

        tiger "github.com/Mparaiso/go-tiger"
    )


    func main() {
        // Create a router
        router := tiger.NewRouter()
        // Use an idiomatic http.Handlerfunc as the app index
        router.Get("/", tiger.FromHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprint(w,"Hi from idiomatic router")
        }))
        // Use a tiger.Handler to read url variables
        router.Get("/greetings/:name", func(container tiger.Container) {
            // container is just an interface that holds both the Request and the ResponseWriter 
            // an interface allows the developer to customize the container without changing the signature 
            // of the handler function.

            // URL variables are merged with the query string, however the prefix 
            // can be modified to avoid collisions
            name := container.GetRequest().URL.Query().Get(":name")
            fmt.Fprintf(container.GetResponseWriter(), "Hello %s ! ", name)
            // using the query string to hold route variables allows any handler of any type and shape 
            // to handle route variables.
        })
        // Compile and use with http.Server
        http.ListenAndServe("127.0.0.1:8080", router.Compile())
    }
```
