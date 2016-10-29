package tiger

import (
	"net/http"
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
