package web

import (
	"net/http"
	"path"
)

// Collection is a collection of routes
type Collection interface {
	Get(string, Handler) *Route
	Post(string, Handler) *Route
	Put(string, Handler) *Route
	Delete(string, Handler) *Route
}

// Resource helps writing structured routing
type Resource interface {
	GetName() string
	Use(Container, Handler)
	Index(Container)
	Get(Container)
	Post(Container)
	Put(Container)
	Delete(Container)
}

// DefaultResource is the default implementation of Resource
type DefaultResource struct{}

// GetName returns the resource's name
func (DefaultResource) GetName() string { return "default_resourcce" }
func (DefaultResource) Use(container Container, next Handler) {
	next(container)
}

// Index lists resources
func (DefaultResource) Index(container Container) {
	container.Error(StatusError(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// Get shows a resource
func (DefaultResource) Get(container Container) {
	container.Error(StatusError(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// Post created a resource
func (DefaultResource) Post(container Container) {
	container.Error(StatusError(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// Put updated a resource
func (DefaultResource) Put(container Container) {
	container.Error(StatusError(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// Delete deletes a resource
func (DefaultResource) Delete(container Container) {
	container.Error(StatusError(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// MountResource mounts a resources in a collection
func MountResource(prefix string, routeCollection Collection, resource Resource) {
	routeCollection.Get(prefix, resource.Index).
		SetName("list_" + resource.GetName()).
		Use(resource.Use)
	routeCollection.Get(path.Join(prefix, ":"+resource.GetName()), resource.Get).
		SetName("show_" + resource.GetName()).
		Use(resource.Use)
	routeCollection.Post(prefix, resource.Post).SetName("create_" + resource.GetName())
	routeCollection.Put(path.Join(prefix, ":"+resource.GetName()), resource.Put).
		SetName("update_" + resource.GetName()).
		Use(resource.Use)

	routeCollection.Delete(path.Join(prefix, ":"+resource.GetName()), resource.Delete).
		SetName("delete_" + resource.GetName()).
		Use(resource.Use)
}
