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

package matcher

import (
	"net/http"
	"path"
	"regexp"
)

type contextKeys int8

const (
	_ contextKeys = iota
	// URLValues is a map of url variable values
	URLValues
)

// Matcher matches a request
type Matcher interface {
	Match(*http.Request) bool
}

// Pattern returns a regexp matcher from a string
// like "/:foo/:bar"
func Pattern(pattern, pathPrefix string, queryValuePrefix ...string) *RegexpMatcher {
	pattern = regexp.MustCompile(`:(\w+)(!:\(.*\))`).ReplaceAllString(pattern, `${1}(\w+)`)
	re := regexp.MustCompile(`:(\w+)`)
	pattern = path.Join("^/", regexp.QuoteMeta(pathPrefix), re.ReplaceAllString(pattern, "(?P<${1}>[^\\s /]+)"), "/?$")
	if pattern == "^/?" {
		pattern = "^/$"
	}
	return NewRegexMatcher(regexp.MustCompile(pattern), queryValuePrefix...)
}

// Method is a shortcut for NewMethodMatcher
func Method(methods ...string) *MethodMatcher { return &MethodMatcher{methods} }

// MethodMatcher matches request by method
type MethodMatcher struct {
	Methods []string
}

// NewMethodMatcher returns a new MethodMatcher
func NewMethodMatcher(methods ...string) *MethodMatcher {
	return &MethodMatcher{methods}
}

// Match matches against the request method
func (mm *MethodMatcher) Match(r *http.Request) bool {
	for _, method := range mm.Methods {
		if method == r.Method {
			return true
		}
		if method == "GET" && r.Method == "HEAD" {
			return true
		}
	}
	return false
}

// URLMatcher is the most basic matcher
// It matches are url by Path
type URLMatcher struct {
	URL string
}

// NewURLMatcher returns a new url matcher
func NewURLMatcher(url string) *URLMatcher {
	return &URLMatcher{url}
}

// Match matches a URL by Path
func (matcher *URLMatcher) Match(r *http.Request) bool {
	if matcher.URL == r.URL.Path {
		return true
	}
	return false
}

// RegexpMatcher matches a path against a regexp
type RegexpMatcher struct {
	Regexp *regexp.Regexp // A regular expression that matches a path
	Prefix string
}

// NewRegexMatcher creates a new RegexpPathMatcher
func NewRegexMatcher(r *regexp.Regexp, prefix ...string) *RegexpMatcher {
	matcher := &RegexpMatcher{Regexp: r}
	if len(prefix) > 0 {
		matcher.Prefix = prefix[0]
	}
	return matcher
}

//Regex is a shortcut for NewRegexMatcher
func Regex(r *regexp.Regexp, prefix ...string) *RegexpMatcher { return NewRegexMatcher(r, prefix...) }

// Match matches a path against a regexp
func (pm *RegexpMatcher) Match(r *http.Request) bool {
	if pm.Prefix == "" {
		pm.Prefix = ":"
	}
	if pm.Regexp.MatchString(r.URL.Path) {

		// We want to take each url parameter and put it in the query string, prefixed by PathPrefix
		subMatches := pm.Regexp.FindStringSubmatch(r.URL.Path)
		subExNames := pm.Regexp.SubexpNames()
		originalValues := r.URL.Query()
		for i, name := range subExNames {
			if name == "" {
				// name = strconv.FormatInt(int64(i), 10)
				continue
			}
			originalValues.Set(pm.Prefix+name, subMatches[i])
		}
		r.URL.RawQuery = originalValues.Encode()
		return true
	}
	return false
}

// Matchers is a list of matchers
type Matchers []Matcher

// MatcherProvider provides matchers
type MatcherProvider interface {
	// GetMatchers returns a colleciton of matchers
	GetMatchers() Matchers
}

// DefaultMatcherProvider MatcherProvider
type DefaultMatcherProvider struct {
	Matchers
}

// GetMatchers returns a collection of matchers
func (r DefaultMatcherProvider) GetMatchers() Matchers {
	return r.Matchers
}

// Routes are a collection of routes
type MatcherProviders []MatcherProvider

// Router routes requests to routes
type RequestMatcher struct {
	MatcherProviders []MatcherProvider
}

func NewRequestMatcher(matcherProviders []MatcherProvider) *RequestMatcher {
	return &RequestMatcher{matcherProviders}
}

// Match returns the route that matches the request
func (requestMatcher *RequestMatcher) Match(r *http.Request) MatcherProvider {
	for _, matcherProvider := range requestMatcher.MatcherProviders {
		match := true
		for _, matcher := range matcherProvider.GetMatchers() {
			if !matcher.Match(r) {
				match = false
				break
			}
		}
		if match {
			return matcherProvider
		}
	}
	return nil
}
