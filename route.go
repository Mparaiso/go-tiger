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
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"path"

	"github.com/Mparaiso/go-tiger/matcher"
)

// Route is an app route
type Route struct {
	Handler     func(Container)
	Matchers    []matcher.Matcher
	Middlewares []Middleware
	Meta        *RouteMeta
}

// GetMeta gets route metadatas
func (r *Route) GetMeta() *RouteMeta {
	if r.Meta == nil {
		r.Meta = &RouteMeta{}
	}
	return r.Meta
}

// SetName sets the route metadata's name
func (r *Route) SetName(name string) *Route {
	r.GetMeta().Name = name
	return r
}

// GetName returns the route metadata name
func (r *Route) GetName() string {
	return r.GetMeta().Name
}

// Use adds middlewares to the route
func (r *Route) Use(middleware ...Middleware) *Route {
	r.Middlewares = append(r.Middlewares, middleware...)
	return r
}

// Match adds matchers to the route
func (r *Route) Match(matchers ...matcher.Matcher) *Route {
	r.Matchers = append(r.Matchers, matchers...)
	return r
}

// String returns a string representation of the route
func (r Route) String() string {
	return fmt.Sprintf("Route{Name:%s,Path:%s,Methods:%+v,Middlewares:%+v}",
		r.GetName(), r.GetMeta().GetPath(), r.GetMeta().Methods, r.Middlewares,
	)
}

// GetMatchers return the request matchers
func (r *Route) GetMatchers() matcher.Matchers {
	return r.Matchers
}

// Routes is a collection of routes
type Routes []*Route

// GetMetadatas returns an array of *RouteMeta
func (routes Routes) GetMetadatas() RouteMetas {
	metadatas := []*RouteMeta{}
	for _, route := range routes {
		metadatas = append(metadatas, route.GetMeta())
	}
	return metadatas
}

// ToMatcherProviders converts routes into matcher providers
func (routes Routes) ToMatcherProviders() matcher.MatcherProviders {
	matcherProviders := matcher.MatcherProviders{}
	for _, route := range routes {
		matcherProviders = append(matcherProviders, route)
	}
	return matcherProviders
}

// RouteMeta is a route metadata
type RouteMeta struct {
	Name         string
	Pattern      string
	Prefix       string
	URLVARPrefix string
	Methods      []string
	// ExtraData holds extra metadata
	// It can be used to add comments to routes
	// or annotations, in order to support Swagger for instance
	ExtraData map[interface{}]interface{}
}

// Get gets ExtraData by key
func (routeMeta *RouteMeta) Get(key interface{}) interface{} {
	return routeMeta.ExtraData[key]
}

// Set sets ExtraData
func (routeMeta *RouteMeta) Set(key interface{}, value interface{}) *RouteMeta {
	routeMeta.ExtraData[key] = value
	return routeMeta
}

// Clone returns an new RouteMeta with the same values as the current one
func (routeMeta RouteMeta) Clone() *RouteMeta {
	return &RouteMeta{Name: routeMeta.Name,
		Pattern:      routeMeta.Pattern,
		Prefix:       routeMeta.Prefix,
		URLVARPrefix: routeMeta.URLVARPrefix,
		Methods:      routeMeta.Methods}
}

// GetPath returns the Pattern prefixed by Prefix
func (routeMeta RouteMeta) GetPath() string {
	return path.Join(routeMeta.Prefix, routeMeta.Pattern)
}

// Generate generates a path by completing RouteMeta.Path with variables in attributes
func (routeMeta RouteMeta) Generate(attributes map[string]interface{}) string {
	path := regexp.MustCompile(`(\:[^\s /]+)`).ReplaceAllStringFunc(routeMeta.GetPath(), func(part string) string {
		value, ok := attributes[strings.TrimLeft(part, ":")]
		if ok {
			delete(attributes, strings.TrimLeft(part, ":"))
		}
		return fmt.Sprint(value)
	})
	if len(attributes) == 0 {
		return path
	}
	values := url.Values{}
	for key, value := range attributes {
		values.Set(key, url.QueryEscape(fmt.Sprint(value)))
	}
	return path + "?" + values.Encode()

}

// RouteMetas is a collection of *RouteMeta
type RouteMetas []*RouteMeta

// FindByName selects a *RouteMeta by name
// this method silently fails and will always return
// a route meta even if it isn't found. You'll need to
// compare to ZeroRouteMeta if you want to know whether
// a route was found or not.
func (routeMetas RouteMetas) FindByName(name string) *RouteMeta {
	for _, routeMeta := range routeMetas {
		if routeMeta.Name == name {
			return routeMeta
		}
	}
	return ZeroRouteMeta
}

var (
	ZeroRouteMeta = new(RouteMeta)
)
