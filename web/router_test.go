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
package web_test

import (
	"fmt"
	app "github.com/Mparaiso/go-tiger/web"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkRouter_ServeHTTP(b *testing.B) {
	router := app.NewRouter()
	for i := 0; i < 1000; i++ {
		router.Get(fmt.Sprint("/", i), func(c app.Container) {})
	}
	router.Get("/foobar", func(c app.Container) {
		c.GetResponseWriter().Write([]byte("Hello World"))
	})
	handler := router.Compile()
	request, _ := http.NewRequest("GET", "https://example.com/foobar", nil)
	response := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(response, request)
	}
}

func BenchmarkDefaultServeMux_ServeHTTP(b *testing.B) {
	handler := http.NewServeMux()
	request, _ := http.NewRequest("GET", "https://example.com/foobar", nil)
	response := httptest.NewRecorder()
	for i := 0; i < 1000; i++ {
		handler.HandleFunc(fmt.Sprint("/", i), func(rw http.ResponseWriter, r *http.Request) {})
	}
	handler.HandleFunc("/foobar", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Hello World"))
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(response, request)
	}
}

func ExampleRouter() {
	router := app.NewRouter()
	router.Use(func(container app.Container, next app.Handler) {
		container.GetResponseWriter().Header().Add("X-Special", "Yes")
		next(container)
	})
	router.Get("/greetings/:firstname/:lastname", func(container app.Container) {
		fmt.Fprintf(container.GetResponseWriter(), "Hello %s %s !",
			container.GetRequest().URL.Query().Get(":firstname"),
			container.GetRequest().URL.Query().Get(":lastname"),
		)
	})
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com/greetings/John-Rodger/Doe", nil)
	router.Compile().ServeHTTP(response, request)
	fmt.Println(response.Body.String())
	fmt.Println(response.Header().Get("X-Special"))
	// Output:
	// Hello John-Rodger Doe !
	// Yes
}

func ExampleRouter_Sub() {
	// Router.Sub allows router inheritance
	// SubRouters can be created to allow a custom middleware queue
	// executed by "sub" handlers
	router := app.NewRouter()
	router.
		Use(func(c app.Container, next app.Handler) {
			c.GetResponseWriter().Header().Set("X-Root", "Yes")
			next(c)
		}).
		Sub("/sub/").
		Use(func(c app.Container, next app.Handler) {
			// Will only get executed by by handlers defined
			// in that sub router
			c.GetResponseWriter().Header().Set("X-Sub", "Yes")
			next(c)
		}).
		Get("/", func(c app.Container) {
			fmt.Fprint(c.GetResponseWriter(), "Sub")
		})
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com/sub/", nil)
	router.Compile().ServeHTTP(response, request)
	fmt.Println(response.Header().Get("X-Root"))
	fmt.Println(response.Header().Get("X-Sub"))
	// Output:
	// Yes
	// Yes
}

func ExampleRouter_Mount() {
	// Mount allows to define modules
	// That can be reused in different application.
	// NewTestModule returns a *TestModule
	// that has a method with the following signature:
	//
	//		func(module TestModule)Connect(collection *app.RouteCollection)
	//
	router := app.NewRouter()
	router.Mount("/mount", NewTestModule())
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com/mount", nil)
	router.Compile().ServeHTTP(response, request)
	fmt.Println("Status", response.Code)
	// Output:
	// Status 200
}

type ContainerDecorator struct {
	app.Container
}

func Decorate(container app.Container) app.Container {
	return ContainerDecorator{container}
}

type TestModule struct{}

func NewTestModule() *TestModule {
	return &TestModule{}
}

func (module TestModule) Connect(collection *app.RouteCollection) {
	collection.
		Use(func(container app.Container, next app.Handler) {
			next(Decorate(container))
		}).
		Get("/", func(c app.Container) {
			_, ok := c.(ContainerDecorator)
			if !ok {
				c.Error(fmt.Errorf("Error container is not a ContainerDecorator"), 500)
				return
			}
			c.GetResponseWriter().WriteHeader(200)
		})

}
