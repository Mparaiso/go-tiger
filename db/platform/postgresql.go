package platform

type PostgreSqlPlatform struct {
	DatabasePlatform
}

func NewPostgreSqlPlatform(databasePlatform DatabasePlatform) *PostgreSqlPlatform {
	return &PostgreSqlPlatform{DatabasePlatform: databasePlatform}
}
