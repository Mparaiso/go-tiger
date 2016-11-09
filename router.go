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

package tiger

import (
	"net/http"

	"github.com/Mparaiso/go-tiger/matcher"
)

// ContainerFactoryFunc transforms a function into a ContainerFactory
type ContainerFactoryFunc func(http.ResponseWriter, *http.Request) Container

// GetContainer returns the container
func (c ContainerFactoryFunc) GetContainer(w http.ResponseWriter, r *http.Request) Container {
	return c(w, r)
}

// RouterOptions are the optional settings of a router
type RouterOptions struct {
	// ContainerFactory is used to create
	// a container passed to each handler
	// on each request.
	ContainerFactory ContainerFactory

	// UrlVarPrefix is the path variable prefix
	// Route variables in the route pattern
	// Will be available in request.URL.Query() prefixed by
	// UrlVarPrefix, the default prefix is ":"
	// ex : Given the pattern "/resource/:foo"
	//
	// 		foo := request.URL.Query().Get(":foo")
	//
	URLVarPrefix string
}

// ContainerFactory allows providing a custom container to the Router
type ContainerFactory interface {
	GetContainer(http.ResponseWriter, *http.Request) Container
}

// DefaultContainerFactory is the default implementation of ContainerFactory
type DefaultContainerFactory struct{}

// GetContainer returns a new Container
func (d DefaultContainerFactory) GetContainer(w http.ResponseWriter, r *http.Request) Container {
	return &DefaultContainer{ResponseWriter: w, Request: r}
}

// Router handles routing for route handlers
type Router struct {
	*RouteCollection
	*RouterOptions

	matcherProviders matcher.MatcherProviders
}

// NewRouter returns a new router
func NewRouter() *Router {
	return &Router{NewRouteCollection(), &RouterOptions{ContainerFactory: &DefaultContainerFactory{}}, matcher.MatcherProviders{}}
}

// NewRouterWithOptions returns a new router with some options
func NewRouterWithOptions(routerOptions *RouterOptions) *Router {
	return &Router{&RouteCollection{UrlVarPrefix: routerOptions.URLVarPrefix}, routerOptions, matcher.MatcherProviders{}}
}

// SetContainerFactoryFunc sets a function as the container factory
func (r *Router) SetContainerFactoryFunc(factoryFunction func(w http.ResponseWriter, r *http.Request) Container) {
	r.ContainerFactory = ContainerFactoryFunc(factoryFunction)
}

// Compile returns an http.Handler to be use with http.Server
func (r *Router) Compile() http.Handler {
	routes := Routes(r.RouteCollection.Compile())

	return &httpHandler{routes, r.ContainerFactory, routes.GetMetadatas()}
}

type httpHandler struct {
	Routes Routes
	ContainerFactory
	RouteMetadatas RouteMetas
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	container := h.ContainerFactory.GetContainer(w, r)
	container.SetRouteMetadatas(h.RouteMetadatas)
	requestMatcher := matcher.NewRequestMatcher(h.Routes.ToMatcherProviders())
	match := requestMatcher.Match(r)
	if match != nil {
		route := match.(*Route)
		Queue(route.Middlewares).
			Finish(route.Handler).
			Handle(container)
	} else {
		container.Error(StatusError(http.StatusNotFound), http.StatusNotFound)
	}
}

type RouteProvider interface {
	Connect(*RouteCollection)
}

type ContainerDecorator interface {
	Decorate(Container) Container
}
