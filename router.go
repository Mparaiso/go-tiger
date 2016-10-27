package tiger

import (
	"net/http"
	"path"
	"sync"

	matcher "github.com/Mparaiso/tiger-go-framework/matcher"
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
	// Route variables in the route pattern
	// Will be available in request.URL.Query() prefixed by
	// UrlVarPrefix, the default prefix is ":"
	// ex : Given the pattern "/resource/:foo"
	//
	// 		foo := request.URL.Query().Get(":foo")
	//
	UrlVarPrefix string
}

func NewRouteCollection() *RouteCollection {
	return &RouteCollection{matchers: []matcher.Matcher{}, childRouteCollections: []*RouteCollection{}, routes: []*Route{}, middlewares: []Middleware{}}
}

type ContainerFactory interface {
	GetContainer(http.ResponseWriter, *http.Request) Container
}

type DefaultContainerFactory struct{}

func (d DefaultContainerFactory) GetContainer(w http.ResponseWriter, r *http.Request) Container {
	return DefaultContainer{w, r}
}

type RouteCollection struct {
	Prefix                string
	UrlVarPrefix          string
	matchers              []matcher.Matcher
	childRouteCollections []*RouteCollection
	routes                []*Route
	middlewares           []Middleware
}

func (r *RouteCollection) AddRequestMaster(matchers ...matcher.Matcher) *RouteCollection {
	r.matchers = append(r.matchers, matchers...)
	return r
}

func (r *RouteCollection) Use(middlewares ...Middleware) *RouteCollection {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

func (r *RouteCollection) Get(pattern string, handler Handler, routeOptions ...RouteOptions) *RouteCollection {
	return r.Match([]string{"GET"}, pattern, handler, routeOptions...)
}

func (r *RouteCollection) Post(pattern string, handler Handler, routeOptions ...RouteOptions) *RouteCollection {
	return r.Match([]string{"POST"}, pattern, handler, routeOptions...)
}

func (r *RouteCollection) Put(pattern string, handler Handler, routeOptions ...RouteOptions) *RouteCollection {
	return r.Match([]string{"PUT"}, pattern, handler, routeOptions...)
}

func (r *RouteCollection) Patch(pattern string, handler Handler, routeOptions ...RouteOptions) *RouteCollection {
	return r.Match([]string{"PATCH"}, pattern, handler, routeOptions...)
}

func (r *RouteCollection) Delete(pattern string, handler Handler, routeOptions ...RouteOptions) *RouteCollection {
	return r.Match([]string{"DELETE"}, pattern, handler, routeOptions...)
}

func (r *RouteCollection) Options(pattern string, handler Handler, routeOptions ...RouteOptions) *RouteCollection {
	return r.Match([]string{"OPTIONS"}, pattern, handler, routeOptions...)
}

func (r *RouteCollection) Match(methods []string, pattern string, handler Handler, routeOptions ...RouteOptions) *RouteCollection {
	route := &Route{
		Handler:  handler,
		Matchers: []matcher.Matcher{matcher.Pattern(pattern, r.Prefix, r.UrlVarPrefix), matcher.Method(methods...)},
	}
	if len(routeOptions) > 0 {
		route.Middlewares = routeOptions[0].Middlewares
		route.Name = routeOptions[0].Name
	}
	r.routes = append(r.routes, route)
	return r
}

func (r *RouteCollection) Compile() []*Route {
	routes := []*Route{}
	for _, route := range r.routes {
		compiledRoute := &Route{}
		compiledRoute.Handler = route.Handler
		compiledRoute.Matchers = route.Matchers
		compiledRoute.Matchers = append(append([]matcher.Matcher{}, r.matchers...), route.Matchers...)
		compiledRoute.Middlewares = append(append([]Middleware{}, r.middlewares...), route.Middlewares...)
		routes = append(routes, compiledRoute)
	}
	for _, routeCollection := range r.childRouteCollections {
		routeCollection.middlewares = append(append([]Middleware{}, r.middlewares...), routeCollection.middlewares...)

		routeCollection.matchers = append(append([]matcher.Matcher{}, r.matchers...), routeCollection.matchers...)
		routes = append(routes, routeCollection.Compile()...)
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

func (r *Router) Compile() http.Handler {
	routes := r.RouteCollection.Compile()
	matcherProviders := matcher.MatcherProviders{}
	for _, route := range routes {
		matcherProviders = append(matcherProviders, route)
	}

	return &httpHandler{matcherProviders, r.ContainerFactory}
}

type httpHandler struct {
	matcher.MatcherProviders
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
