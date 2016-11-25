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
	"encoding/json"
	"log"
	"net/http"

	"net/url"

	"github.com/Mparaiso/go-tiger/logger"
	tiger "github.com/Mparaiso/go-tiger/web"
)

func main() {
	// tiger support dependency injection like Martini
	// it is completely optional.
	// All is needed is to                          wrap the default container into
	// a container that has an injector provider
	// then wrap handlers with tiger.Inject
	router := tiger.NewRouter()
	router.Use(func(c tiger.Container, next tiger.Handler) {
		container := &tiger.ContainerWithInjector{Container: c}
		container.GetInjector().SetLogger(&logger.DefaultLogger{})
		// let's register a few services for demonstration purposes
		container.GetInjector().RegisterValue(&DB{"mysql://localhost"})
		container.GetInjector().MustRegisterFactory(func(w http.ResponseWriter) (*json.Encoder, error) { return json.NewEncoder(w), nil })
		container.GetInjector().MustRegisterFactory(func(r *http.Request) (url.Values, error) { return r.URL.Query(), nil })
		// we need to pass the wrapped container to the next handlers
		next(container)
	})
	router.Post("/user/:name", tiger.Inject(func(db *DB, encoder *json.Encoder, w http.ResponseWriter, query url.Values) {
		id, err := db.Execute("INSERT INTO TABLE FOO(name) values(?)", query.Get(":name"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		encoder.Encode(map[string]interface{}{"Status": "Created", "ID": id})
	}))
	addr := ":8080"
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router.Compile()))
}

// DB is a database stub
type DB struct {
	ConnectionString string
}

// Execute stubs a command to the database
func (db *DB) Execute(query string, arguments ...interface{}) (int, error) {
	return 1, nil
}
