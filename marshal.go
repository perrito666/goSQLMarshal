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

const baseCREATE = `CREATE TABLE %s %s;`

// Create returns a SQL Create Statement for the passed object
// if the type passed is different from the one of the marshaller
// it returns an error.
func (s *SQLMarshaller) Create(driver SQLDriver) (string, error) {
	fields, err := s.tokenized.fieldsAndTypes(driver)
	if err != nil {
		return "", fmt.Errorf("crafting the fields for CREATE statement: %v", err)
	}
	createFields := make([]string, len(fields))
	i := 0
	for k, v := range fields {
		createFields[i] = fmt.Sprintf("%s %s", k, v)
		i++
	}
	return fmt.Sprintf(baseCREATE, s.typeOf.Name(), strings.Join(createFields, ",")), nil
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
