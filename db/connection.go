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

	"regexp"

	"strings"

	"github.com/Mparaiso/go-tiger/db/expression"
	"github.com/Mparaiso/go-tiger/db/platform"
	"github.com/Mparaiso/go-tiger/logger"
)

var (
	// ErrNotAPointer is yield when a pointer was expected
	ErrNotAPointer = fmt.Errorf("Error, value is not a pointer.")
	// ErrNotAStruct is yield when a struct was expected
	ErrNotAStruct = fmt.Errorf("Error this value is not a struct")
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
	Quote(input string, inputType ...string) string
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

// Delete all rows of a table matching the given identifier, where keys are column names.
func (connection *DefaultConnection) Delete(table string, criteria map[string]interface{}) (sql.Result, error) {
	qb := connection.CreateQueryBuilder().Delete(table)
	expression, data := mapToExpression(criteria)
	return qb.Where(expression).Exec(data...)
}

func (connection *DefaultConnection) QuoteIdentifier(identifier string) string {
	return connection.GetDatabasePlatform().QuoteIdentifier(identifier)
}

// Quote quotes a string
// TODO: obviously a PDO method, try to find an equivalent in Go
func (connection *DefaultConnection) Quote(input string, inputType ...string) string {
	return connection.GetDatabasePlatform().Quote(input, inputType...)
}

// Update executes an SQL UPDATE statement on a table.
func (connection *DefaultConnection) Update(table string, criteria map[string]interface{}, data interface{}) (sql.Result, error) {
	switch dataType := data.(type) {
	case map[string]interface{}:
		data := []interface{}{}
		qb := connection.CreateQueryBuilder().
			Update(table)
		for key, value := range dataType {
			qb.Set(key, "?")
			data = append(data, value)
		}
		for key, value := range criteria {
			qb.AndWhere(expression.Eq(key, "?"))
			data = append(data, value)
		}

		return qb.Exec(data...)
	default:
		Value := reflect.Indirect(reflect.ValueOf(data))
		Type := Value.Type()
		if reflect.Struct != Value.Kind() {
			return nil, ErrNotAStruct
		}
		values := map[string]interface{}{}
		for i := 0; i < Type.NumField(); i++ {
			Field := Value.Field(i)
			Tag, ok := Value.Type().Field(i).Tag.Lookup("sql")
			if ok {
				if Tag == "-" {
					continue
				}
				sqlTag := SQLStructTagBuilder{}.BuildFromString(Tag)
				if sqlTag.ColumnName != "" {
					values[sqlTag.ColumnName] = Field.Interface()
					continue
				}
			}
			values[Type.Field(i).Name] = Field.Interface()
		}
		return connection.Update(table, criteria, values)
	}
}

// Insert persists a new record into the database
// it supports both map[string]interface{} and structs
// as data. STruct field names ARE NOT automatically lowecased !
// use `slq:"column_name"` struct tag to explicitely denominate the db column
// omit a struct field with `sql:"-"` struct tag.
func (connection *DefaultConnection) Insert(tableName string, data interface{}) (sql.Result, error) {
	switch dataType := data.(type) {
	case map[string]interface{}:
		data := []interface{}{}
		qb := connection.CreateQueryBuilder().
			Insert(tableName)
		for key, value := range dataType {
			qb.SetValue(key, "?")
			data = append(data, value)
		}
		return qb.Exec(data...)
	default:
		Value := reflect.Indirect(reflect.ValueOf(data))
		Type := Value.Type()
		if Type.Kind() != reflect.Struct {
			return nil, ErrNotAStruct
		}
		data := map[string]interface{}{}
		for i := 0; i < Type.NumField(); i++ {
			fieldValue, fieldType := Value.Field(i), Value.Field(i).Type()
			stringTag, ok := Type.Field(i).Tag.Lookup("sql")
			if ok {
				tags := SQLStructTagBuilder{}.BuildFromString(stringTag)
				// if no "persistZeroValue" tag and value is zero then don't persist
				if !tags.PersistZeroValue && fieldValue.Interface() == reflect.Zero(fieldType).Interface() {
					continue
				}
				if tags.ColumnName != "" {
					data[tags.ColumnName] = fieldValue.Interface()
					continue
				} else if stringTag == "-" {
					// omit this field
					continue
				}
			}
			// if value is zero and no tag, don't persist field
			if fieldType.Comparable() && fieldValue.Interface() == reflect.Zero(fieldType).Interface() {
				continue
			}
			// use the field name as the db field name
			data[Type.Field(i).Name] = fieldValue.Interface()
		}
		return connection.Insert(tableName, data)
	}
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

// GetSingleResult returns the first column of the queried row
// pointer must be a pointer to the type of result
func (row *Row) GetSingleResult(pointer interface{}) error {
	defer row.rows.Close()
	if reflect.TypeOf(pointer).Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if !row.rows.Next() {
		return sql.ErrNoRows
	}
	return row.rows.Scan(pointer)
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

func mapToExpression(criteria map[string]interface{}) (Expression *expression.Expression, data []interface{}) {
	parts := []interface{}{}
	for key, value := range criteria {
		switch {
		case regexp.MustCompile(`^\w+$`).MatchString(strings.TrimSpace(key)):
			parts = append(parts, expression.Eq(key, "?"))
		default:
			parts = append(parts, key+" ? ")
		}
		data = append(data, value)
	}
	Expression = expression.And(parts...)
	return
}
