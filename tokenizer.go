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
	name string
	kind ANSISQLFieldKind
}

// tokenized holds a set of fields of a given struct and
// its sql types.
type tokenized struct {
	fields []tokenizedField
}

// fieldsAndTypes retuns a map of the fields in the tokenized type
// and its SQL types.
// TODO(perrito666) use tags in fields for fk
// TODO(perrito666) fix FK
func (t *tokenized) fieldsAndTypes(d SQLDriver) (map[string]string, error) {
	fields := make(map[string]string, len(t.fields))
	ansiD := ANSISQLDriver{}
	for i := range t.fields {
		field := t.fields[i]
		kind, ok := d.Type(field.kind)
		if !ok {
			kind, ok = ansiD.Type(field.kind)
		}
		if !ok {
			return nil, fmt.Errorf("cannot determine an SQL name for field %q in the provided driver or the ANSI driver")
		}
		fields[field.name] = kind
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
		fields[i].kind = sqlType
	}
	return &tokenized{fields: fields}, nil
}
