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
	"github.com/Mparaiso/go-tiger/db/platform"
	"github.com/Mparaiso/go-tiger/db/schema"
)

// SchemaManager manages db schema
type SchemaManager interface {
	ListDatabases() ([]*schema.Database, error)
	ListSequences(database ...string) ([]*schema.Sequence, error)
	ListTableColumns(table string, database ...string) ([]*schema.Table, error)
}

// DefaultSchemaManager is a default implementation of SchemaManager
type DefaultSchemaManager struct {
	connection Connection
	platform   platform.DatabasePlatform
}

// ListSequences lists the available sequences for this connection
func (sm DefaultSchemaManager) ListSequences(database ...string) ([]*schema.Sequence, error) {
	return nil, nil
}
func (sm DefaultSchemaManager) ListDatabases() ([]*schema.Database, error) {
	sql := sm.GetPlatform().GetListDatabaseSQL()
	var databases []map[string]interface{}
	err := sm.GetConnection().Query(sql).GetResults(&databases)
	if err != nil {
		return nil, err
	}
	return sm.getPortableDatabaseList(databases), err
}

// TODO: implement
func (sm DefaultSchemaManager) ListTableColumns(table string, database ...string) ([]*schema.Column, error) {
	if len(database) == 0 {
		database = []string{""}
	}
	sql := sm.GetPlatform().GetListTableColumnsSQL(table, database[0])
	var tableColumns []map[string]interface{}
	err := sm.GetConnection().Query(sql).GetResults(&tableColumns)
	if err != nil {
		return nil, err
	}
	return sm.getPortableTableColumnList(table, database[0], tableColumns), nil
}

// TODO: implement
func (sm DefaultSchemaManager) getPortableTableColumnList(table, database string, tableColumns []map[string]interface{}) []*schema.Column {
	return []*schema.Column{}
}

// TODO: implement
func (sm DefaultSchemaManager) getPortableTableColumnDefinition(tableColumn map[string]interface{}) *schema.Column {
	return &schema.Column{}
}

// TODO: implement
func (sm DefaultSchemaManager) getPortableDatabaseList(databases []map[string]interface{}) []*schema.Database {
	return []*schema.Database{}
}
func (sm DefaultSchemaManager) GetPlatform() platform.DatabasePlatform {
	return sm.platform
}

func (sm DefaultSchemaManager) GetConnection() Connection {
	return sm.connection
}

func NewDefaultSchemaManager(connection Connection, Platform platform.DatabasePlatform) SchemaManager {
	return &DefaultSchemaManager{connection: connection, platform: Platform}
}

type SqliteSchemaManager struct {
	SchemaManager
}

func NewSqliteSchemaManager(sm SchemaManager) SchemaManager {
	return &SqliteSchemaManager{sm}
}

func (sm SqliteSchemaManager) ListDatabases() ([]*schema.Database, error) {
	return []*schema.Database{}, nil
}

// ListSequences lists the available sequences for this connection
func (sm SqliteSchemaManager) ListSequences(database ...string) ([]*schema.Sequence, error) {
	return nil, ErrUnsupportedMethod
}
func NewMysqlSchemaManager(sm SchemaManager) SchemaManager {
	return sm
}

func NewPostgreSQLSchemaManager(sm SchemaManager) SchemaManager {
	return sm
}
