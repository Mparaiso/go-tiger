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

package db_test

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Mparaiso/go-tiger/db"
	"github.com/Mparaiso/go-tiger/db/expression"
	"github.com/Mparaiso/go-tiger/logger"
	"github.com/Mparaiso/go-tiger/test"
	_ "github.com/amattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rubenv/sql-migrate"
)

type AppUser struct {
	ID    string `sql:"column:id"`
	Name  string `sql:"column:name"`
	Email string `sql:"column:email"`
	*UserInfos
}

type UserInfos struct {
}

func ExampleConnection() {
	t := test.ExampleTester{log.New(os.Stderr, "log-tester", log.LstdFlags)}
	// Define a type that represents a table
	type TestUser struct {
		ID           int64     `sql:"column:id"`
		Name         string    `sql:"column:name"`  // db field name will match the content of the tag if declared
		Email        string    `sql:"column:email"` // fieldnames are not automatically lower-cased to match db field names
		Created      time.Time `sql:"column:created"`
		Nickname     string    `sql:"column:nick_name,persistzerovalue"` // allow empty strings or zero values to be persisted
		PhoneNumbers []string  `sql:"-"`                                 // ignore field
	}
	var err error
	// initialize the driver
	DB, err := sql.Open("sqlite3", ":memory:")
	test.Fatal(t, err, nil)
	// create a connection
	connection := db.NewConnection("sqlite3", DB)
	connection.SetLogger(logger.NewDefaultLogger())
	// create a table
	_, err = connection.Exec(`
		CREATE TABLE users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			nick_name VARCHAR(255) NOT NULL,
			created TIMESTAMP NOT NULL DEFAULT(datetime('now'))
		);`)
	test.Fatal(t, err, nil)
	user := &TestUser{Name: "John Doe", Email: "john.doe@example.com"}
	// Insert a new user into the database
	result, err := connection.Insert("users", user)
	test.Fatal(t, err, nil)

	id, err := result.LastInsertId()
	test.Fatal(t, err, nil)

	fmt.Println("last inserted id", id)
	// insert another user thanks to the query builder
	result, err = connection.CreateQueryBuilder().
		Insert("users").
		SetValue("name", "?").
		SetValue("email", "?").
		SetValue("nick_name", "?").
		Exec("Jane Doe", "jane.doe@example.com", "Jannie")
	test.Fatal(t, err, nil)
	// fetch the first inserted user
	candidate := &TestUser{}
	err = connection.QueryRow("SELECT u.* FROM users u WHERE u.id = ?", id).
		GetResult(candidate)
	test.Fatal(t, err, nil)
	fmt.Println("candidate.Name:", candidate.Name)
	// update the record
	candidate.Name = "John Robert Doe"
	result, err = connection.Update("users", map[string]interface{}{"id": candidate.ID}, candidate)
	test.Fatal(t, err, nil)

	affectedRows, err := result.RowsAffected()
	test.Fatal(t, err, nil)
	fmt.Println("rows affected by update:", affectedRows)
	// delete the record
	result, err = connection.Delete("users", map[string]interface{}{"id": candidate.ID})
	test.Fatal(t, err, nil)
	affectedRows, err = result.RowsAffected()
	test.Fatal(t, err, nil)
	fmt.Println("rows affected by delete:", affectedRows)
	// let' make sure there is only one user in the database
	var count int
	err = connection.CreateQueryBuilder().
		Select("COUNT(u.id)").From("users", "u").
		QueryRow().
		GetSingleResult(&count)
	test.Fatal(t, err, nil)
	fmt.Println("user count:", count)
	// Output:
	// last inserted id 1
	// candidate.Name: John Doe
	// rows affected by update: 1
	// rows affected by delete: 1
	// user count: 1

}

func TestRowGetResult(t *testing.T) {
	connection := db.NewConnection(GetDB(t))
	err := LoadFixtures(connection)
	test.Fatal(t, err, nil)
	result := map[string]interface{}{}
	err = connection.CreateQueryBuilder().
		Select("u.*").
		From("users", "u").
		Where("u.name = ?").
		QueryRow("John Doe").
		GetResult(&result)
	test.Fatal(t, err, nil)
	test.Fatal(t, result["email"], "john.doe@acme.com")
}

func TestConnectionGet(t *testing.T) {
	connection := db.NewConnection(GetDB(t))
	err := LoadFixtures(connection)
	test.Fatal(t, err, nil)
	user := new(AppUser)
	err = connection.QueryRow("SELECT name as Name,email as Email from users ;").GetResult(user)
	test.Fatal(t, err, nil)
	test.Fatal(t, user.Name, "John Doe")
	user2 := AppUser{}
	err = connection.QueryRow("SELECT * from users ;").GetResult(user2)
	test.Fatal(t, err, db.ErrNotAPointer)

}

func TestConnectionSelect(t *testing.T) {
	connection := db.NewConnection(GetDB(t))
	result, err := connection.Exec("INSERT INTO users(name,email) values('john doe','johndoe@acme.com'),('jane doe','jane.doe@acme.com');")
	test.Fatal(t, err, nil)
	r, err := result.RowsAffected()
	test.Fatal(t, err, nil)
	test.Fatal(t, r, int64(2), "2 records should have been created")
	t.Log(result.LastInsertId())

	// test query
	users := []*AppUser{}
	err = connection.Query("SELECT users.name as Name, users.email as Email from users ORDER BY users.id ASC ;").GetResults(&users)
	test.Fatal(t, err, nil)
	t.Logf("%#v", users)
	test.Fatal(t, users[0].Name, "john doe")
}

func TestConnectionSelectMap(t *testing.T) {
	connection := GetConnection(t)
	defer connection.Close()
	if err := LoadFixtures(connection); err != nil {
		t.Fatal(err)
	}
	var result []map[string]interface{}
	err := connection.Query("SELECT * FROM users ORDER BY ID").GetResults(&result)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(result); l != 3 {
		t.Fatalf("length should be 3, got %d", l)
	}
	if created, ok := result[0]["created"].(time.Time); !ok {
		t.Fatalf("created should be time.Time, got %#v", created)
	}
}

func TestConnectionSelectSlice(t *testing.T) {
	connection := GetConnection(t)
	defer connection.Close()
	if err := LoadFixtures(connection); err != nil {
		t.Fatal(err)
	}
	var result [][]interface{}
	err := connection.Query("SELECT id,name,created FROM users ORDER BY ID").GetResults(&result)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(result); l != 3 {
		t.Fatalf("length should be 3, got %d", l)
	}
	if created, ok := result[0][2].(time.Time); !ok {
		t.Fatalf("created should be time.Time, got %#v", created)
	}
}
func TestConnectionPrepare(t *testing.T) {
	connection := GetConnection(t)
	defer connection.Close()
	result, err := connection.Prepare("SELECT * from users").Exec()
	test.Fatal(t, err, nil)
	test.Fatal(t, result != nil, true)
}

func TestConnectionCreateQueryBuilderPrepareExec(t *testing.T) {
	connection := GetConnection(t)
	user := &AppUser{Name: "robert", Email: "robert@example.com"}
	result, err := connection.CreateQueryBuilder().
		Insert("users").
		SetValue("name", "?").
		SetValue("email", "?").
		Prepare().Exec(user.Name, user.Email)
	test.Fatal(t, err, nil)
	lastInsertedID, err := result.LastInsertId()
	test.Fatal(t, err, nil)
	test.Fatal(t, lastInsertedID, int64(1))
}

func TestConnectionCreateQueryBuilderQuery(t *testing.T) {
	connection := GetConnection(t)
	err := LoadFixtures(connection)
	test.Fatal(t, err, nil)
	connection.SetLogger(logger.NewTestLogger(t))
	users := []*AppUser{}
	err = connection.CreateQueryBuilder().
		Select("u.name , u.email").
		From("users", "u").
		Where(expression.Neq("Name", "?")).
		OrderBy("name", "ASC").
		Query("Jack Doe").
		GetResults(&users)
	test.Fatal(t, err, nil)
	test.Fatal(t, len(users), 2)
	test.Fatal(t, users[0].Name, "Jane Doe")

}

func TestConnectionArray(t *testing.T) {
	connection := GetConnection(t)
	LoadFixtures(connection)
	var users []*AppUser
	err := connection.Query("SELECT * FROM users WHERE name IN (?,?)", []interface{}{"John Doe", "Jane Doe"}...).GetResults(&users)
	test.Fatal(t, err, nil)
	test.Fatal(t, len(users), 2)
}

/**
 * HELPERS
 */
// Fixtures are not loaded automatically, this function should
// be called explicitly
func LoadFixtures(connection *db.DefaultConnection) error {
	for _, user := range []AppUser{
		{Name: "John Doe", Email: "john.doe@acme.com"},
		{Name: "Jane Doe", Email: "jane.doe@acme.com"},
		{Name: "Jack Doe", Email: "jack.doe@acme.com"},
	} {
		_, err := connection.Exec("INSERT INTO users(name,email) values(?,?);", user.Name, user.Email)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetConnection(t *testing.T) *db.DefaultConnection {
	return db.NewConnection(GetDB(t))
}

func GetDB(t *testing.T) (string, *sql.DB) {
	driver, datasource, migrationDirectory := "sqlite3", ":memory:", "./testdata/migrations/development.sqlite3"
	arguments := flag.Args()
	for _, argument := range arguments {
		switch argument {
		case "mysql":
			// go test ./... -v -run ConnectionGet -args mysql
			// https://github.com/go-sql-driver/mysql#examples
			driver, datasource, migrationDirectory = "mysql", "user@/test?parseTime=true", "./testdata/migrations/test.mysql"
		}
	}
	t.Log("Using driver ", driver)
	db, err := sql.Open(driver, datasource)
	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: migrationDirectory,
	}
	_, err = migrate.Exec(db, driver, migrations, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	return driver, db
}

func DropDB(t *testing.T) *sql.DB {
	driver, datasource, migrationDirectory := "sqlite3", ":memory:", "./testdata/migrations/development.sqlite3"
	arguments := flag.Args()
	for _, argument := range arguments {
		switch argument {
		case "mysql":
			// go test ./... -v -run ConnectionGet -args mysql
			driver, datasource, migrationDirectory = "mysql", "user@/test?parseTime=true", "./testdata/migrations/test.mysql"

		default:
			return nil
		}
	}
	t.Log("Using driver ", driver)
	db, err := sql.Open(driver, datasource)
	if err != nil {
		t.Fatal(err)
	}
	migrations := &migrate.FileMigrationSource{
		Dir: migrationDirectory,
	}
	_, err = migrate.Exec(db, driver, migrations, migrate.Down)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestMain(m *testing.M) {
	code := m.Run()
	DropDB(new(testing.T))
	os.Exit(code)

}
