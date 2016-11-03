//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//      http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package logger

import (
	"fmt"
	"log"
	"testing"
)

const (
	// Debug is the level 0 for a logger
	Debug int = iota
	// Info is the level 1 for a logger
	Info
	// Warning is the level 2 for a logger
	Warning
	// Error is the level 3 for a logger
	Error
	// Critical i the level 4 for a logger
	Critical
)

// ToString returns a string representation of the level
func ToString(level int) string {
	switch level {
	case Debug:
		return "Debug"
	case Info:
		return "Info"
	case Warning:
		return "Warning"
	case Error:
		return "Error"
	case Critical:
		return "Critical"
	default:
		return fmt.Sprint(level)
	}
}

// Provider provides a Logger to a container
type Provider interface {
	GetLogger() (Logger, error)
	MustGetLogger() Logger
}

// Logger is a logger
type Logger interface {
	Log(level int, args ...interface{})
	LogF(level int, format string, args ...interface{})
}

// DefaultLogger is the default implementation of Logger
type DefaultLogger struct {
	aLogger *log.Logger
}

// NewDefaultLogger returns a Logger
func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

// NewDefaultLoggerWith returns a Logger with a *log.Logger as argument
func NewDefaultLoggerWith(aLogger *log.Logger) Logger {
	return &DefaultLogger{aLogger}
}

// Log logs a messages
func (logger *DefaultLogger) Log(level int, args ...interface{}) {
	if logger.aLogger != nil {
		logger.aLogger.Print(append([]interface{}{ToString(level)}, args...)...)
		return
	}
	log.Print(append([]interface{}{ToString(level)}, args...)...)
}

// LogF logs a message with a format
func (logger *DefaultLogger) LogF(level int, format string, args ...interface{}) {
	if logger.aLogger != nil {
		logger.aLogger.Print(append([]interface{}{ToString(level)}, fmt.Sprintf(format, args...))...)
		return
	}
	log.Print(append([]interface{}{ToString(level)}, fmt.Sprintf(format, args...))...)
}

// TestLogger is a logger used during tests
type TestLogger struct {
	t *testing.T
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) Logger {
	return &TestLogger{t}
}

// Log logs a messages
func (logger *TestLogger) Log(level int, args ...interface{}) {
	logger.t.Log(append([]interface{}{ToString(level)}, args...))
}

// LogF logs a message with a format
func (logger *TestLogger) LogF(level int, format string, args ...interface{}) {
	logger.t.Log(append([]interface{}{ToString(level)}, fmt.Sprintf(format, args...))...)
}
