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
	"path"
	"regexp"
	"strings"

	"github.com/Mparaiso/go-tiger/matcher"
)

type RouteCollection struct {
	Prefix                string
	UrlVarPrefix          string
	matchers              []matcher.Matcher
	childRouteCollections []*RouteCollection
	routes                []*Route
	middlewares           []Middleware
}

func NewRouteCollection() *RouteCollection {
	return &RouteCollection{matchers: []matcher.Matcher{}, childRouteCollections: []*RouteCollection{}, routes: []*Route{}, middlewares: []Middleware{}}
}

func (r *RouteCollection) AddRequestMaster(matchers ...matcher.Matcher) *RouteCollection {
	r.matchers = append(r.matchers, matchers...)
	return r
}

func (r *RouteCollection) Use(middlewares ...Middleware) *RouteCollection {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

func (r *RouteCollection) Get(pattern string, handler Handler) *Route {
	return r.Match([]string{"GET"}, pattern, handler)
}

func (r *RouteCollection) Post(pattern string, handler Handler) *Route {
	return r.Match([]string{"POST"}, pattern, handler)
}

func (r *RouteCollection) Put(pattern string, handler Handler) *Route {
	return r.Match([]string{"PUT"}, pattern, handler)
}

func (r *RouteCollection) Patch(pattern string, handler Handler) *Route {
	return r.Match([]string{"PATCH"}, pattern, handler)
}

func (r *RouteCollection) Delete(pattern string, handler Handler) *Route {
	return r.Match([]string{"DELETE"}, pattern, handler)
}

func (r *RouteCollection) Options(pattern string, handler Handler) *Route {
	return r.Match([]string{"OPTIONS"}, pattern, handler)
}

func (r *RouteCollection) Match(methods []string, pattern string, handler Handler) *Route {
	route := &Route{
		Handler:     handler,
		Matchers:    []matcher.Matcher{},
		Middlewares: []Middleware{},
		Meta:        &RouteMeta{Pattern: pattern, Methods: methods},
	}
	r.routes = append(r.routes, route)
	return route
}

func (r *RouteCollection) compileRoute(route *Route) *Route {
	compiledRoute := &Route{
		Handler:     route.Handler,
		Matchers:    route.Matchers,
		Middlewares: route.Middlewares,
		Meta: &RouteMeta{Name: route.GetMeta().Name,
			Pattern:      route.GetMeta().Pattern,
			Prefix:       r.Prefix,
			Methods:      route.GetMeta().Methods,
			URLVARPrefix: r.UrlVarPrefix},
	}
	compiledRoute.Matchers = append(
		append(
			matcher.Matchers{},
			matcher.Pattern(compiledRoute.GetMeta().Pattern, r.Prefix, r.UrlVarPrefix),
			matcher.Method(route.GetMeta().Methods...),
		),
		compiledRoute.Matchers...)
	if route.GetMeta().Name == "" {
		route.Meta.Name = strings.Trim(regexp.MustCompile(`[^a-z A-Z 0-9]`).ReplaceAllString(route.GetMeta().GetPath(), "_"), "_")
		if route.Meta.Name == "" {
			route.Meta.Name = "_"
		}
	}
	compiledRoute.Matchers = append(append([]matcher.Matcher{}, r.matchers...), compiledRoute.Matchers...)
	compiledRoute.Middlewares = append(append([]Middleware{}, r.middlewares...), compiledRoute.Middlewares...)
	return compiledRoute

}

// Compile add the route collection specific configuration to each route in the colelction
// and returns the collection of compiled routes
func (r *RouteCollection) Compile() []*Route {
	routes := []*Route{}

	for _, routeCollection := range r.childRouteCollections {
		routeCollection.middlewares = append(append([]Middleware{}, r.middlewares...), routeCollection.middlewares...)

		routeCollection.matchers = append(append([]matcher.Matcher{}, r.matchers...), routeCollection.matchers...)
		routes = append(routes, routeCollection.Compile()...)
	}

	for _, route := range r.routes {
		compiledRoute := r.compileRoute(route)
		routes = append(routes, compiledRoute)
	}

	return routes
}

func (r *RouteCollection) Mount(prefix string, provider RouteProvider) *RouteCollection {
	routeCollection := r.Sub(prefix)
	provider.Connect(routeCollection)
	return r
}

func (r *RouteCollection) Sub(prefix string) *RouteCollection {
	childrouteCollection := &RouteCollection{
		Prefix: path.Join(r.Prefix, prefix),
	}
	r.childRouteCollections = append(r.childRouteCollections, childrouteCollection)
	return childrouteCollection
}
