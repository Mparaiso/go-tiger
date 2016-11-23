/*
Package tag provides a tag parser for struct tags.
This is the grammar for a single struct tag :

	<tag>           ::= <definition> { ";" <definition> }
	<definition>    ::= <id>
	<definition>    ::= <id> ":" <value>
	<definition>    ::= <id>  "(" <parameter> { "," <parameter> } ")"
	<parameter> 	::= <key> ":" <value>
	<key> 			::= string { string | "_" | digit }
	<value>         ::= string { string | "_" | digit }
	<value>         ::= digit

You can either use the parser on the whole struct tag or on a single key of the struct tag.
See the example for further informations.

*/
package tag

import (
	"fmt"
	"io"
	"text/scanner"

	"github.com/Mparaiso/go-tiger/logger"
)

// Parser parses the content of a single struct tag
type Parser interface {
	SetLogger(logger.Logger)
	Parse() ([]*Definition, error)
}

type defaultParser struct {
	reader  io.Reader
	logger  logger.Logger
	scanner scanner.Scanner
}

// NewParser creates a *Parser value
func NewParser(reader io.Reader) Parser {
	return &defaultParser{reader: reader}
}

// SetLogger enables logging for the parser
func (parser *defaultParser) SetLogger(logger logger.Logger) {
	parser.logger = logger
}
func (parser defaultParser) log(level int, messages ...interface{}) {
	if parser.logger != nil {
		parser.logger.Log(level, messages...)
	}
}

// Definition is either a complex or a simple definition.
type Definition struct {
	Name       string
	Value      string
	Parameters []Parameter
}

// IsSimple returns true if the definition doesn't have parameters
func (d Definition) IsSimple() bool {
	return len(d.Parameters) == 0
}
func (d Definition) String() string {
	return fmt.Sprintf("[ Name '%s' , Value '%s' , Params '%+v' ]", d.Name, d.Value, d.Parameters)
}

// Parameter is a key,value pair
type Parameter struct {
	Key, Value string
}

// Parse parses a tag
func (parser *defaultParser) Parse() (definitions []*Definition, err error) {

	parser.scanner = scanner.Scanner{}
	parser.scanner.Init(parser.reader)
	// while not the end of file
	for token := parser.scanner.Scan(); token != scanner.EOF; token = parser.scanner.Scan() {
		parser.log(logger.Debug, token, string(token))
		// if definition separator continue
		if token == ';' {
			parser.log(logger.Debug, "found ; continue")
			continue
		} else if token == scanner.Ident {
			// if identifier , add a definition with identifier as name
			parser.log(logger.Debug, "found identifer ", parser.scanner.TokenText())
			definition := &Definition{Name: parser.scanner.TokenText()}
			definitions = append(definitions, definition)
			if token = parser.scanner.Scan(); token != ':' && token != '(' && token != ';' && token != scanner.EOF {
				parser.log(logger.Error, "found", string(token))
				return nil, parser.errorUnexpectedToken()
			} else if token == ':' {
				// found Value for a simple definition
				parser.log(logger.Debug, "found :")
				if token = parser.scanner.Scan(); token != scanner.Ident && token != scanner.Int {
					return nil, parser.errorUnexpectedToken()
				}

				definition.Value = parser.scanner.TokenText()
				parser.log(logger.Debug, "found value ", definition.Value)
				if token = parser.scanner.Peek(); token != scanner.EOF && token != ';' {
					parser.scanner.Scan()
					return nil, parser.errorUnexpectedToken()
				}
				continue
			} else if token == '(' {
				// found a complex definition
				parser.log(logger.Debug, "found complex field", string(token))
				token = parser.scanner.Scan()
				for token != ')' {
					if token != scanner.Ident && token != scanner.String {
						parser.log(logger.Error, "no identity found ", string(token))
						return nil, parser.errorUnexpectedToken()
					}
					// found the key
					key := parser.scanner.TokenText()
					parser.log(logger.Debug, "key", key)
					if token = parser.scanner.Scan(); token != ':' {
						return nil, parser.errorUnexpectedToken()
					}
					if token = parser.scanner.Scan(); token != scanner.Ident && token != scanner.Int {
						return nil, parser.errorUnexpectedToken()
					}
					// found the value
					value := parser.scanner.TokenText()
					parser.log(logger.Debug, "value", value)
					definition.Parameters = append(definition.Parameters, Parameter{Key: key, Value: value})
					if token = parser.scanner.Scan(); token != ',' && token != ')' {
						return nil, parser.errorUnexpectedToken()
					} else if token == ',' {
						// there should be another Parameter ahead
						token = parser.scanner.Scan()
					}
				}

			}
		} else {
			return nil, parser.errorUnexpectedToken()
		}
	}
	return
}
func (parser *defaultParser) errorUnexpectedToken() error {
	parser.log(logger.Error, "found", parser.scanner.TokenText())
	return fmt.Errorf("Error unexpected token '%s' at position %d in ", parser.scanner.TokenText(), parser.scanner.Pos().Column)
}
