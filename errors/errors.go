//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

// Package errors provides helper that makes working
// with Go errors easier.
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// New returns an error that formats as the given text.
func New(message string) error {
	return errors.New(message)
}

// Wrap wraps an error into another error in order to track
// the source of the error
func Wrap(err error, messages ...string) ErrorWrapper {
	if err == nil {
		err = New("error")
	}
	_, file, line, _ := runtime.Caller(1)
	message := strings.Join(messages, "")
	return defaultErrorWrapper{file, line, message, err}
}

type defaultErrorWrapper struct {
	file          string
	line          int
	message       string
	originalError error
}

func (errorWrapper defaultErrorWrapper) Error() string {
	return fmt.Sprintf("%s:%d %s\r%s", errorWrapper.file, errorWrapper.line, errorWrapper.message, errorWrapper.originalError.Error())
}

func (errorWrapper defaultErrorWrapper) Original() error {
	if err, ok := errorWrapper.originalError.(ErrorWrapper); ok {
		return err.Original()
	}
	return errorWrapper.originalError
}

// ErrorWrapper helps tracking errors
// by wrapping them with additional informations such
// as the line and the file where the error was yield
type ErrorWrapper interface {

	// Error returns the error message with file and line info
	Error() string

	// Original unwrap all wrapped errors and returns the original error
	Original() error
}
