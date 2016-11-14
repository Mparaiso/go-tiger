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
	GetDatabasePlatform() platform.DatabasePlatform
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

// GetDriverName returns the DriverName
func (connection *DefaultConnection) GetDriverName() string {
	return connection.DriverName
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
	case "sqlite3":
		connection.Platform = platform.NewSqlitePlatform(databasePlatform)
	case "mysql":
		connection.Platform = platform.NewMySqlPlatform(databasePlatform)
	case "postgresql":
		connection.Platform = platform.NewPostgreSqlPlatform(databasePlatform)
	default:
		connection.Platform = databasePlatform
	}
}

// DB returns Go standard *sql.DB type
func (connection *DefaultConnection) DB() *sql.DB {
	return connection.Db
}

// Prepare prepares a statement
func (connection *DefaultConnection) Prepare(sql string) (*sql.Stmt, error) {
	return connection.DB().Prepare(sql)
}

// Exec will execute a query like INSERT,UPDATE,DELETE.
func (connection *DefaultConnection) Exec(query string, parameters ...interface{}) (sql.Result, error) {
	defer connection.log(append([]interface{}{query}, parameters...)...)
	return connection.DB().Exec(query, parameters...)
}

// Select with fetch multiple records.
func (connection *DefaultConnection) Select(records interface{}, query string, parameters ...interface{}) error {
	defer connection.log(append([]interface{}{query}, parameters...)...)

	rows, err := connection.Db.Query(query, parameters...)
	if err != nil {
		return err
	}
	err = MapRowsToSliceOfStruct(rows, records, true)

	return err
}

// SelectMap queries the database and populates an array of mpa[string]interface{}
func (connection *DefaultConnection) SelectMap(Map *[]map[string]interface{}, query string, parameters ...interface{}) error {
	defer connection.log(append([]interface{}{query}, parameters...)...)

	rows, err := connection.Db.Query(query, parameters...)
	if err != nil {
		return err
	}
	return MapRowsToSliceOfMaps(rows, Map)

}

// SelectSlice queryies the database an populates an array of arrays
func (connection *DefaultConnection) SelectSlice(slices *[][]interface{}, query string, parameters ...interface{}) error {
	defer connection.log(append([]interface{}{query}, parameters...)...)

	rows, err := connection.Db.Query(query, parameters...)
	if err != nil {
		return err
	}
	return MapRowsToSliceOfSlices(rows, slices)

}

// Get will fetch a single record.
// expects record to be a pointer to a struct
// Exemple:
//    user := new(User)
//    err := connection.get(user,"SELECT * from users WHERE users.id = ?",1)
//
func (connection *DefaultConnection) Get(record interface{}, query string, parameters ...interface{}) error {
	// make a slice from the record type
	// pass a pointer to that slice to connection.Select
	// if the slice's length == 1 , put back the first value of that
	// slice in the record value.
	if reflect.TypeOf(record).Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	recordValue := reflect.ValueOf(record)
	recordType := recordValue.Type()
	sliceOfRecords := reflect.MakeSlice(reflect.SliceOf(recordType), 0, 1)
	pointerOfSliceOfRecords := reflect.New(sliceOfRecords.Type())
	pointerOfSliceOfRecords.Elem().Set(sliceOfRecords)
	//
	err := connection.Select(pointerOfSliceOfRecords.Interface(), query, parameters...)
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
