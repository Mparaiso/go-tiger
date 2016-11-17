package platform

import "fmt"
import "strings"

type SqlitePlatform struct {
	DatabasePlatform
}

func NewSqlitePlatform(databasePlatform DatabasePlatform) *SqlitePlatform {
	return &SqlitePlatform{DatabasePlatform: databasePlatform}
}

func (platform SqlitePlatform) GetListTableColumnsSQL(table string, database ...string) string {
	table = strings.Replace(".", "__", table, -1)
	table = platform.QuoteStringLiteral(table)
	return "PRAGMA table_info($table)"
}

func (platform SqlitePlatform) ModifyLimitQuery(query string, limit, offset int) string {
	if limit == 0 {
		return query + " LIMIT -1 " + fmt.Sprint(offset)
	}
	query += " LIMIT " + fmt.Sprint(limit)
	query += " OFFSET " + fmt.Sprint(offset)
	return query
}
