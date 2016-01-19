// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"reflect"
)

type ANSISQLFieldKind int

const (
	sqlInvalid = iota
	sqlFK

	// Character strings
	sqlChar
	sqlVarchar
	sqlNchar
	sqlNVarchar

	// Bit strings
	sqlBit
	sqlBitVarying

	// Numbers
	sqlInt
	sqlSmallInt
	sqlBigInt
	sqlFloat
	sqlReal
	sqlDouble
	sqlNumeric
	sqlDecimal
)

// tokenizedField holds the name of a struct field and its
// sql type.
type tokenizedField struct {
	name     string
	kind     ANSISQLFieldKind
	isPk     bool
	isUnique bool
	// TODO (perrito666) implement here a way to recursively tokenize for fk
	references *tokenized
}

// tokenized holds a set of fields of a given struct and
// its sql types.
type tokenized struct {
	name   string
	fields []tokenizedField
}

func (t *tokenized) primary() []string {
	primary := []string{}
	for _, f := range t.fields {
		if f.isPk {
			primary = append(primary, f.name)
		}
	}
	return primary
}

// fieldsAndTypes retuns a map of the fields in the tokenized type
// and its SQL types.
// TODO(perrito666) use tags in fields for fk
// TODO(perrito666) fix FK
func (t *tokenized) fieldsAndTypes(d SQLDriver) ([]string, error) {
	fields := make([]string, len(t.fields))
	ansiD := ANSISQLDriver{}
	for i := range t.fields {
		field := t.fields[i]
		// FIXME (perrito666) call fk specific function here with the extra data
		// to keep the basic one clean
		var definition string
		var ok bool
		switch field.kind {
		case sqlFK:
			pk := t.primary()
			// TODO(perrito666) insert an _id field when there is no pk
			if len(pk) == 0 {
				pk = []string{field.name}
			}
			definition = d.DefineFK(field.name, field.references.name, pk)
		default:
			definition, ok = d.Define(field.kind, field.name)
			if !ok {
				definition, ok = ansiD.Define(field.kind, field.name)
			}
			if !ok {
				return nil, fmt.Errorf("cannot determine an SQL Definition for field %q in the provided driver or the ANSI driver")
			}
		}
		fields[i] = definition
	}
	return fields, nil
}

// resolveType tries to map the values on the struct
// to valid ANSI SQL types, for now it is quite rudimentary
// and arbitrary, it also asumes all pointers to be struct ptr.
func resolveType(f reflect.StructField) (ANSISQLFieldKind, bool) {
	t := f.Type
	var sqlType ANSISQLFieldKind
	switch t.Kind() {
	case reflect.Bool:
		sqlType = sqlInt
	case reflect.Int, reflect.Int8:
		sqlType = sqlSmallInt
	case reflect.Int16, reflect.Int32, reflect.Int64:
		sqlType = sqlBigInt
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		sqlType = sqlBigInt
	case reflect.Float32:
		sqlType = sqlFloat
	case reflect.Float64:
		sqlType = sqlDouble
	case reflect.String:
		sqlType = sqlVarchar
	case reflect.Struct, reflect.Ptr:
		sqlType = sqlFK
	default:
		return sqlInvalid, false
	}
	return sqlType, true
}

// tokenize returns a new tokenized struct containing the
// passed struct fields and their sql types.
func tokenize(t reflect.Type) (*tokenized, error) {
	fieldCount := t.NumField()
	fields := make([]tokenizedField, fieldCount)
	for i := 0; i < fieldCount; i++ {
		f := t.Field(i)
		fields[i].name = f.Name
		sqlType, ok := resolveType(f)
		if !ok {
			return nil, fmt.Errorf("cannot resolve SQL equivalent for %q", f.Name)
		}
		if sqlType == sqlFK {
			fieldType := f.Type
			// if it is a ptr we need it dereferenced.
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() != reflect.Struct {
				return nil, fmt.Errorf("expected %v got %v", reflect.Struct, fieldType.Kind())
			}
			fk, err := tokenize(fieldType)
			if err != nil {
				return nil, fmt.Errorf("resolving foreign key for field %q: %v", f.Name, err)
			}
			fields[i].references = fk
		}
		fields[i].kind = sqlType
	}
	return &tokenized{fields: fields, name: t.Name()}, nil
}
