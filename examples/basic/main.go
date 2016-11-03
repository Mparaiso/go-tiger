//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.
//
//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Mparaiso/go-tiger"
)

// Run this basic example with :
// go run examples/basic/main.go

func main() {
	addr := "127.0.0.1:8080"
	// Create a router
	router := tiger.NewRouter()
	// Use a tiger.Handler to read url variables
	router.Get("/greetings/:name", func(container tiger.Container) {
		name := container.GetRequest().URL.Query().Get(":name")
		fmt.Fprintf(container.GetResponseWriter(), "Hello %s ! ", name)
	})
	// Use an idiomatic http.Handlerfunc as the app index
	router.Get("/", tiger.ToHandler(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
		<html>
			<body>
				<h3>Index</h3>
				<p>Welcome to tiger !
				<p><a href="/secure">Login</a>
			</body>
		</html>
		`)
	}))

	// Compile and use with http.Server
	log.Printf("Listening on %s \n", addr)
	log.Fatal(http.ListenAndServe(addr, router.Compile()))
}
