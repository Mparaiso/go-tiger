package platform

import "fmt"

type SqlitePlatform struct {
	DatabasePlatform
}

func NewSqlitePlatform(databasePlatform DatabasePlatform) *SqlitePlatform {
	return &SqlitePlatform{DatabasePlatform: databasePlatform}
}

func (platform SqlitePlatform) ModifyLimitQuery(query string, limit, offset int) string {
	if limit == 0 {
		return query + " LIMIT -1 " + fmt.Sprint(offset)
	}
	query += " LIMIT " + fmt.Sprint(limit)
	query += " OFFSET " + fmt.Sprint(offset)
	return query
}
