package db

import "fmt"

var (
	// ErrUnsupportedMethod is yied when a method was called and is not supported by the
	// underlying driver
	ErrUnsupportedMethod = fmt.Errorf("Error this method is not supported by the current driver.")
)
