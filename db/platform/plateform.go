package platform

import (
	"fmt"
)

// DatabasePlatform is a database platform
type DatabasePlatform interface {
	ModifyLimitQuery(query string, maxResults int, firstResult int) string
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

func (platform DefaultPlatform) ModifyLimitQuery(query string, limit, offset int) string {
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}
	return query
}
