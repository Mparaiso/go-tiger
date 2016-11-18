package db_test

func ExampleSchema() {
	connection := db.NewConnectionWithString("sqlite3", ":memory")
	schema := db.NewSchema()
	table := schema.CreateTable("my_table")
	table.AddColumn("id", "integer", map[string]interface{}{"unsigned": true})
	table.AddColumn("username", "string", map[string]interface{}{"length": 32})
	table.SetPrimaryKey("id")
	table.AddUniqueIndex("username")
	schema.CreateSequence("my_table_sequence")

	foreignTable := schema.CreateTable("my_foreign")
	foreignTable.AddColumn("id", "integer")
	foreignTable.AddColumn("user_id", "integer")
	foreignTable.AddForeignKeyConstraint(table, []string{"user_id"}, []string{"id"}, map[string]interface{}{
		"onUpdate": "CASCADE",
	})
	queries := schema.GetSQL(connection.GetDatabasePlatform())
	dropSchema := schema.GetDropSQL(connection.GetDatabasePlatform())
}
