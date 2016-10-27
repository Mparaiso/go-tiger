simple-row-mapper
-----------------

simple-row-mapper helps Go developers map db rows to structs and slice of structs.
With simple-row-mapper, developers no longer need to write a lot of boilerplate to 
transform db rows into structs as the result of a db query. simple-row-mapper is written in Go.

author: mparaiso <mparaiso@online.fr>

license: GPL-3.0

# Installation 

    go get github.com/mparaiso/simple-row-mapper-go


# Basic Usage

    package main

    import(
        "database/sql"
        _ "github.com/amattn/go-sqlite3"
        ex "github.com/mparaiso/expect-go"
        mapper "github.com/mparaiso/simple-row-mapper-go"
        "testing"
        "time"
        "fmt"
    )

    func main(){
        type User struct {
                ID           int64
                Name         string
                DateCreation *time.Time
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

        rows, _ := db.Query("SELECT id as ID,name as Name,date_creation as DateCreation FROM users;")
        users := []*User{}
        err = mapper.MapRowsToSliceOfStruct(rows, &users, false)
        fmt.Print(err) // Output: nil
        fmt.Print(len(users)) // Output: 2
    }