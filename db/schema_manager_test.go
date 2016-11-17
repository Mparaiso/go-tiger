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
	"log"
	"os"

	"fmt"

	"github.com/Mparaiso/go-tiger/db"
	"github.com/Mparaiso/go-tiger/test"
)

func ExampleSchemaManager() {
	//  Schema-Manager
	// @see http://docs.doctrine-project.org/projects/doctrine-dbal/en/latest/reference/schema-manager.html
	t := test.ExampleTester{log.New(os.Stdout, "tester", log.LstdFlags)}
	DB, err := sql.Open("sqlite3", ":memory:")
	test.Fatal(t, err, nil)
	connection := db.NewConnection("sqlite3", DB)
	sm := connection.GetSchemaManager()
	databases, err := sm.ListDatabases()
	test.Error(t, err, nil)
	fmt.Println("databases:", databases)
	sequences, err := sm.ListSequences()
	test.Error(t, err, nil)
	fmt.Println("sequences:", sequences)
	columns, err := sm.ListTableColumns("users")
	test.Error(t, err, nil)
	for _, column := range columns {
		fmt.Println(column.GetName(), column.GetType())
	}
	table := sm.ListTableDetails("users")
	fmt.Print(table.AddColumn("email", "string"))
	foreignKeys := sm.ListTableForeignKeys("user")
	for _, foreignKey := range foreignKeys {
		fmt.Println(foreignKey.GetName(), foreignKey.GetLocalTableName())
	}
	tables := sm.ListTables()
	for _, table := range tables {
		fmt.Println(table.GetName())
		for _, column := range table.GetColumns {
			fmt.Println("\t", column.GetName())
		}
	}
	views := sm.ListViews()
	for _, view := range views {
		fmt.Println(view.GetName(), view.GetSql())
	}
	fromSchema := sm.CreateSchema()
	toSchema := fromSchema.Clone()
	toSchema.DropTable("users")
	sql := fromSchema.GetMigrateSql(toSchema, connection.GetDatabasePlatform())
	fmt.Println(sql)

	// Output:
	//
}
