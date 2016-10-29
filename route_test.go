package tiger_test

import (
	"testing"

	"github.com/Mparaiso/expect-go"
	"github.com/Mparaiso/tiger-go-framework"
)

func TestRouteMeta(t *testing.T) {
	routeMeta := tiger.RouteMeta{Pattern: "/category/:category/resource/:id"}
	path := routeMeta.Generate(map[string]interface{}{"category": "movies", "id": 6000, "extra": "true"})
	expect.Expect(t, path, "/category/movies/resource/6000?extra=true")
}
