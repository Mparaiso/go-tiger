package tiger

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"path"

	matcher "github.com/Mparaiso/tiger-go-framework/matcher"
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
	return fmt.Sprintf("Route{Name:%s,Path:%s,Methods:%+v,Handler:%+v,Middlewares:%+v}", r.GetMeta().Name, r.GetMeta().Pattern,
		r.GetMeta().Methods,
		r.Handler, r.Middlewares)
}

// Routes is a collection of routes
type Routes []*Route

func (r Routes) GetMetadatas() RouteMetas {
	return nil
}

// ToMatcherProviders converts routes into matcher providers
func (r Routes) ToMatcherProviders() matcher.MatcherProviders {
	matcherProviders := matcher.MatcherProviders{}
	for _, route := range r {
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
}

// Clone returns an new RouteMeta with the same values as the current one
func (routeMeta RouteMeta) Clone() *RouteMeta {
	return &RouteMeta{Name: routeMeta.Name,
		Pattern:      routeMeta.Pattern,
		Prefix:       routeMeta.Prefix,
		URLVARPrefix: routeMeta.URLVARPrefix,
		Methods:      routeMeta.Methods}
}

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
	values := url.Values{}
	for key, value := range attributes {
		values.Set(key, url.QueryEscape(fmt.Sprint(value)))
	}
	return path + "?" + values.Encode()

}

// RouteMetas is a collection of *RouteMeta
type RouteMetas []*RouteMeta

// FindByName selects a *RouteMeta by name
func (routeMetas RouteMetas) FindByName(name string) *RouteMeta {
	for _, routeMeta := range routeMetas {
		if routeMeta.Name == name {
			return routeMeta
		}
	}
	return nil
}

type RouteOptions struct {
	Name        string
	Middlewares []Middleware
}

// GetMatchers return the request matchers
func (route *Route) GetMatchers() matcher.Matchers {
	return route.Matchers
}
