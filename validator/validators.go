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

package validator

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// Error is a validation error
type Error interface {
	HasErrors() bool
	Append(key, value string)
	Error() string
}

// ConcreteError holds errors in a map
type ConcreteError map[string][]string

func NewConcreteError() *ConcreteError {
	c := ConcreteError(map[string][]string{})
	return &c
}

// Append adds an new error to a map
func (v ConcreteError) Append(key string, value string) {
	v[key] = append(v[key], value)
}

func (v *ConcreteError) GetErrors() map[string][]string {
	return *v
}

func (v ConcreteError) Error() string {
	return fmt.Sprintf("%#v", v)
}

// HasErrors returns true if error exists
func (v ConcreteError) HasErrors() bool {
	return len(v) != 0
}

// MarshalXML marshalls a ConcreteError
func (v ConcreteError) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Errors struct {
		Field []struct {
			Name  string
			Error []string
		}
	}
	errors := Errors{[]struct {
		Name  string
		Error []string
	}{}}
	for key, value := range v {
		errors.Field = append(errors.Field, struct {
			Name  string
			Error []string
		}{key, value})
	}
	return e.EncodeElement(errors, start)
}

// StringNotEmptyValidator checks if a string is empty
func StringNotEmptyValidator(field string, value string, errors Error) {
	if StringEmpty(value) {
		errors.Append(field, "should not be empty")
	}
}

// StringEmpty returns true if the string is empty
func StringEmpty(value string) bool {
	return strings.Trim(value, "\t\r ") == ""
}

// StringMinLengthValidator validates a string by minimum length
func StringMinLengthValidator(field, value string, minlength int, errors Error) {
	if len(value) < minlength {
		errors.Append(field, fmt.Sprintf("should be at least %d character long", minlength))
	}
}

// StringMaxLengthValidator validates a string by maximum length
func StringMaxLengthValidator(field, value string, maxlength int, errors Error) {
	if len(value) > maxlength {
		errors.Append(field, "should be at most  %d character long")
	}
}

// StringLengthValidator validates a string by minimum and maxium length
func StringLengthValidator(field, value string, minLength int, maxLength int, errors Error) {
	StringMinLengthValidator(field, value, minLength, errors)
	StringMaxLengthValidator(field, value, maxLength, errors)
}

// MatchValidator validates a string by an expected value
func MatchValidator(field1 string, field2 string, value1, value2 interface{}, errors Error) {
	if value1 != value2 {
		errors.Append(field1, fmt.Sprintf("should match %s ", field2))
	}
}

// EmailValidator validates an email
func EmailValidator(field, value string, errors Error) {
	if !isEmail(value) {
		errors.Append(field, "should be a valid email")
	}
}

// URLValidator validates a URL
func URLValidator(field, value string, errors Error) {
	if !IsURL(value) {
		errors.Append(field, "should be a valid URL and begin with http:// or https:// ")
	}
}

// PatternValidator valides a value according to a regexp pattern
func PatternValidator(field, value string, pattern *regexp.Regexp, errors Error) {
	if !pattern.MatchString(value) {
		errors.Append(field, "should match the following pattern : "+pattern.String())
	}
}

// IsURL returns true if is url
func IsURL(candidate string) bool {
	return regexp.MustCompile(`^(https?\:\/\/)(\S+\.)?\S+\.\S+(\.\S+)?\/?\S+$`).MatchString(candidate)
}

// IsEmail returns true if is email
func isEmail(candidate string) bool {
	return regexp.MustCompile(`\S+@\S+\.\S+`).MatchString(candidate)
}
