//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package web_test

import (
	"testing"

	"github.com/Mparaiso/expect-go"
	tiger "github.com/Mparaiso/go-tiger/web"
)

func TestRouteMeta(t *testing.T) {
	routeMeta := tiger.RouteMeta{Pattern: "/category/:category/resource/:id"}
	path := routeMeta.Generate(map[string]interface{}{"category": "movies", "id": 6000, "extra": "true"})
	expect.Expect(t, path, "/category/movies/resource/6000?extra=true")
}
