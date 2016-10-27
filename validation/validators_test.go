//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.
//
//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package validator_test

import (
	"fmt"

	"github.com/mparaiso/tiger-go-framework/validator"
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
