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

package httpx

import (
	"log"
	"net/http"
	"path"
	"strings"

	"context"
)

type key byte

const (
	CurrentRoute key = iota
	RequestID
)

type (
	// RouteOption is a route option
	RouteOption interface {
		handle(*Route)
	}
	applicationOption interface {
		option(*Application)
	}
	// Filter is an application filter.
	// filters are called before the route handler
	Filter func(http.HandlerFunc) http.HandlerFunc
)

// Grouper is a logical group of routes
type Grouper struct {
	prefix  string
	filters []Filter
	routes  []Route
	groups  []Grouper
}

// AddFilter add a filter to the group. Filters are called
// before the request handler, in order of registration
func (g *Grouper) AddFilter(filters ...Filter) *Grouper {
	g.filters = append(g.filters, filters...)
	return g
}

// HandleResource takes a struct and register its methods as route handlers
// according to these interfaces
//
// List(w http.ResponseWriter,r *http.Request)
// Get(w http.ResponseWriter,r *http.Request)
// Post(w http.ResponseWriter,r *http.Request)
// Put(w http.ResponseWriter,r *http.Request)
// Patch(w http.ResponseWriter,r *http.Request)
// Delete(w http.ResponseWriter,r *http.Request)
// Options(w http.ResponseWriter,r *http.Request)
//
// Routes are registered according to this pattern:
// resource.HandleFunc("/:{resource}_id",method,WithMethod({method}),WithName({method}-{resource}))
//
// The remaining options will be applied to all routes
//
// If the struct implements interface{Filter(http.HandlerFunc)http.HandlerFunc} ,
// The filter will be call before any route handler
func (g *Grouper) HandleResource(resourceName string, controller interface{}, options ...RouteOption) *Grouper {
	resource := g.Group("/")
	if filter, ok := controller.(interface {
		Filter(http.HandlerFunc) http.HandlerFunc
	}); ok {
		resource.AddFilter(filter.Filter)
	}
	resource_id := "/:" + resourceName + "_id"
	if handler, ok := controller.(interface {
		List(http.ResponseWriter, *http.Request)
	}); ok {
		resource.HandleFunc("/", handler.List, append(options, WithMethod("GET"), WithName("list-"+resourceName))...)
	}
	if handler, ok := controller.(interface {
		Post(http.ResponseWriter, *http.Request)
	}); ok {
		resource.HandleFunc("/", handler.Post, append(options, WithMethod("POST"), WithName("post-"+resourceName))...)
	}
	if handler, ok := controller.(interface {
		Get(http.ResponseWriter, *http.Request)
	}); ok {
		resource.HandleFunc(resource_id, handler.Get, append(options, WithMethod("GET"), WithName("get-"+resourceName))...)
	}

	if handler, ok := controller.(interface {
		Put(http.ResponseWriter, *http.Request)
	}); ok {
		resource.HandleFunc(resource_id, handler.Put, append(options, WithMethod("PUT"), WithName("put-"+resourceName))...)
	}
	if handler, ok := controller.(interface {
		Patch(http.ResponseWriter, *http.Request)
	}); ok {
		resource.HandleFunc(resource_id, handler.Patch, append(options, WithMethod("PATCH"), WithName("patch-"+resourceName))...)
	}
	if handler, ok := controller.(interface {
		Delete(http.ResponseWriter, *http.Request)
	}); ok {
		resource.HandleFunc(resource_id, handler.Delete, append(options, WithMethod("DELETE"), WithName("delete-"+resourceName))...)
	}
	if handler, ok := controller.(interface {
		Options(http.ResponseWriter, *http.Request)
	}); ok {
		resource.HandleFunc(resource_id, handler.Options, append(options, WithMethod("OPTIONS"), WithName("options-"+resourceName))...)
	}
	return g
}

// HandleFunc registers a new route in the router
func (g *Grouper) HandleFunc(pattern string, handlerFunc http.HandlerFunc, options ...RouteOption) *Grouper {
	g.Handle(pattern, handlerFunc, options...)
	return g
}

// Handle registers a new route in the router
func (g *Grouper) Handle(pattern string, handler http.Handler, options ...RouteOption) *Grouper {
	route := Route{path: path.Join(g.prefix, pattern), handler: handler, methods: map[string]bool{}}
	for _, option := range options {
		option.handle(&route)
	}
	g.routes = append(g.routes, route)
	return g
}

// Group add a new subgroup to the current group
func (g *Grouper) Group(prefix string) *Grouper {
	var filters []Filter
	filters = append(filters, g.filters...)
	group := Grouper{prefix: path.Join(g.prefix, prefix), filters: filters}
	g.groups = append(g.groups, group)
	return &g.groups[len(g.groups)-1]
}

// Routes returns all routes in the group and its sub groups
func (g Grouper) Routes() []Route {
	var routes []Route
	routes = append(routes, g.routes...)
	for i := range routes {
		routes[i].filters = append(routes[i].filters, g.filters...)
	}
	if len(g.groups) > 0 {
		for _, group := range g.groups {
			routes = append(routes, group.Routes()...)
		}
	}
	return routes
}

// Application is a http router
type Application struct {
	*Grouper
}

// NewApplication is a constructor
func NewApplication() *Application {
	return &Application{Grouper: &Grouper{prefix: "/"}}
}

func (a Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		route Route
		err   error
	)
	defer func() {
		if e := recover(); e != nil {
			log.Println(r.Context().Value(RequestID).(string), "\t", e)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()

	if route, err = resolveRoute(a.Routes(), r); err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	handler := route.handler.ServeHTTP
	for i := len(route.filters) - 1; i >= 0; i-- {
		handler = route.filters[i](handler)
	}
	WithValue(r, map[interface{}]interface{}{CurrentRoute: route, RequestID: uuid()})
	handler(w, r)

}

// ResolveRoute match a request with a series of routes and returns
// the matched route or an error if not match found
func resolveRoute(routes []Route, request *http.Request) (Route, error) {
	for _, r := range routes {
		if !r.HasMethod(request.Method) {
			continue
		}
		if variables, match := r.Match(request.URL.Path); match {
			query := request.URL.Query()
			for key, value := range variables {
				query.Add(key, value)
			}
			request.URL.RawQuery = query.Encode()
			return r, nil
		}
	}
	return Route{}, errRouteNotFound
}

type optionPrototype struct {
	handler func(*Route)
}

func (option optionPrototype) handle(r *Route) {
	if option.handler != nil {
		option.handler(r)
	}
}

// WithName names the route
func WithName(name string) RouteOption {
	return optionPrototype{handler: func(r *Route) {
		r.name = name
	}}
}

// WithMethod add one of more methods to a route
func WithMethod(methods ...string) RouteOption {
	return optionPrototype{handler: func(r *Route) {
		for _, method := range methods {
			r.methods[strings.ToUpper(method)] = true
		}
		if _, ok := r.methods["GET"]; ok {
			r.methods["HEAD"] = true
		}
	}}
}

func WithValue(request *http.Request, values map[interface{}]interface{}) {
	for key, value := range values {
		*request = *(request.WithContext(context.WithValue(request.Context(), key, value)))
	}
}
