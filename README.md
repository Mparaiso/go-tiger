TIGER
------

[![GoDoc](https://godoc.org/github.com/Mparaiso/go-tiger?status.png)](https://godoc.org/github.com/Mparaiso/go-tiger)
[![Build Status](https://travis-ci.org/Mparaiso/go-tiger.svg?branch=master)](https://travis-ci.org/Mparaiso/go-tiger)
[![Code Climate](https://codeclimate.com/github/Mparaiso/go-tiger/badges/gpa.svg)](https://codeclimate.com/github/Mparaiso/go-tiger)
[![Test Coverage](https://codeclimate.com/github/Mparaiso/go-tiger/badges/coverage.svg)](https://codeclimate.com/github/Mparaiso/go-tiger/coverage)
[![Issue Count](https://codeclimate.com/github/Mparaiso/go-tiger/badges/issue_count.svg)](https://codeclimate.com/github/Mparaiso/go-tiger)
[![codebeat badge](https://codebeat.co/badges/bff186bc-1b39-4d22-9c07-159844cc1c87)](https://codebeat.co/projects/github-com-mparaiso-go-tiger)

author: mparaiso <mparaiso@online.fr>

license: Apache 2-0

copyright 2016

Package tiger is a toolkit that aims at making go web development, web security,
and data modelling easy and simple.


#### Installation

    go get github.com/mparaiso/go-tiger

#### Router Features

   
#### Tiger toolkit Features

	
	- [x] web: micro framework
		- [x] Path Variables
	    - [x] Middleware queue
	    - [x] Sub routes
	    - [x] Modular architecture
	    - [x] Custom HTTP verbs
	    - [x] support for idiomatic http.HandlerFunc as handlers
	    - [x] support idiomatic Go middlewares with the following signature : func(http.HandlerFunc)http.HandlerFunc

    - [x] acl: Access control lists
	
    - [x] container: generic datastructures
        - [x] Ordered map
    
	- [x] db: Database tools:
		- [x] Query builder:
	        - [x] Mysql support
	        - [x] SQlite support
	        - [x] PostgreSQL support
        - [x] DB row to struct value mapper
		
	- [x] mongo: MongoDB Object document mapper
		
	- [x] funcs: Functional programming helpers
		- [x] map
		- [x] reduce
		- [x] filter
		- [x] find
		- [x] groupBy
		- [x] keyBy
		
    - [x] injector: Dependency injection 
    - [x] signal: Signals (event listeners)
    - [x] validator: Validation
    - [x] logger: Logging
    - [x] test: Test helpers
	- [x] tag: struct tag parser

#### Basic Usage

```go
    package main

    import (
        "fmt"
        "net/http"

        tiger "github.com/Mparaiso/go-tiger/web"
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
            // using an interface allows the developer to customize the container 
			// without changing the signature of the handler function.

            // URL variables are merged with the query string, however the prefix 
            // can be modified to avoid collisions
            name := container.GetRequest().URL.Query().Get(":name")
            fmt.Fprintf(container.GetResponseWriter(), "Hello %s ! ", name)
            // using the query string to hold route variables also 
			// allows any handler of any type and shape 
            // to handle route variables.
        })
        // create an http.Handler use it with http.Server
        http.ListenAndServe("127.0.0.1:8080", router.Compile())
    }
```
