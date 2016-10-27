package tiger_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	app "github.com/Mparaiso/tiger-go-framework"
)

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
			print("executed 2")
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
