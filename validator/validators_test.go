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

package validator_test

import (
	"fmt"

	"github.com/mparaiso/go-tiger/validator"
)

func ExampleIsURL() {
	for _, url := range []string{
		"https://active-object./introduction.com",
		"https://at.baz.co.uk/foo.com/?&bar=booo",
		"http://baz.com/bar?id=bizz",
		"http://presentation.opex.com/index.html?foobar=biz#baz",
	} {
		fmt.Println(validator.IsURL(url))
	}

	for _, url := range []string{
		"at.baz.co.uk/foo.com/?&bar=booo",
		"foo.com",
		"foo",
		"biz/baz",
		"something.com/ with space",
	} {
		fmt.Println(validator.IsURL(url))
	}

	// Output:
	// true
	// true
	// true
	// true
	// false
	// false
	// false
	// false
	// false
}
