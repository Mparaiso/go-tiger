package platform

import (
	"fmt"
	"strings"
)

// DatabasePlatform is a database platform
type DatabasePlatform interface {
	GetListDatabaseSQL() string
	ModifyLimitQuery(query string, maxResults int, firstResult int) string
	GetListTableColumnsSQL(table string, database ...string) string
	GetListSequencesSQL() string
	GetStringLiteralQuoteCharacter() string
	QuoteStringLiteral(literal string) string
	Quote(string, ...string) string
	QuoteIdentifier(string) string
}

type PlatformOptions struct {
	CreateIndexes           int
	CreateForeignKeys       int
	DateIntervalUnitSecond  string
	DateIntervalUnitMinute  string
	DateIntervalUnitHour    string
	DateIntervalUnitDay     string
	DateIntervalUnitWeek    string
	DateIntervalUnitMonth   string
	DateIntervalUnitQuarter string
	DateIntervalUnitYear    string
	TrimUnspecified         int
	TrimLeading             int
	TrimTrailing            int
	TrimBoth                int
}
type DefaultPlatform struct {
	PlatformOptions
}

func NewDefaultPlatform() *DefaultPlatform {
	p := &DefaultPlatform{}
	p.PlatformOptions = PlatformOptions{
		CreateIndexes:           1,
		CreateForeignKeys:       2,
		DateIntervalUnitSecond:  "SECOND",
		DateIntervalUnitMinute:  "MINUTE",
		DateIntervalUnitHour:    "HOUR",
		DateIntervalUnitDay:     "DAY",
		DateIntervalUnitWeek:    "WEEK",
		DateIntervalUnitMonth:   "MONTH",
		DateIntervalUnitQuarter: "QUARTER",
		DateIntervalUnitYear:    "YEAR",
		TrimUnspecified:         0,
		TrimLeading:             1,
		TrimTrailing:            2,
		TrimBoth:                3,
	}
	return p
}

// TODO: complete
func (plateform DefaultPlatform) GetListSequencesSQL() string { return "" }

// TODO: complete
func (platform DefaultPlatform) GetListDatabaseSQL() string { return "" }

// TODO: complete
func (platform DefaultPlatform) GetListTableColumnsSQL(table string, database ...string) string {
	return ""
}

func (platform DefaultPlatform) ModifyLimitQuery(query string, limit, offset int) string {
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}
	return query
}

// Quote quotes a string
// TODO: obviously a PDO method, try to find an equivalent in Go
func (platform DefaultPlatform) Quote(input string, inputType ...string) string {
	return input
}

// QuoteIdentifier quotes a string so that it can be safely used as a table or column name,
// even if it is a reserved word of the platform. This also detects identifier
// chains separated by dot and quotes them independently.
func (platform DefaultPlatform) QuoteIdentifier(identifier string) string {
	if strings.Contains(identifier, ".") {
		parts := strings.Split(identifier, ".")
		return strings.Join(mapStringsToStrings(parts, func(str string) string {
			return platform.QuoteSingleIdentifier(str)
		}), ".")
	}
	return platform.QuoteSingleIdentifier(identifier)
}
func (platform DefaultPlatform) QuoteSingleIdentifier(identifier string) string {
	c := platform.GetIdentifierQuoteCharacter()
	return c + strings.Replace(c, c+c, identifier, -1) + c
}
func (platform DefaultPlatform) GetIdentifierQuoteCharacter() string {
	return `"`
}
func (platform DefaultPlatform) GetStringLiteralQuoteCharacter() string {
	return "'"
}

func (platform DefaultPlatform) QuoteStringLiteral(literal string) string {
	c := platform.GetStringLiteralQuoteCharacter()

	return c + strings.Replace(literal, c, c+c, -1) + c
}
