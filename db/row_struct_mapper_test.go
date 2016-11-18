//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//      http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package db_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Mparaiso/go-tiger/db"
	"github.com/Mparaiso/go-tiger/logger"
	"github.com/Mparaiso/go-tiger/test"
	_ "github.com/amattn/go-sqlite3"
	ex "github.com/mparaiso/expect-go"
	mapper "github.com/mparaiso/go-tiger/db"
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

func TestSQLTagBuilder(t *testing.T) {
	tag := "column:nick_name"
	sqlTag := db.SQLStructTagBuilder{logger.NewTestLogger(t)}.BuildFromString(tag)
	test.Fatal(t, sqlTag.ColumnName, "nick_name")
	test.Fatal(t, sqlTag.PersistZeroValue, false)
	tag = "column:email_address,persistzerovalue"
	sqlTag = db.SQLStructTagBuilder{logger.NewTestLogger(t)}.BuildFromString(tag)
	test.Fatal(t, sqlTag.ColumnName, "email_address")
	test.Fatal(t, sqlTag.PersistZeroValue, true)
}

func Example() {
	type User struct {
		ID           int64      `sql:"column:id"`
		Name         string     `sql:"column:name"`
		DateCreation *time.Time `sql:"column:date_creation"`
	}
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	for _, statement := range []string{
		`CREATE TABLE users(
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                date_creation TIMESTAMP NOT NULL DEFAULT(DATETIME('now'))
            );`,
		"INSERT INTO users(name) values('john doe'),('jane doe');",
	} {
		db.Exec(statement)
	}

	rows, _ := db.Query("SELECT * FROM users;")
	users := []*User{}
	err := mapper.MapRowsToSliceOfStruct(rows, &users, false)
	fmt.Println(err)
	fmt.Println(len(users))
	fmt.Println(users[0].Name)
	// Output:
	// <nil>
	// 2
	// john doe

}
