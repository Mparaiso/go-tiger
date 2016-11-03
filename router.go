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
	"sync"

	"github.com/Mparaiso/go-tiger/matcher"
)

type ContainerFactoryFunc func(http.ResponseWriter, *http.Request) Container

func (c ContainerFactoryFunc) GetContainer(w http.ResponseWriter, r *http.Request) Container {
	return c(w, r)
}

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
	UrlVarPrefix string
}

type ContainerFactory interface {
	GetContainer(http.ResponseWriter, *http.Request) Container
}

type DefaultContainerFactory struct{}

func (d DefaultContainerFactory) GetContainer(w http.ResponseWriter, r *http.Request) Container {
	return DefaultContainer{w, r}
}

type Router struct {
	*RouteCollection
	*RouterOptions
	*sync.Once

	matcherProviders matcher.MatcherProviders
}

func NewRouter() *Router {
	return &Router{NewRouteCollection(), &RouterOptions{ContainerFactory: &DefaultContainerFactory{}}, new(sync.Once), matcher.MatcherProviders{}}
}

func NewRouterWithOptions(routerOptions *RouterOptions) *Router {
	return &Router{&RouteCollection{UrlVarPrefix: routerOptions.UrlVarPrefix}, routerOptions, new(sync.Once), matcher.MatcherProviders{}}
}

// Compile returns an http.Handler to be use with http.Server
func (r *Router) Compile() http.Handler {
	routes := Routes(r.RouteCollection.Compile()).ToMatcherProviders()

	return &httpHandler{routes, r.ContainerFactory}
}

type httpHandler struct {
	MatcherProviders []matcher.MatcherProvider
	ContainerFactory
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	container := h.ContainerFactory.GetContainer(w, r)
	requestMatcher := matcher.NewRequestMatcher(h.MatcherProviders)
	match := requestMatcher.Match(r)
	if match != nil {
		route := match.(*Route)
		Queue(route.Middlewares).Finish(route.Handler).Handle(container)
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
