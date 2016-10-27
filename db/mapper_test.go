//    sql-row-mapper provides utitlies to map db rows to structs, maps and arrays
//
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

package mapper_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/amattn/go-sqlite3"
	ex "github.com/mparaiso/expect-go"
	mapper "github.com/mparaiso/tiger-go-framework/db"
)

func TestMapRowsToSliceOfStruct(t *testing.T) {
	type User struct {
		ID           int64
		Name         string
		DateCreation *time.Time
	}
	db, err := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	ex.Expect(t, err, nil, "sql.Open")
	for _, statement := range []string{
		`CREATE TABLE users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			date_creation TIMESTAMP NOT NULL DEFAULT(DATETIME('now'))
		);`,
		"INSERT INTO users(name) values('john doe');",
		"INSERT INTO users(name) values('jane doe');",
	} {
		_, err := db.Exec(statement)
		ex.Expect(t, err, nil, "db.Exec("+statement+")")
	}
	t.Run("Simple Row/Struct Mapping", func(t *testing.T) {
		rows, err := db.Query("SELECT id as ID,name as Name,date_creation as DateCreation FROM users;")
		ex.Expect(t, err, nil, "db.Query")
		users := []*User{}
		err = mapper.MapRowsToSliceOfStruct(rows, &users, false)
		ex.Expect(t, err, nil, "map rows to slice of structs")
		ex.Expect(t, len(users), 2, "len(users)")
	})
}
