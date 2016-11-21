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

###Example 

	// Let's parse the meta tag in Title field of type Book
	type Book struct {
		Title string `meta:"isbn:11010303030;published;author(firstname:john,lastname:doe);editor:apress"`
	}
	// let's get the tag
	titleField, _ := reflect.TypeOf(Book{}).FieldByName("Title")
	metaTag := titleField.Tag.Get("meta")
	// let's create a parser
	parser := tag.NewParser(strings.NewReader(metaTag))
	// do parse
	definitions, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(definitions)
	// Output:
	// [[ Name 'isbn' , Value '11010303030' , Params '[]' ] [ Name 'published' , Value '' , Params '[]' ] [ Name 'author' , Value '' , Params '[{Key:firstname Value:john} {Key:lastname Value:doe}]' ] [ Name 'editor' , Value 'apress' , Params '[]' ]]




