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

// ValidationError allows to collect
// multiple errors from different fields (in a form for instance)
// and get them through a map[string][]string to be displayed in
// an html page or a API response.
type ValidationError interface {
	HasErrors() bool
	Append(key, value string)
	GetErrors() map[string][]string
	ReturnNilOrErrors() ValidationError
	Error() string
	MarshalXML(e *xml.Encoder, start xml.StartElement) error
}

type concreteValidationError struct {
	errors map[string][]string
}

// NewValidationError returns a ValidationErron
func NewValidationError() ValidationError {
	return &concreteValidationError{errors: map[string][]string{}}
}

// Append adds an error to a map
func (validationError *concreteValidationError) Append(field string, value string) {
	validationError.errors[field] = append(validationError.errors[field], value)
}

// GetErrors gets all errors as a map
func (validationError *concreteValidationError) GetErrors() map[string][]string {
	return validationError.errors
}

// ReturnNilOrErrors is an helper that will return nil if there is no errors
// useful when returning a Error interface from validation
func (validationError *concreteValidationError) ReturnNilOrErrors() ValidationError {
	if validationError.HasErrors() {
		return validationError
	}
	return nil
}

func (validationError concreteValidationError) Error() string {
	return fmt.Sprintf("%#v", validationError.errors)
}

// HasErrors returns true if error exists
func (validationError concreteValidationError) HasErrors() bool {
	return len(validationError.errors) != 0
}

// MarshalXML marshalls a ConcreteError
func (validationError concreteValidationError) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Error struct {
		Name  string
		Error []string
	}

	type Errors struct {
		Error []Error
	}

	errors := Errors{}
	for key, value := range validationError.errors {
		errors.Error = append(errors.Error, Error{key, value})
	}
	return e.EncodeElement(errors, start)
}

// StringNotEmptyValidator checks if a string is empty
func StringNotEmptyValidator(field string, value string, errors ValidationError) {
	if StringEmpty(value) {
		errors.Append(field, "should not be empty")
	}
}

// StringEmpty returns true if the string is empty
func StringEmpty(value string) bool {
	return strings.Trim(value, "\t\r ") == ""
}

// StringMinLengthValidator validates a string by minimum length
func StringMinLengthValidator(field, value string, minlength int, errors ValidationError) {
	if len(value) < minlength {
		errors.Append(field, fmt.Sprintf("should be at least %d character long", minlength))
	}
}

// StringMaxLengthValidator validates a string by maximum length
func StringMaxLengthValidator(field, value string, maxlength int, errors ValidationError) {
	if len(value) > maxlength {
		errors.Append(field, "should be at most  %d character long")
	}
}

// StringLengthValidator validates a string by minimum and maxium length
func StringLengthValidator(field, value string, minLength int, maxLength int, errors ValidationError) {
	StringMinLengthValidator(field, value, minLength, errors)
	StringMaxLengthValidator(field, value, maxLength, errors)
}

// MatchValidator validates a string by an expected value
func MatchValidator(field1 string, field2 string, value1, value2 interface{}, errors ValidationError) {
	if value1 != value2 {
		errors.Append(field1, fmt.Sprintf("should match %s ", field2))
	}
}

// EmailValidator validates an email
func EmailValidator(field, value string, errors ValidationError) {
	if !isEmail(value) {
		errors.Append(field, "should be a valid email")
	}
}

// URLValidator validates a URL
func URLValidator(field, value string, errors ValidationError) {
	if !IsURL(value) {
		errors.Append(field, "should be a valid URL and begin with http:// or https:// ")
	}
}

// PatternValidator valides a value according to a regexp pattern
func PatternValidator(field, value string, pattern *regexp.Regexp, errors ValidationError) {
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
