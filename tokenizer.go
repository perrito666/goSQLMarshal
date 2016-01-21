// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"reflect"
	"strings"
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

func (t *tokenized) primary() []string {
	primary := []string{}
	for _, f := range t.fields {
		if f.isPk {
			primary = append(primary, f.name)
		}
	}
	return primary
}

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

func (t *tokenized) fieldKind(name string) (ANSISQLFieldKind, bool) {
	for _, f := range t.fields {
		if f.name == name {
			return f.kind, true
		}
	}
	return sqlInvalid, false
}

// fieldsAndTypes retuns a map of the fields in the tokenized type
// and its SQL types.
func (t *tokenized) fieldsAndTypes(d SQLDriver) ([]string, error) {
	ansiD := ANSISQLDriver{}
	partialFields := []string{}
	partialFKs := []string{}
	for i := range t.fields {
		field := t.fields[i]
		var definition string
		var err error
		switch field.kind {
		case sqlFK:
			pk := field.references.primary()
			// TODO(perrito666) insert an _id field when there is no pk
			// for the moment I make a big BIG assumption
			if len(pk) == 0 {
				partialFKs = append(partialFKs, d.DefineFK(field.references.name, []string{field.name}, []string{"_ID"}))
				definition, err := define(sqlInt, field.name, d, &ansiD)
				if err != nil {
					return nil, fmt.Errorf("trying to define %q foreign key field without knowledge of remote pks: %v", field.name, err)
				}

				partialFields = append(partialFields, definition)
				continue
			}
			fieldNames := make([]string, len(pk))
			for i := range pk {
				pkName := pk[i]
				name := fmt.Sprintf("%s_%s_fk", field.name, pkName)
				fieldKind, ok := field.references.fieldKind(pkName)
				if !ok {
					return nil, fmt.Errorf("cannot determine the type for referenced pk %q in type %q", pkName, field.references.name)
				}
				definition, err := define(fieldKind, name, d, &ansiD)
				if err != nil {
					return nil, fmt.Errorf("trying to define %q foreign key field: %v", field.name, err)
				}
				partialFields = append(partialFields, definition)
				fieldNames[i] = name
			}

			partialFKs = append(partialFKs, d.DefineFK(field.references.name, fieldNames, pk))
			continue
		default:
			definition, err = define(field.kind, field.name, d, &ansiD)
			if err != nil {
				return nil, fmt.Errorf("trying to define %q field: %v", field.name, err)
			}
		}
		partialFields = append(partialFields, definition)
	}
	fields := append(partialFields, partialFKs...)
	pk, ok := d.DefinePK(t.primary())
	if ok {
		fields = append(fields, pk)
	}
	return fields, nil
}

func (t *tokenized) primaryFieldsAndValuess(name string, remote reflect.Value) ([]string, []string, error) {
	pks := t.primary()
	fields := make([]string, len(pks))
	values := make([]string, len(pks))
	for i := range pks {
		current := pks[i]
		fmt.Println(pks)

		value := remote.FieldByName(current)
		s, ok := valueStringer(value)
		if !ok {
			return nil, nil, fmt.Errorf("cannot determine primary key values, failed on %q", current)
		}

		fields[i] = fmt.Sprintf("%s_%s_fk", name, current)
		values[i] = s
	}
	return fields, values, nil

}

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

func (t *tokenized) fieldsAndValues(d SQLDriver, in interface{}) ([]string, []string, error) {
	fields := []string{}
	values := []string{}
	concreteElem := reflect.ValueOf(in)
	for i := range t.fields {
		current := t.fields[i]
		value := concreteElem.FieldByName(current.name)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() == reflect.Struct {
			f, val, err := current.references.primaryFieldsAndValuess(current.name, value)
			if err != nil {
				return nil, nil, fmt.Errorf("crafting foreign key: %v", err)
			}
			fields = append(fields, f...)
			values = append(values, val...)
			continue
		}
		stringValue, ok := valueStringer(value)
		if !ok {
			continue
		}

		fields = append(fields, current.name)
		values = append(values, stringValue)

	}
	return fields, values, nil
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

const (
	tagPrimary = "primary"
	tagUnique  = "unique"
)

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
