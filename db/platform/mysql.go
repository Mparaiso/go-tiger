package platform

import "fmt"

type MySqlPlatform struct {
	DatabasePlatform
}

func NewMySqlPlatform(databasePlatform DatabasePlatform) *MySqlPlatform {
	return &MySqlPlatform{DatabasePlatform: databasePlatform}
}
func (platform MySqlPlatform) ModifyLimitQuery(query string, limit, offset int) string {
	if limit == 0 {
		query += " LIMIT 18446744073709551615"
	} else {
		query += " LIMIT " + fmt.Sprint(limit)
	}
	query += " OFFSET " + fmt.Sprint(offset)
	return query
}
