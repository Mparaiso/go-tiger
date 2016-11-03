package db

import (
	"fmt"
	"strings"

	"github.com/Mparaiso/go-tiger/db/expression"
)

// QueryBuilder is a query builder
type QueryBuilder struct {
	connection   Connection
	builderType  Type
	sql          string
	state        State
	sqlParts     map[string][]interface{}
	maxResults   int
	firstResult  int
	isLimitquery bool
}

// NewQueryBuilder returns a new query builder
func NewQueryBuilder(connection Connection) *QueryBuilder {
	return &QueryBuilder{connection: connection, builderType: Select, sqlParts: map[string][]interface{}{}}
}

// GetType returns the Type
func (b QueryBuilder) GetType() Type {
	return b.builderType
}

// GetConnection returns the db connection
func (b QueryBuilder) GetConnection() Connection {
	return b.connection
}

// Select Specifies an item that is to be returned in the query result.
// Replaces any previously specified selections, if any.
func (b *QueryBuilder) Select(parts ...string) *QueryBuilder {
	b.builderType = Select
	content := []interface{}{}
	for _, part := range parts {
		content = append(content, part)
	}
	return b.add(select_, content, false)
}

// AddSelect adds an item that is to be returned in the query result.
func (b *QueryBuilder) AddSelect(parts ...string) *QueryBuilder {
	b.builderType = Select
	content := []interface{}{}
	for _, part := range parts {
		content = append(content, part)

	}
	return b.add(select_, content, true)
}

// From creates and adds a query root corresponding to the table identified by the
// given alias, forming a cartesian product with any existing query roots.
func (b *QueryBuilder) From(table string, alias ...string) *QueryBuilder {
	fromMap := From{Table: table}
	if len(alias) > 0 {
		fromMap.Alias = alias[0]
	}
	return b.add(from, []interface{}{fromMap}, true)
}
func (b *QueryBuilder) Where(parts ...interface{}) *QueryBuilder {
	// Note : in theory there should be only one where part
	if len(parts) > 1 {
		return b.add(where, []interface{}{expression.And(parts...)}, false)
	} else if len(parts) == 1 {
		return b.add(where, parts, false)
	}
	return b
}
func (b *QueryBuilder) OrWhere(parts ...interface{}) *QueryBuilder {
	wherePart := b.sqlParts[where]
	if len(wherePart) == 1 {
		if exp, ok := wherePart[0].(*expression.Expression); ok && exp.Type == expression.OR {
			return b.add(where, []interface{}{exp.Add(parts...)}, false)
		}
		return b.add(where,
			[]interface{}{expression.Or(append([]interface{}{wherePart[0]}, parts...)...)},
			false)
	}
	return b.add(where, []interface{}{expression.Or(parts...)}, false)
}

// AndWhere adds one or more restrictions to the query results, forming a logical
// conjunction with any previously specified restrictions.
func (b *QueryBuilder) AndWhere(parts ...interface{}) *QueryBuilder {
	wherePart := b.sqlParts[where]
	// if there is already a where part
	if len(wherePart) == 1 {
		// if this where part is an expression
		if exp, ok := wherePart[0].(*expression.Expression); ok && exp.Type == expression.AND {
			// if that expression is of type AND
			// just add the parts to that expression
			return b.add(where, []interface{}{exp.Add(parts...)}, false)
		}
		// if not and wherePart and parts to a AND expression
		return b.add(where,
			[]interface{}{expression.And(append([]interface{}{wherePart[0]}, parts...)...)},
			false)
	}
	return b.Where(parts...)
}

// GroupBy Specifies a grouping over the results of the query.
// Replaces any previously specified groupings, if any.
func (b *QueryBuilder) GroupBy(groups ...string) *QueryBuilder {
	if len(groups) == 0 {
		return b
	}
	convertedGroups := []interface{}{}
	for _, group := range groups {
		convertedGroups = append(convertedGroups, group)
	}
	return b.add(groupBy, convertedGroups, false)
}

// AddGroupBy adds a grouping expression to the query
func (b *QueryBuilder) AddGroupBy(groups ...string) *QueryBuilder {
	if len(groups) == 0 {
		return b
	}
	convertedGroups := []interface{}{}
	for _, group := range groups {
		convertedGroups = append(convertedGroups, group)
	}
	return b.add(groupBy, convertedGroups, true)
}

// Having Specifies a restriction over the groups of the query.
// Replaces any previous having restrictions, if any.
func (b *QueryBuilder) Having(parts ...interface{}) *QueryBuilder {
	return b.add(having, []interface{}{expression.And(parts...)}, false)
}

// AndHaving adds a restriction over the groups of the query, forming a logical
// conjunction with any existing having restrictions.
func (b *QueryBuilder) AndHaving(parts ...interface{}) *QueryBuilder {
	if len(b.sqlParts[having]) > 0 {
		return b.add(having, []interface{}{expression.And(b.sqlParts[having][0]).Add(parts...)}, false)
	}
	return b.add(having, []interface{}{expression.And(parts...)}, false)
}

// OrHaving adds a restriction over the groups of the query, forming a logical
// disjunction with any existing having restrictions.
func (b *QueryBuilder) OrHaving(parts ...interface{}) *QueryBuilder {
	if len(b.sqlParts[having]) > 0 {
		return b.add(having, []interface{}{expression.Or(b.sqlParts[having][0]).Add(parts...)}, false)
	}
	return b.add(having, []interface{}{expression.Or(parts...)}, false)
}

// OrderBy specifies an ordering for the query results.
// Replaces any previously specified orderings, if any.
func (b *QueryBuilder) OrderBy(order string, direction ...string) *QueryBuilder {
	if len(direction) == 0 {
		direction = []string{"ASC"}
	}
	return b.add(orderBy, []interface{}{order + " " + strings.Join(direction, "")}, false)
}

// AddOrderBy adds an ordering to the query results.
func (b *QueryBuilder) AddOrderBy(order string, direction ...string) *QueryBuilder {
	if len(direction) == 0 {
		direction = []string{"ASC"}
	}
	return b.add(orderBy, []interface{}{order + " " + strings.Join(direction, "")}, true)
}

// LeftJoin creates and adds a left join to the query.
func (b *QueryBuilder) LeftJoin(fromAlias, joinTable, joinAlias string, joinCondition ...interface{}) *QueryBuilder {
	if len(joinCondition) > 1 {
		joinCondition = []interface{}{expression.And(joinCondition...)}
	}
	return b.add(join, []interface{}{Join{
		FromAlias:  fromAlias,
		Table:      joinTable,
		Alias:      joinAlias,
		Type:       leftJoinType,
		Conditions: joinCondition,
	}}, true)
}

// Join creates and adds a join to the query.
func (b *QueryBuilder) Join(fromAlias, joinTable, joinAlias string, joinCondition ...interface{}) *QueryBuilder {
	if len(joinCondition) > 1 {
		joinCondition = []interface{}{expression.And(joinCondition...)}
	}
	return b.add(join, []interface{}{Join{
		FromAlias:  fromAlias,
		Table:      joinTable,
		Alias:      joinAlias,
		Type:       joinType,
		Conditions: joinCondition,
	}}, true)
}

func (b *QueryBuilder) InnerJoin(fromAlias, joinTable, joinAlias string, joinCondition ...interface{}) *QueryBuilder {
	if len(joinCondition) > 1 {
		joinCondition = []interface{}{expression.And(joinCondition...)}
	}
	return b.add(join, []interface{}{Join{
		FromAlias:  fromAlias,
		Table:      joinTable,
		Alias:      joinAlias,
		Type:       innerJoinType,
		Conditions: joinCondition,
	}}, true)
}

func (b *QueryBuilder) RightJoin(fromAlias, joinTable, joinAlias string, joinCondition ...interface{}) *QueryBuilder {
	if len(joinCondition) > 1 {
		joinCondition = []interface{}{expression.And(joinCondition...)}
	}
	return b.add(join, []interface{}{Join{
		FromAlias:  fromAlias,
		Table:      joinTable,
		Alias:      joinAlias,
		Type:       rightJoinType,
		Conditions: joinCondition,
	}}, true)
}

// Update turns the query being built into a bulk update query that ranges over
// a certain table
func (b *QueryBuilder) Update(table string, alias ...string) *QueryBuilder {
	b.builderType = Update
	if table == "" {
		return b
	}
	for _, a := range alias {
		return b.add(from, []interface{}{From{Table: table, Alias: a}}, false)
	}
	return b.add(from, []interface{}{From{Table: table}}, false)
}

// Set sets a new value for a column in a bulk update query.
func (b *QueryBuilder) Set(field string, value interface{}) *QueryBuilder {
	return b.add(set, []interface{}{field + " = " + fmt.Sprintf("%v", value)}, true)
}

func (b *QueryBuilder) Delete(table string, alias ...string) *QueryBuilder {
	b.builderType = Delete
	if len(alias) > 0 {
		for _, a := range alias {
			return b.add(from, []interface{}{From{Table: table, Alias: a}}, false)
		}
	}
	return b.add(from, []interface{}{From{Table: table}}, false)
}

func (b *QueryBuilder) Insert(table string) *QueryBuilder {
	b.builderType = Insert
	return b.add(insert, []interface{}{table}, false)
}

func (b *QueryBuilder) Values(FieldsAndValues map[string]interface{}) *QueryBuilder {
	return b.add(values, []interface{}{Values(FieldsAndValues)}, false)
}

func (b *QueryBuilder) SetValue(field string, value interface{}) *QueryBuilder {
	if valuePart, ok := b.sqlParts[values]; ok && len(valuePart) > 0 {
		valuePart[0].(Values)[field] = value
		return b
	}
	return b.add(values, []interface{}{Values(map[string]interface{}{field: value})}, false)
}

// convert converts an []interface{} to []string
func (b *QueryBuilder) convert(values ...interface{}) []string {
	result := []string{}
	for _, value := range values {
		result = append(result, fmt.Sprint(value))
	}
	return result
}

func (b *QueryBuilder) add(part string, contents []interface{}, doAppend bool) *QueryBuilder {
	if b.sqlParts[part] == nil || !doAppend {
		b.sqlParts[part] = contents
	} else {
		b.sqlParts[part] = append(b.sqlParts[part], contents...)
	}
	b.state = Dirty
	return b
}

func (b *QueryBuilder) getSQLForUpdate() string {
	query := update + " " + strings.Join(b.convert(b.sqlParts[from]...), "")
	query += " " + set + " " + strings.Join(b.convert(b.sqlParts[set]...), ", ")
	if wherePart, ok := b.sqlParts[where]; ok {
		query += " " + where + " " + strings.Join(b.convert(wherePart...), "")
	}
	return query
}

func (b *QueryBuilder) getSQLForSelect() string {
	query := "SELECT " + strings.Join(b.convert(b.sqlParts[select_]...), ", ")
	if fromPart, ok := b.sqlParts[from]; ok {
		query += " FROM " + strings.Join(b.convert(fromPart...), ", ")
	}
	if joinPart, ok := b.sqlParts[join]; ok {
		query += " " + strings.Join(b.convert(joinPart...), " ")
	}
	if wherePart, ok := b.sqlParts[where]; ok {
		query += " WHERE " + strings.Join(b.convert(wherePart...), "")
	}
	if groupByPart, ok := b.sqlParts[groupBy]; ok {
		query += " GROUP BY " + strings.Join(b.convert(groupByPart...), ", ")
	}
	if b.sqlParts[having] != nil {
		query += " HAVING " + strings.Join(b.convert(b.sqlParts[having]...), "")
	}
	if b.sqlParts[orderBy] != nil {
		query += " ORDER BY " + strings.Join(b.convert(b.sqlParts[orderBy]...), ", ")
	}
	if b.isLimitquery {
		return b.connection.GetDatabasePlatform().ModifyLimitQuery(
			query,
			b.maxResults,
			b.firstResult,
		)
	}
	return query
}

func (b *QueryBuilder) getSQLDelete() string {
	query := "DELETE FROM " + strings.Join(b.convert(b.sqlParts[from]...), "")
	if wherePart, ok := b.sqlParts[where]; ok {
		query += " WHERE " + strings.Join(b.convert(wherePart...), "")
	}
	if b.isLimitquery {
		return b.connection.GetDatabasePlatform().ModifyLimitQuery(
			query,
			b.maxResults,
			b.firstResult,
		)
	}
	return query

}

func (b *QueryBuilder) getSQLForInsert() string {
	query := "INSERT INTO " + strings.Join(b.convert(b.sqlParts[insert]...), "")
	query += " " + strings.Join(b.convert(b.sqlParts[values]...), "")
	return query
}

func (b *QueryBuilder) String() string {
	if b.sql != "" && b.state == Clean {
		return b.sql
	}
	switch b.builderType {
	case Insert:
		b.sql = b.getSQLForInsert()
	case Delete:
		b.sql = b.getSQLDelete()
	case Update:
		b.sql = b.getSQLForUpdate()
	default:
		b.sql = b.getSQLForSelect()
	}
	b.state = Clean
	return b.sql
}

type Type int

const (
	_ Type = iota
	Select
	Delete
	Update
	Insert
)

type State int

const (
	Dirty State = iota
	Clean
)

const (
	select_                = "SELECT"
	insert                 = "INSERT"
	values                 = "VALUES"
	update                 = "UPDATE"
	from                   = "FROM"
	where                  = "WHERE"
	groupBy                = "GROUP BY"
	having                 = "HAVING"
	orderBy                = "ORDER BY"
	set                    = "SET"
	join                   = "JOIN"
	joinType      JoinType = "JOIN"
	leftJoinType  JoinType = "LEFT JOIN"
	rightJoinType JoinType = "RIGHT JOIN"
	innerJoinType JoinType = "INNER JOIN"
)

type JoinType string

type Values map[string]interface{}

func (v Values) String() string {
	keys := []string{}
	values := []string{}
	for key, value := range v {
		keys = append(keys, key)
		values = append(values, fmt.Sprint(value))
	}
	return fmt.Sprintf("(%s) VALUES(%s)", strings.Join(keys, ", "), strings.Join(values, ", "))
}

// FROM represents a part of a FROM statement
type From struct {
	Table, Alias string
}

func (f From) String() string {
	return strings.Trim(f.Table+" "+f.Alias, " ")
}

// Join represents a JOIN statement
type Join struct {
	FromAlias, Table, Alias string
	Type                    JoinType
	Conditions              []interface{}
}

func (join Join) String() string {
	return fmt.Sprint(append([]interface{}{join.Type, " ", join.Table, " ", join.Alias, " ON "}, join.Conditions...)...)
}
