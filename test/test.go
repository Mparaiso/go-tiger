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

package test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Mparaiso/go-tiger/logger"
)

// Fatal is a helper used to reduce the boilerplate during tests
func Fatal(t *testing.T, got, want interface{}, comments ...string) {
	var comment string
	if want != got {
		if len(comments) > 0 {
			comment = comments[0]

		} else {
			comment = "Expect"
		}
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf(fmt.Sprintf("Expect\r%s:%d:\r\t%s : %s", filepath.Base(file), line, comment, "want '%v' got '%v'."), want, got)
	}
}

// Error is a helper used to reduce the boilerplate during tests
func Error(t *testing.T, got, want interface{}, comments ...string) {
	var comment string
	if want != got {
		if len(comments) > 0 {
			comment = comments[0]

		} else {
			comment = "Expect"
		}
		_, file, line, _ := runtime.Caller(1)
		t.Errorf(fmt.Sprintf("Expect\r%s:%d:\r\t%s : %s", filepath.Base(file), line, comment, "want '%v' got '%v'."), want, got)
	}

}

var _ logger.Logger = &TestLogger{}

type TestLogger struct {
	t *testing.T
}

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t}
}

func (t TestLogger) Log(level int, arguments ...interface{}) {
	t.t.Log(append([]interface{}{logger.ToString(level)}, arguments...)...)
}
func (t TestLogger) LogF(level int, format string, arguments ...interface{}) {
	t.t.Logf("%d - "+format, append([]interface{}{logger.ToString(level)}, arguments...)...)
}
