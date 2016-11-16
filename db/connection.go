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

package db

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/Mparaiso/go-tiger/db/platform"
	"github.com/Mparaiso/go-tiger/logger"
)

var (
	// ErrNotAPointer is yield when a pointer was expected
	ErrNotAPointer = fmt.Errorf("Error, value is not a pointer.")
)

// Connection is the db connection
type Connection interface {
	Begin() (*Transaction, error)
	Close() error
	DB() *sql.DB
	Ping() error
	CreateQueryBuilder() *QueryBuilder
	Exec(query string, parameters ...interface{}) (sql.Result, error)
	GetDatabasePlatform() platform.DatabasePlatform
	GetDriverName() string
	Prepare(sql string) *Statement
	Query(sql string, arguments ...interface{}) *Rows
	QueryRow(sql string, arguments ...interface{}) *Row
	SetLogger(logger.Logger)
}

// ConnectionOptions gather options related to Connection type.
type ConnectionOptions struct {
	Logger              logger.Logger
	IgnoreMissingFields bool
}

// DefaultConnection is a database connection.
// Please use NewConnectionto create a Connection.
type DefaultConnection struct {
	Db         *sql.DB
	DriverName string
	Options    *ConnectionOptions
	Platform   platform.DatabasePlatform
}

// NewConnection creates an new Connection
func NewConnection(driverName string, DB *sql.DB) *DefaultConnection {
	connection := NewConnectionWithOptions(driverName, DB, &ConnectionOptions{})
	return connection
}

// NewConnectionWithOptions creates an new connection with optional settings such as Logging.
func NewConnectionWithOptions(driverName string, DB *sql.DB, options *ConnectionOptions) *DefaultConnection {
	return &DefaultConnection{Db: DB, DriverName: driverName, Options: options}
}
func (connection *DefaultConnection) GetOptions() *ConnectionOptions {
	if connection.Options == nil {
		connection.Options = &ConnectionOptions{}
	}
	return connection.Options
}
func (connection *DefaultConnection) SetLogger(Logger logger.Logger) {
	connection.GetOptions().Logger = Logger
}

// GetDriverName returns the DriverName
func (connection *DefaultConnection) GetDriverName() string {
	return connection.DriverName
}

// CreateQueryBuilder creates a *QueryBuilder value
func (connection *DefaultConnection) CreateQueryBuilder() *QueryBuilder {
	return NewQueryBuilder(connection)
}

// GetDatabasePlatform returns the database platform
func (connection *DefaultConnection) GetDatabasePlatform() platform.DatabasePlatform {
	if connection.Platform == nil {
		connection.detectDatabasePlatform()
	}
	return connection.Platform
}
func (connection *DefaultConnection) detectDatabasePlatform() {
	databasePlatform := platform.NewDefaultPlatform()
	switch connection.GetDriverName() {
	case "sqlite3", "sqlite":
		connection.Platform = platform.NewSqlitePlatform(databasePlatform)
	case "mysql":
		connection.Platform = platform.NewMySqlPlatform(databasePlatform)
	case "postgresql":
		connection.Platform = platform.NewPostgreSqlPlatform(databasePlatform)
	default:
		connection.Platform = databasePlatform
	}
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (connection *DefaultConnection) Ping() error {
	return connection.Db.Ping()
}

// DB returns Go standard *sql.DB type
func (connection *DefaultConnection) DB() *sql.DB {
	return connection.Db
}

// Prepare prepares a statement
func (connection *DefaultConnection) Prepare(sql string) *Statement {
	stmt, err := connection.DB().Prepare(sql)
	return &Statement{query: sql, logger: connection.Options.Logger, statement: stmt, err: err}
}

// Exec will execute a query like INSERT,UPDATE,DELETE.
func (connection *DefaultConnection) Exec(query string, parameters ...interface{}) (sql.Result, error) {
	defer connection.log(append([]interface{}{query}, parameters...)...)
	return connection.DB().Exec(query, parameters...)
}

// Query queries the database and creates Rows that can then be iterated upon
func (connection *DefaultConnection) Query(query string, parameters ...interface{}) *Rows {
	defer connection.log(append([]interface{}{query}, parameters...)...)

	rows, err := connection.Db.Query(query, parameters...)

	return &Rows{err: err, rows: rows}
}

// QueryRow will fetch a single record.
// expects record to be a pointer to a struct
// Exemple:
//    user := new(User)
//    err := connection.get(user,"SELECT * from users WHERE users.id = ?",1)
//
func (connection *DefaultConnection) QueryRow(query string, arguments ...interface{}) *Row {
	// make a slice from the record type
	// pass a pointer to that slice to connection.Select
	// if the slice's length == 1 , put back the first value of that
	// slice in the record value.
	rows, err := connection.Db.Query(query, arguments...)
	if err != nil {
		return &Row{err: err}
	}
	return &Row{rows: rows}
}

func (connection *DefaultConnection) log(messages ...interface{}) {
	if connection.Options.Logger != nil {
		connection.Options.Logger.Log(logger.Debug, messages...)
	}
}

// Begin initiates a transaction
func (connection *DefaultConnection) Begin() (*Transaction, error) {
	defer connection.log("Begin transaction")
	transaction, err := connection.DB().Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{Logger: connection.Options.Logger, Tx: transaction}, nil
}

// Close closes the connection
func (connection *DefaultConnection) Close() error {
	return connection.Db.Close()
}

// Statement is a prepared statement and a wrapper around *sql.Stmt
type Statement struct {
	statement *sql.Stmt
	query     string
	logger    logger.Logger
	err       error
}

// Exec executes a prepared statement with the given arguments and
// returns a Result summarizing the effect of the statement.
func (statement *Statement) Exec(arguments ...interface{}) (sql.Result, error) {
	if statement.err != nil {
		return nil, statement.err
	}
	defer statement.log("Executing statement", statement.query, arguments)
	return statement.statement.Exec(arguments...)
}

// GetStatement returns the original statement
func (statement *Statement) GetStatement() *sql.Stmt {
	return statement.statement
}
func (statement Statement) log(arguments ...interface{}) {
	if statement.logger != nil {
		statement.logger.Log(logger.Debug, arguments...)
	}
}

// Query executes a prepared query statement with the given arguments
// and returns the query results as a *Rows.
func (statement *Statement) Query(arguments ...interface{}) *Rows {
	if statement.err != nil {
		return &Rows{err: statement.err}
	}
	defer statement.log(statement.query, arguments)
	rows, err := statement.statement.Query(arguments...)
	return &Rows{rows: rows, err: err}
}

// QueryRow executes a prepared query statement with the given arguments.
func (statement *Statement) QueryRow(arguments ...interface{}) *Row {
	if statement.err != nil {
		return &Row{err: statement.err}
	}
	defer statement.log(statement.query, arguments)
	rows, err := statement.statement.Query(arguments)
	return &Row{rows: rows, err: err}
}

// Row is wrapper around *sql.Row
type Row struct {
	err  error
	rows *sql.Rows
	row  *sql.Row
}

func NewRow(rows *sql.Rows, err error) *Row {
	return &Row{rows: rows, err: err}
}

// GetResult assigns the db row to pointerToStruct
func (row *Row) GetResult(pointerToStruct interface{}) error {
	defer row.rows.Close()
	if reflect.TypeOf(pointerToStruct).Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	recordValue := reflect.ValueOf(pointerToStruct)
	recordType := recordValue.Type()
	sliceOfRecords := reflect.MakeSlice(reflect.SliceOf(recordType), 0, 1)
	pointerOfSliceOfRecords := reflect.New(sliceOfRecords.Type())
	pointerOfSliceOfRecords.Elem().Set(sliceOfRecords)
	//
	err := MapRowsToSliceOfStruct(row.rows, pointerOfSliceOfRecords.Interface(), true)
	if err != nil {
		return err
	}
	if pointerOfSliceOfRecords.Elem().Len() >= 1 {
		recordValue.Elem().Set(reflect.Indirect(pointerOfSliceOfRecords).Index(0).Elem())
	} else {
		return sql.ErrNoRows
	}
	return nil
}

// Rows is a wrapper around *sql.Rows
// Allowing to map db rows to structs directly
type Rows struct {
	err  error
	rows *sql.Rows
}

func NewRows(rows *sql.Rows, err error) *Rows {
	return &Rows{rows: rows, err: err}
}

// GetRows returns the underlying *sql.Rows
func (rows *Rows) GetRows() (*sql.Rows, error) {
	if rows.err != nil {
		return nil, rows.err
	}
	return rows.rows, nil
}

// GetResults assign results to pointer .
// pointer can either be :
// - a pointer to a slice of structs
// - a pointer to a slice of slice of empty intefaces : *[][]interface{}
// - a pointer to a slice of map of strings as key and empty interfaces as value : *[]map[string]interface{}
func (rows *Rows) GetResults(pointer interface{}) error {
	if rows.err != nil {
		return rows.err
	}
	defer rows.rows.Close()

	switch Type := pointer.(type) {
	case *[][]interface{}:
		return MapRowsToSliceOfSlices(rows.rows, Type)
	case *[]map[string]interface{}:
		return MapRowsToSliceOfMaps(rows.rows, Type)
	default:
		return MapRowsToSliceOfStruct(rows.rows, Type, true)
	}
}
