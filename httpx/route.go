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
	"errors"
	"net/http"
	"strings"
)

var errRouteNotFound = errors.New("Error route not found")

type Route struct {
	methods map[string]bool
	path    string
	handler http.Handler
	name    string
	filters []Filter
}

func NewRoute(path string, handler http.Handler, option ...RouteOption) Route {
	r := Route{path: path, handler: handler}
	for _, option := range option {
		option.handle(&r)
	}
	return r
}
func (r Route) Name() string          { return r.name }
func (r Route) Path() string          { return r.path }
func (r Route) Handler() http.Handler { return r.handler }
func (r Route) Methods() (methods []string) {
	for method := range r.methods {
		methods = append(methods, method)
	}
	return methods
}

func (r Route) Match(path string) (variables map[string]string, match bool) {
	if path == "/" && r.path == "/" {
		return variables, true
	}
	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	variables = map[string]string{}
	patternParts := strings.Split(r.path, "/")
	parts := strings.SplitN(path, "/", len(patternParts))
	if len(parts) < len(patternParts) {
		return variables, false
	}
	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, ":") {
			if strings.Contains(parts[i], "/") && !strings.HasPrefix(patternPart, ":*") {
				return variables, false
			}
			variables[patternPart] = parts[i]
		} else {
			if patternPart != parts[i] {
				return variables, false
			}
		}
	}
	return variables, true
}

func (r Route) HasMethod(m string) bool {
	if len(r.methods) == 0 {
		return true
	}
	for method := range r.methods {
		if method == m {
			return true
		}
	}
	return false
}
