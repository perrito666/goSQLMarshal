// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"reflect"
	"strings"
)

// SQLMarshaller is a marshaller for a given type of object.
type SQLMarshaller struct {
	typeOf    reflect.Type
	tokenized *tokenized
}

const (
	baseCREATE = `CREATE TABLE %s (%s);`
	baseInsert = `INSERT INTO %s (%s) VALUES (%s);`
)

// Create returns a SQL CREATE Statement for the type of this marshaller
// or error if it cannot generate it.
// Fields that hold structs or pointers will be considered Foreign Keys
// Only Ptr of Stucts are supported for the moment.
func (s *SQLMarshaller) Create(driver SQLDriver) (string, error) {
	fields, err := s.tokenized.fieldsAndTypes(driver)
	if err != nil {
		return "", fmt.Errorf("crafting the fields for CREATE statement: %v", err)
	}
	return fmt.Sprintf(baseCREATE, s.typeOf.Name(), strings.Join(fields, ", ")), nil
}

// Insert returns a SQL INSERT statements for the passed object or
// error if it cannot process the passed object.
// If there are Fields which are structs or pointers to structs
// it will consider them Foreign Keys up to only one level of
// indirection.
func (s *SQLMarshaller) Insert(driver SQLDriver, in interface{}) (string, error) {
	fields, values, err := s.tokenized.fieldsAndValues(driver, in)
	if err != nil {
		return "", fmt.Errorf("crafting the fields/values for INSERT statement: %v", err)
	}

	if len(fields) != len(values) {
		return "", fmt.Errorf("the amount of fields and values differ %d vs %d", fields, values)
	}
	if len(fields) == 0 {
		return "", fmt.Errorf("could not determine fields and values to insert, the resulting query would be invalid")
	}
	return fmt.Sprintf(baseInsert, s.typeOf.Name(), strings.Join(fields, ", "), strings.Join(values, ", ")), nil
}

// NewTypeSQLMarshaller returns a marshaller for the type of the passed
// object, if it is not a struct it will fail.
func NewTypeSQLMarshaller(in interface{}) (*SQLMarshaller, error) {
	t := reflect.TypeOf(in)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("interface must be a struct")
	}
	tok, err := tokenize(t)
	if err != nil {
		return nil, fmt.Errorf("creating a marshaller: %v", err)
	}
	return &SQLMarshaller{typeOf: t, tokenized: tok}, nil
}
