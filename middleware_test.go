package tiger_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	tiger "github.com/mparaiso/tiger-go-framework"
)

func ExampleMiddleware_Then() {

	// Let's chain middlewares thanks to the Then method

	middleware1, middleware2, middleware3 := func(c tiger.Container, next tiger.Handler) {
		fmt.Print(1)
		next(c)
	},
		func(c tiger.Container, next tiger.Handler) {
			fmt.Print(2)
			next(c)
		},
		func(c tiger.Container, next tiger.Handler) {
			fmt.Print(3)
			next(c)
		}

	tiger.Middleware(middleware1).
		Then(middleware2).
		Then(middleware3).
		Finish(func(tiger.Container) { fmt.Println("Handle the request") }).
		Handle(nil)

	// Output:
	// 123Handle the request
}

func ExampleMiddleware_Queue() {
	tiger.Queue([]tiger.Middleware{
		func(c tiger.Container, next tiger.Handler) {
			fmt.Print(1)
			next(c)
		},
		func(c tiger.Container, next tiger.Handler) {
			fmt.Print(2)
			next(c)
		},
		func(c tiger.Container, next tiger.Handler) {
			fmt.Print(3)
			next(c)
		},
	}).Finish(func(c tiger.Container) {
		fmt.Print("Finish")
	}).Handle(nil)

	// Output:
	// 123Finish
}

func ExampleHandler_Wrap() {
	tiger.Handler(func(c tiger.Container) {
		fmt.Print("Done")
	}).Wrap(func(c tiger.Container, next tiger.Handler) { fmt.Print(1); next(c) }, func(c tiger.Container, next tiger.Handler) { fmt.Print(2); next(c) }).
		Handle(nil)
	// Output:
	// 12Done
}

func ExampleToMiddleware() {
	// Let's convert a classic http middleware into a middleware supported by this package

	classicCORSMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			// this middleware handles corss origin requests from browsers
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(rw, r)
		}
	}

	convertedMiddleware := tiger.ToMiddleware(classicCORSMiddleware)

	// Let's test our converted middleware
	request, _ := http.NewRequest("GET", "https://acme.com", nil)
	response := httptest.NewRecorder()

	convertedMiddleware.
		Finish(func(c tiger.Container) { c.GetResponseWriter().Write([]byte("done")) }).
		Handle(&tiger.DefaultContainer{response, request})

	fmt.Println(response.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println(response.Body.String())

	// Output:
	// *
	// done

}
