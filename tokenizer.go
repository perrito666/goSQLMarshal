// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"reflect"
	"strings"
)

// ANSISQLFieldKind represents any SQL kind that is currently supported.
type ANSISQLFieldKind int

const (
	SqlInvalid ANSISQLFieldKind = iota
	SqlFK

	// Character strings
	SqlChar
	SqlVarchar
	SqlNchar
	SqlNVarchar

	// Bit strings
	SqlBit
	SqlBitVarying

	// Numbers
	SqlInt
	SqlSmallInt
	SqlBigInt
	SqlFloat
	SqlReal
	SqlDouble
	SqlNumeric
	SqlDecimal
)

// tokenizedField holds the name of a struct field and its
// sql type.
type tokenizedField struct {
	name     string
	kind     ANSISQLFieldKind
	goType   reflect.Type
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

// primary returns a slice of the names for the fields that are
// considered primary keys.
func (t *tokenized) primary() []string {
	primary := []string{}
	for _, f := range t.fields {
		if f.isPk {
			primary = append(primary, f.name)
		}
	}
	return primary
}

// define is a convenience function that returns the SQL definition for the given field
// name with the passed kind using the primary and fallback sql driver or error if its
// not possible to creat the definition.
func define(kind ANSISQLFieldKind, name string, driver, fallback SQLDriver) (string, error) {
	definition, ok := driver.Define(kind, name)
	if !ok {
		definition, ok = fallback.Define(kind, name)
	}
	if !ok {
		return "", fmt.Errorf("cannot determine an SQL Definition for field %q in the provided driver or the Fallback driver")
	}
	return definition, nil

}

// fieldKind returns the SQL kind for the given field and a boolean
// indicating if it was possible to determine it.
func (t *tokenized) fieldKind(name string) (ANSISQLFieldKind, bool) {
	for _, f := range t.fields {
		if f.name == name {
			return f.kind, true
		}
	}
	return SqlInvalid, false
}

type FieldDefinition struct {
	Name string
	Type ANSISQLFieldKind
}

type FKDefinition struct {
	Names       []string
	RemoteNames []string
	RemoteTable string
}

// fieldsAndTypes three slices containing, field definitions, foreign key definitions and list of pks (which
// are also included in the field definitions as a field) or error if it could not gatter the data.
func (t *tokenized) fieldsAndTypes() ([]FieldDefinition, []FKDefinition, []string, error) {
	partialFields := []FieldDefinition{}
	partialFKs := []FKDefinition{}
	for i := range t.fields {
		field := t.fields[i]
		switch field.kind {
		case SqlFK:
			pk := field.references.primary()
			// FIXME: This "invents" an _ID fields which should be inserted
			// automatically in create statements that dont find pks.
			if len(pk) == 0 {
				partialFKs = append(partialFKs,
					FKDefinition{
						RemoteTable: field.references.name,
						Names:       []string{field.name},
						RemoteNames: []string{"_ID"},
					})
				partialFields = append(partialFields,
					FieldDefinition{
						Name: field.name,
						Type: SqlInt,
					})

				continue
			}

			fieldNames := make([]string, len(pk))
			for i := range pk {
				pkName := pk[i]
				name := fmt.Sprintf("%s_%s_fk", field.name, pkName)
				fieldKind, ok := field.references.fieldKind(pkName)
				if !ok {
					return nil, nil, nil, fmt.Errorf("cannot determine the type for referenced pk %q in type %q", pkName, field.references.name)
				}
				fieldNames[i] = name
				partialFields = append(partialFields,
					FieldDefinition{
						Name: name,
						Type: fieldKind,
					})

			}
			partialFKs = append(partialFKs,
				FKDefinition{
					RemoteTable: field.references.name,
					Names:       fieldNames,
					RemoteNames: pk,
				})

			continue

		default:
			partialFields = append(partialFields,
				FieldDefinition{
					Name: field.name,
					Type: field.kind,
				})
		}
	}
	return partialFields, partialFKs, t.primary(), nil
}

// primaryFieldsAndValuess returns two slices with the fields and values for primary keys
// of this tokenized type using "remote" value which should be an instance of the
// same.
// TODO(perrito666) add a type check
func (t *tokenized) primaryFieldsAndValuess(name string, remote reflect.Value) (*FieldsWithValue, error) {
	pks := t.primary()
	fields := NewFieldsWithValue()
	for i := range pks {
		current := pks[i]

		value := remote.FieldByName(current)
		s, ok := valueStringer(value)
		if !ok {
			return nil, fmt.Errorf("cannot determine primary key values, failed on %q", current)
		}

		fields.Add(FieldWithValue{
			Name:  fmt.Sprintf("%s_%s_fk", name, current),
			Value: s,
		})
	}
	return fields, nil

}

// valueStringer tries to return a string representing the value
// of the passed reflect.Value and a boolean indicating if it
// was possible.
func valueStringer(value reflect.Value) (string, bool) {
	var stringValue string
	switch value.Kind() {
	case reflect.Bool:
		v := value.Bool()
		if v {
			stringValue = "1"
		} else {
			stringValue = "0"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := value.Int()
		stringValue = fmt.Sprintf("%d", v)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := value.Uint()
		stringValue = fmt.Sprintf("%d", v)
	case reflect.Float32, reflect.Float64:
		v := value.Float()
		stringValue = fmt.Sprintf("%f", v)
	case reflect.String:
		stringValue = fmt.Sprintf("%q", value.String())
	default:
		return "", false
	}
	return stringValue, true

}

// fieldsAndValues returns two slices representing the fields in the passed interface
// and its values, all in strings or errors if it was not possible to determine them.
// The passed object should be of the same type as the tokenized.
// TODO(perrito666) add a type check for the interface.
func (t *tokenized) fieldsAndValues(in interface{}) (*FieldsWithValue, error) {
	fields := NewFieldsWithValue()
	concreteElem := reflect.ValueOf(in)
	for i := range t.fields {
		current := t.fields[i]
		value := concreteElem.FieldByName(current.name)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() == reflect.Struct {
			f, err := current.references.primaryFieldsAndValuess(current.name, value)
			if err != nil {
				return nil, fmt.Errorf("crafting foreign key: %v", err)
			}
			fields.Append(f)
			continue
		}
		stringValue, ok := valueStringer(value)
		if !ok {
			continue
		}
		fields.Add(FieldWithValue{
			Name:  current.name,
			Value: stringValue,
		})

	}
	return fields, nil
}

// pksFieldsAndValues returns fieldsAndValues result separated in pks and fields.
// FIXME(perrito666) there is some repetition between here and fieldsAndValues
// perhaps I could reverse the order and get fieldsAndValues to just get
// both separated from this and append them.
func (t *tokenized) pksFieldsAndValues(in interface{}) (*FieldsWithValue, *FieldsWithValue, error) {
	f, err := t.fieldsAndValues(in)
	if err != nil {
		return nil, nil, fmt.Errorf("determining fields and values: %v", err)
	}
	pks := t.primary()
	p := NewFieldsWithValue()
	for _, k := range pks {
		pf, ok := f.Pop(k)
		if ok {
			p.Add(pf)
		}
	}
	return p, f, nil
}

// resolveType tries to map the values on the struct
// to valid ANSI SQL types, for now it is quite rudimentary
// and arbitrary, it also asumes all pointers to be struct ptr.
func resolveType(f reflect.StructField) (ANSISQLFieldKind, bool) {
	t := f.Type
	var sqlType ANSISQLFieldKind
	switch t.Kind() {
	case reflect.Bool:
		sqlType = SqlInt
	case reflect.Int, reflect.Int8:
		sqlType = SqlSmallInt
	case reflect.Int16, reflect.Int32, reflect.Int64:
		sqlType = SqlBigInt
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		sqlType = SqlBigInt
	case reflect.Float32:
		sqlType = SqlFloat
	case reflect.Float64:
		sqlType = SqlDouble
	case reflect.String:
		sqlType = SqlVarchar
	case reflect.Struct, reflect.Ptr:
		sqlType = SqlFK
	default:
		return SqlInvalid, false
	}
	return sqlType, true
}

const (
	tagPrimary = "primary"
	tagUnique  = "unique"
)

// processTags is a convenience method that checks if
// the passed tag has sql information.
func (f *tokenizedField) processTags(tag reflect.StructTag) {
	tagstring := tag.Get("sql")
	tags := strings.Split(tagstring, ",")
	for _, t := range tags {
		switch t {
		case tagPrimary:
			f.isPk = true
		case tagUnique:
			f.isUnique = true
		}
	}

}

// tokenize returns a new tokenized struct containing the
// passed struct fields and their sql types.
func tokenize(t reflect.Type) (*tokenized, error) {
	fieldCount := t.NumField()
	fields := make([]tokenizedField, fieldCount)
	for i := 0; i < fieldCount; i++ {
		f := t.Field(i)
		fields[i].name = f.Name
		fields[i].goType = f.Type
		sqlType, ok := resolveType(f)
		if !ok {
			return nil, fmt.Errorf("cannot resolve SQL equivalent for %q", f.Name)
		}
		fields[i].processTags(f.Tag)
		if sqlType == SqlFK {
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
