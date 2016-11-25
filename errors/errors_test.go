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

package errors_test

import (
	"fmt"
	"strings"

	"github.com/Mparaiso/go-tiger/errors"
)

func ExampleErrorWrapper() {
	// Let's wrap an error into a errors.ErrorWrapper that can
	// track where it was yielded.
	err := errors.New("Original Error")
	wrappedError := errors.Wrap(err, "Wrapped Error")
	fmt.Println(strings.Contains(wrappedError.Error(), "Wrapped Error"))  // contains the wrapper message
	fmt.Println(strings.Contains(wrappedError.Error(), "Original Error")) // contains the original message
	fmt.Println(strings.Contains(wrappedError.Error(), "28"))             // contains line
	fmt.Println(strings.Contains(wrappedError.Error(), "errors_test.go")) // contains file name
	fmt.Println(wrappedError.Original().Error())
	// Output:
	// true
	// true
	// true
	// true
	// Original Error
}
