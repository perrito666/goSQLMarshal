// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"reflect"
)

// SQLMarshaller is a marshaller for a given type of object.
type SQLMarshaller struct {
	typeOf    reflect.Type
	tokenized *tokenized
}

// UpdatePK return an update statement for the passed object that
// should update the entry represented by the pk/s on the passed struct
// with the values it has set.
func (s *SQLMarshaller) UpdatePK(in interface{}) (string, error) {
	pks, fields, err := s.tokenized.pksFieldsAndValues(in)
	if err != nil {
		return "", fmt.Errorf("extracting the pks, fields and values: %v", err)
	}
	return CraftUpdate(s.Name(), pks, fields), nil
}

// Name returns the current name of the marshaller based on the type
// if no type is provided, it uses the tokenized name
func (s *SQLMarshaller) Name() string {
	name := s.typeOf.Name()
	if name == "" {
		name = s.tokenized.name
	}
	return name
}

// Create returns a SQL CREATE Statement for the type of this marshaller
// or error if it cannot generate it.
// Fields that hold structs or pointers will be considered Foreign Keys
// Only Ptr of Stucts are supported for the moment.
func (s *SQLMarshaller) Create(driver SQLDriver) (string, error) {
	fields, fks, pks, err := s.tokenized.fieldsAndTypes()
	if err != nil {
		return "", fmt.Errorf("gattering the fields for CREATE statement: %v", err)
	}
	return CraftCreate(driver, s.Name(), fields, fks, pks)
}

// Insert returns a SQL INSERT statements for the passed object or
// error if it cannot process the passed object.
// If there are Fields which are structs or pointers to structs
// it will consider them Foreign Keys up to only one level of
// indirection.
func (s *SQLMarshaller) Insert(in interface{}) (string, error) {
	fields, err := s.tokenized.fieldsAndValues(in)
	if err != nil {
		return "", fmt.Errorf("crafting the fields/values for INSERT statement: %v", err)
	}

	if fields.Len() == 0 {
		return "", fmt.Errorf("could not determine fields and values to insert, the resulting query would be invalid")
	}
	return CraftInsert(s.Name(), fields), nil
}

// NewTypeSQLMarshaller returns a marshaller for the type of the passed
// object, if it is not a struct it will fail.
func NewTypeSQLMarshaller(in interface{}, name string) (*SQLMarshaller, error) {
	t := reflect.TypeOf(in)

	var err error
	var tokens *tokenized

	switch t.Kind() {
	case reflect.Map:
		{
			tokens, err = TokenizeMap(in.(map[interface{}]interface{}), name)
		}
	case reflect.Struct:
		{
			if name == "" {
				name = t.Name()
			}
			tokens, err = TokenizeType(t, name)
		}
	default:
		{
			return nil, fmt.Errorf("Only Map and Struct types are currently supported for marshalling")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("creating a marshaller: %v", err)
	}

	return &SQLMarshaller{typeOf: t, tokenized: tokens}, nil

}
