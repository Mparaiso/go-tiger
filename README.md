TIGER
------

[![GoDoc](https://godoc.org/github.com/Mparaiso/tiger-go-framework?status.png)](https://godoc.org/github.com/Mparaiso/tiger-go-framework)

[![Build Status](https://travis-ci.org/Mparaiso/tiger-go-framework.svg?branch=master)](https://travis-ci.org/Mparaiso/tiger-go-framework)

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

```go
    package main

    import (
        "fmt"
        "log"
        "net/http"

        "github.com/Mparaiso/tiger-go-framework"
    )

    // Run this basic example with :
    // go run examples/basic/main.go

    func main() {
        addr := "127.0.0.1:8080"
        // Create a router
        router := tiger.NewRouter()
        // Use an idiomatic http middleware to log requests
        router.Use(tiger.ToMiddleware(func(next http.HandlerFunc) http.HandlerFunc {
            return func(w http.ResponseWriter, r *http.Request) {
                next(w, r)
                log.Printf("%s %s", r.Method, r.URL.RequestURI())
            }
        }))
        // Use an idiomatic http.Handlerfunc as the app index
        router.Get("/", tiger.ToHandler(func(w http.ResponseWriter, r *http.Request) {
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
        // Use a tiger.Handler to read url variables
        router.Get("/greetings/:name", func(container tiger.Container) {
            name := container.GetRequest().URL.Query().Get(":name")
            fmt.Fprintf(container.GetResponseWriter(), "Hello %s ! ", name)
        })
        // Create a subrouter
        secureRouter := router.Sub("/secure")
        // Basic security middleware that will be executed
        // on each request matching that subroute
        secureRouter.Use(func(c tiger.Container, next tiger.Handler) {
            // use the default implementation of tiger.Container injected in the router by default
            if login, password, ok := c.GetRequest().BasicAuth(); ok {
                if login == "login" && password == "password" {
                    next(c)
                    return
                }
            }
            // or return a 401 status code
            c.GetResponseWriter().Header().Set("WWW-Authenticate", `Basic realm="Secure"`)
            c.Error(tiger.StatusError(http.StatusUnauthorized), http.StatusUnauthorized)
        })

        secureRouter.Get("/", func(c tiger.Container) {
            fmt.Fprintf(c.GetResponseWriter(), `<body>
                <p>You are in the secure zone ! congrats!</p>
            </body>`)
        })
        // Compile and use with http.Server
        log.Printf("Listening on %s \n", addr)
        log.Fatal(http.ListenAndServe(addr, router.Compile()))
    }
```