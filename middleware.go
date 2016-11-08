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

	"fmt"

	"github.com/Mparaiso/go-tiger/injector"
	"github.com/Mparaiso/go-tiger/logger"
)

// StatusError is a status error
// it can be used to convert a http status to
// a go error interface
type StatusError int

func (se StatusError) Error() string {
	return http.StatusText(int(se))
}

// Code returns the status code
func (se StatusError) Code() int {
	return int(se)
}

// Container contains server values
type Container interface {
	GetResponseWriter() http.ResponseWriter
	GetRequest() *http.Request
	Error(err error, statusCode int)
	Redirect(url string, statusCode int)
	GetRouteMetadatas() RouteMetas
	SetRouteMetadatas(RouteMetas)
	GetLogger() logger.Logger
	IsDebug() bool
}

// DefaultContainer is the default implementation of the Container
type DefaultContainer struct {
	// ResponseWriter is an http.ResponseWriter
	ResponseWriter http.ResponseWriter
	// Request is an *http.Request
	Request        *http.Request
	RouteMetadatas RouteMetas
	Debug          bool
	Logger         logger.Logger
}

// GetResponseWriter returns a response writer
func (dc DefaultContainer) GetResponseWriter() http.ResponseWriter { return dc.ResponseWriter }

// GetRequest returns a request
func (dc DefaultContainer) GetRequest() *http.Request { return dc.Request }

// GetLogger returns a logger
func (dc *DefaultContainer) GetLogger() logger.Logger {
	if dc.Logger == nil {
		dc.Logger = logger.NewDefaultLogger()
	}
	return dc.Logger
}

// IsDebug returns true if the router is in debug mode
func (dc DefaultContainer) IsDebug() bool { return dc.Debug }

// Error writes an error to the client and logs an error to stdout
func (dc DefaultContainer) Error(err error, statusCode int) {
	if dc.IsDebug() {
		http.Error(dc.GetResponseWriter(), err.Error(), statusCode)

	} else {
		dc.GetLogger().LogF(logger.Error, "%s", err)
		http.Error(dc.GetResponseWriter(), StatusError(statusCode).Error(), statusCode)
	}
}

// Redirect replies with a redirection
func (dc DefaultContainer) Redirect(url string, statusCode int) {
	http.Redirect(dc.GetResponseWriter(), dc.GetRequest(), url, statusCode)
}

// SetRouteMetadatas sets the route metadatas
// They can be used for the creation of URL from a route name
func (dc *DefaultContainer) SetRouteMetadatas(metadatas RouteMetas) {
	dc.RouteMetadatas = metadatas
}

// GetRouteMetadatas returns RouteMetas
func (dc DefaultContainer) GetRouteMetadatas() RouteMetas {
	return dc.RouteMetadatas
}

// Handler is a controller that takes a context
type Handler func(Container)

// Wrap wraps Route.Handler with each middleware and returns a new Route
func (h Handler) Wrap(middlewares ...Middleware) Handler {
	return Queue(middlewares).Finish(h)
}

// Handle handles a request
func (h Handler) Handle(c Container) {
	h(c)
}

// ToHandlerFunc converts Handler to http.Handler
func (h Handler) ToHandlerFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := &DefaultContainer{ResponseWriter: w, Request: r}
		h(c)
	}
}

// ToHandler converts an idiomatic http.HandlerFunc to a Handler
func ToHandler(handlerFunc func(http.ResponseWriter, *http.Request)) func(Container) {
	return func(c Container) {
		handlerFunc(c.GetResponseWriter(), c.GetRequest())
	}
}

// ToMiddleware wraps a classic net/http middleware (func(http.HandlerFunc) http.HandlerFunc)
// into a Middleware compatible with this package
func ToMiddleware(classicMiddleware func(http.HandlerFunc) http.HandlerFunc) Middleware {
	return func(c Container, next Handler) {
		classicMiddleware(func(w http.ResponseWriter, r *http.Request) {

			next(c)
		})(c.GetResponseWriter(), c.GetRequest())
	}
}

//Queue is a reusable queue of middlewares
type Queue []Middleware

// Then returns a new middleware with middleware argument queued after
// the current middleware
func (q Queue) Then(middleware Middleware) Middleware {
	var current Middleware
	for _, middleware := range q {
		if current == nil {
			current = middleware
		} else {
			current = current.Then(middleware)
		}
	}
	return current
}

// Finish returns a new queue of middlewares
func (q Queue) Finish(h Handler) Handler {
	var current Middleware
	if len(q) == 0 {
		return h
	}
	for _, middleware := range q {
		if current == nil {
			current = middleware
		} else {
			current = current.Then(middleware)
		}
	}
	return current.Finish(h)
}

// Middleware is a function that takes an Handler and returns a new Handler
type Middleware func(container Container, next Handler)

// Then allows to chain middlewares by returning a
// new middleware wrapped by the previous middleware in the chain
func (m Middleware) Then(middleware Middleware) Middleware {
	return func(c Container, next Handler) {
		m(c, func(c Container) {
			middleware(c, next)
		})
	}
}

// Finish returns a handler that is preceded by the middleware
func (m Middleware) Finish(h Handler) Handler {
	return func(c Container) {
		m(c, h)
	}
}

// Inject enable dependency injection and returns an handler
func Inject(function injector.Function) Handler {
	return func(c Container) {
		if _, ok := c.(InjectorProvider); !ok {
			c.Error(fmt.Errorf("container %s doesn't implement InjectorProvider", c), http.StatusInternalServerError)
			return
		}
		err := c.(InjectorProvider).GetInjector().Call(function)
		if err != nil {
			c.Error(err, http.StatusInternalServerError)
		}
	}
}

type InjectorProvider interface {
	GetInjector() *injector.Injector
}

// ContainerWithInjector is a container with a dependency injector
type ContainerWithInjector struct {
	Container
	injector *injector.Injector
}

// GetInjector returns an injector
func (container *ContainerWithInjector) GetInjector() *injector.Injector {
	if container.injector == nil {
		container.injector = injector.NewInjector()
		container.injector.RegisterValue(container.GetRequest())
		container.injector.RegisterValue(container.GetResponseWriter())
	}
	return container.injector
}
