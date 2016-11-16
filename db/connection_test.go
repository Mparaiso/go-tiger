package db_test

import (
	"database/sql"
	"flag"
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
	Name  string `sql:"name"`
	Email string `sql:"email"`
	*UserInfos
}

type UserInfos struct {
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
