//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//      http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package tiger_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Mparaiso/go-tiger"
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
		Handle(&tiger.DefaultContainer{ResponseWriter: response, Request: request})

	fmt.Println(response.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println(response.Body.String())

	// Output:
	// *
	// done

}
