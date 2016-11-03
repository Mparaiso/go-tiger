package expression

import "fmt"

// ExpressionType is the type of an expression
type ExpressionType int

const (
	_ ExpressionType = iota
	// EQ is =
	EQ
	// NEQ is <>
	NEQ
	// LT is <
	LT
	// LTE is <=
	LTE
	// GT is >
	GT
	// GTE is >=
	GTE
	// AND is AND
	AND
	// OR is OR
	OR
	// ISNULL is 'NULL'
	ISNULL
	// ISNOTNULL is 'NOT NULL'
	ISNOTNULL
	// LIKE is LIKE
	LIKE
	// NOTLIKE is 'NOT LIKE'
	NOTLIKE
	// IN is IN
	IN
	// NOTIN is 'NOT IN'
	NOTIN
)

func (et ExpressionType) String() string {
	switch et {
	case EQ:
		return "="
	case NEQ:
		return "<>"
	case LT:
		return "<"
	case LTE:
		return "<="
	case GT:
		return ">"
	case GTE:
		return ">="
	case AND:
		return "AND"
	case OR:
		return "OR"
	case ISNULL:
		return "IS NULL"
	case ISNOTNULL:
		return "IS NOT NULL"
	case LIKE:
		return "LIKE"
	case NOTLIKE:
		return "NOT LIKE"
	case IN:
		return "IN"
	case NOTIN:
		return "NOT IN"
	default:
		return ""
	}
}

// Expression is an SQL expression
type Expression struct {
	Type  ExpressionType
	Parts []interface{}
}

// Add parts to expression parts
func (e *Expression) Add(parts ...interface{}) *Expression {
	e.Parts = append(e.Parts, parts...)
	return e
}

// Count returns the number of parts
func (e Expression) Count() int {
	return len(e.Parts)
}

func (e Expression) String() string {
	switch e.Type {
	case EQ, NEQ, LT, GT, GTE, LIKE, NOTLIKE:
		return fmt.Sprintf("%s %s %v", e.Parts[0], e.Type, e.Parts[1])
	case AND, OR:
		if len(e.Parts) <= 1 {
			return fmt.Sprint(e.Parts...)
		}
		result := ""
		parts := []interface{}{}
		for _, part := range e.Parts {
			if result == "" {
				result += "(%s)"
				parts = append(parts, part)
			} else {
				result += " %s (%s)"
				parts = append(parts, e.Type, part)
			}
		}
		return fmt.Sprintf(result, parts...)
	case ISNULL, ISNOTNULL:
		return fmt.Sprintf("%s %s", e.Parts[0], e.Type)
	case IN, NOTIN:
		result := ""
		parts := []interface{}{e.Parts[0], e.Type}
		for i := 1; i < len(e.Parts); i++ {
			if result == "" {
				result += "( %v"
			} else {
				result += " , %v"
			}
			parts = append(parts, e.Parts[i])
		}
		result = "%s %s " + result + " )"
		return fmt.Sprintf(result, parts...)
	default:
		return ""
	}
}

// Eq is field = value
func Eq(field string, value interface{}) *Expression {
	return &Expression{Type: EQ, Parts: []interface{}{field, value}}
}

// Neq is field != value
func Neq(field string, value interface{}) *Expression {
	return &Expression{Type: NEQ, Parts: []interface{}{field, value}}
}

// Lt is field < value
func Lt(field string, value interface{}) *Expression {
	return &Expression{Type: LT, Parts: []interface{}{field, value}}
}

// Lt is field <= value
func Lte(field string, value interface{}) *Expression {
	return &Expression{Type: LTE, Parts: []interface{}{field, value}}
}

// Gt is field > value
func Gt(field string, value interface{}) *Expression {
	return &Expression{Type: GT, Parts: []interface{}{field, value}}
}

// Gte is field >= value
func Gte(field string, value interface{}) *Expression {
	return &Expression{Type: GTE, Parts: []interface{}{field, value}}
}

// Or is (part1) OR (part2) OR ...
func Or(parts ...interface{}) *Expression {
	return &Expression{Type: OR, Parts: parts}
}

// And is (part1) AND (part2) AND (part3) AND ...
func And(parts ...interface{}) *Expression {
	return &Expression{Type: AND, Parts: parts}
}

func IsNull(field string) *Expression {
	return &Expression{Type: ISNULL, Parts: []interface{}{field}}
}

func IsNotNull(field string) *Expression {
	return &Expression{Type: ISNOTNULL, Parts: []interface{}{field}}
}

func Like(field string, value interface{}) *Expression {
	return &Expression{Type: LIKE, Parts: []interface{}{field, value}}
}

func NotLike(field string, value interface{}) *Expression {
	return &Expression{Type: NOTLIKE, Parts: []interface{}{field, value}}
}

func In(field string, values ...interface{}) *Expression {
	return &Expression{Type: IN, Parts: append([]interface{}{field}, values...)}
}

func NotIn(field string, values ...interface{}) *Expression {
	return &Expression{Type: NOTIN, Parts: append([]interface{}{field}, values...)}
}
