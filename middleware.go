package tiger

import "net/http"

// StatusError is a status error
// it can be used to convert a http status to
// a go error interface
type StatusError int

func (se StatusError) Error() string {
	return http.StatusText(int(se))
}

func (se StatusError) Code() int {
	return int(se)
}

// Container contains server values
type Container interface {
	GetResponseWriter() http.ResponseWriter
	GetRequest() *http.Request
	Error(err error, statusCode int)
	Redirect(url string, statusCode int)
}

const (
	Debug int = iota
	Info
	Warning
	Error
	Critical
)

type LoggerProvider interface {
	GetLogger() (Logger, error)
	MustGetLogger() Logger
}

type Logger interface {
	Log(level int, args ...interface{})
	LogF(level int, format string, args ...interface{})
}

// DefaultContainer is the default implementation of the Container
type DefaultContainer struct {
	ResponseWriter http.ResponseWriter // ResponseWriter
	Request        *http.Request       // Request
}

// ResponseWriter returns a response writer
func (dc DefaultContainer) GetResponseWriter() http.ResponseWriter { return dc.ResponseWriter }

// Request returns a request
func (dc DefaultContainer) GetRequest() *http.Request { return dc.Request }

// Error writes an error to the client and logs an error to stdout
func (dc DefaultContainer) Error(err error, statusCode int) {
	http.Error(dc.GetResponseWriter(), err.Error(), statusCode)
}

// Redirect replies with a redirection
func (dc DefaultContainer) Redirect(url string, statusCode int) {
	http.Redirect(dc.GetResponseWriter(), dc.GetRequest(), url, statusCode)
}

// Handler is a controller that takes a context
type Handler func(Container)

// Wrap wraps Route.Handler with each middleware and returns a new Route
func (h Handler) Wrap(middlewares ...Middleware) Handler {
	return Queue(middlewares).Finish(h)
}

func (h Handler) Handle(c Container) {
	h(c)
}

// ToHandlerFunc converts Handler to http.Handler
func (h Handler) ToHandlerFunc(containerFactory func(http.ResponseWriter, *http.Request) Container) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Container
		if containerFactory == nil {
			c = &DefaultContainer{w, r}
		} else {
			c = containerFactory(w, r)
		}
		h(c)
	}
}

// ToMiddleware wraps a classic net/http middleware (func(http.HandlerFunc) http.HandlerFunc)
// into a Middleware compatible with this package
func ToMiddleware(classicMiddleware func(http.HandlerFunc) http.HandlerFunc) Middleware {
	return func(c Container, next Handler) {
		classicMiddleware(func(w http.ResponseWriter, r *http.Request) {
			next(&DefaultContainer{w, r})
		})(c.GetResponseWriter(), c.GetRequest())
	}
}

//Queue is a reusable queue of middlewares
type Queue []Middleware

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

func (m Middleware) Finish(h Handler) Handler {
	return func(c Container) {
		m(c, h)
	}
}
