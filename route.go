package tiger

import (
	"fmt"

	matcher "github.com/Mparaiso/tiger-go-framework/matcher"
)

type Routes []matcher.MatcherProvider

// Route is an app route
type Route struct {
	Handler     func(Container)
	Matchers    []matcher.Matcher
	Middlewares []Middleware
	Name        string
}

func (r Route) String() string {
	return fmt.Sprintf("Route{Name:%+v,Handler:%+v,Middlewares:%+v}", r.Name, r.Handler, r.Middlewares)
}

type RouteOptions struct {
	Name        string
	Middlewares []Middleware
}

// GetMatchers return the request matchers
func (route *Route) GetMatchers() matcher.Matchers {
	return route.Matchers
}
