// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"strings"
)

type SQLDriver interface {
	// Define returns a given driver version of the provided
	// ANSI type Definition for Creation and a bool indicating
	// the result indicating if said type was defined.
	Define(ANSISQLFieldKind, string) (string, bool)

	// DefineFK returns the definition for a Foreign Key
	// composed with the field name, the foreign table name
	// and the pk/pks of the referenced table.
	DefineFK(string, []string, []string) string

	// DefinePK returns the Primary key definition for the
	// field or fields passed and a boolean indicating if
	// there is a pk.
	DefinePK([]string) (string, bool)
}

var ansiTypes = map[ANSISQLFieldKind]string{
	SqlFK:         "FOREIGN KEY",
	SqlChar:       "CHAR",
	SqlVarchar:    "VARCHAR",
	SqlNchar:      "NCHAR",
	SqlNVarchar:   "NVARCHAR",
	SqlBit:        "BIT",
	SqlBitVarying: "BIT VARYING",
	SqlInt:        "INT",
	SqlSmallInt:   "SMALLINT",
	SqlBigInt:     "BIGINT",
	SqlFloat:      "FLOAT",
	SqlReal:       "REAL",
	SqlDouble:     "DOUBLE",
	SqlNumeric:    "NUMERIC",
	SqlDecimal:    "DECIMAL",
}

// ANSISQLDriver is the reference implementation of SQLDriver
// it provides the ANSI SQL types.
type ANSISQLDriver struct {
}

// customers_services_fk FOREIGN KEY (service_id) REFERENCES services (service_id) ON DELETE CASCADE ON UPDATE CASCADE
const (
	baseCREATE = `CREATE TABLE %s (%s);`
	baseInsert = `INSERT INTO %s (%s) VALUES (%s);`
	baseUpdate = `UPDATE %s SET %s WHERE %s;`

	fkTemplate   = `FOREIGN KEY (%s) REFERENCES %s (%s) ON DELETE CASCADE ON UPDATE CASCADE`
	pkTemplate   = `PRIMARY KEY (%s)`
	baseTemplate = `%s %s`
)

// Type implements SQLDriver.
// TODO(perrito666) support sizes for things like varchar.
func (*ANSISQLDriver) Define(k ANSISQLFieldKind, name string) (string, bool) {
	v, ok := ansiTypes[k]
	if ok {
		v = fmt.Sprintf(baseTemplate, name, v)
	}
	return v, ok
}

// DefineFK implements SQLDriver
func (*ANSISQLDriver) DefineFK(referenceName string, fieldNames, referenceFields []string) string {
	referenceField := strings.Join(referenceFields, ", ")
	localFieldNames := strings.Join(fieldNames, ", ")
	return fmt.Sprintf(fkTemplate, localFieldNames, referenceName, referenceField)
}

// DefinePK implements SQLDriver
func (*ANSISQLDriver) DefinePK(pkFields []string) (string, bool) {
	if len(pkFields) == 0 {
		return "", false
	}
	return fmt.Sprintf(pkTemplate, strings.Join(pkFields, " ,")), true
}

// CraftCreate will take the name of the type, the fields, fks and pks information and
// craft a valid CREATE statement.
func CraftCreate(d SQLDriver, typeName string, fields []FieldDefinition, fks []FKDefinition, pks []string) (string, error) {
	if len(fields) == 0 {
		return "", fmt.Errorf("the table %q has no fields", typeName)
	}
	fieldDefinitions := make([]string, len(fields))
	for i, f := range fields {
		definition, ok := d.Define(f.Type, f.Name)
		if !ok {
			return "", fmt.Errorf("cannot determine an SQL Definition for field %q in the provided driver")
		}
		fieldDefinitions[i] = definition
	}

	fkDefinitions := make([]string, len(fks))
	for i, f := range fks {
		definition := d.DefineFK(f.RemoteTable, f.Names, f.RemoteNames)
		fkDefinitions[i] = definition
	}
	if len(fkDefinitions) != 0 {
		fieldDefinitions = append(fieldDefinitions, fkDefinitions...)
	}

	pkDefinition, ok := d.DefinePK(pks)
	if ok {
		fieldDefinitions = append(fieldDefinitions, pkDefinition)
	}

	return fmt.Sprintf(baseCREATE, typeName, strings.Join(fieldDefinitions, ", ")), nil
}

// CraftInsert will take a FieldsWithValue and returns the corresponding INSERT
// statement.
// TODO(perrito666): Make th Insert template part of the driver?
func CraftInsert(typeName string, fields *FieldsWithValue) string {
	return fmt.Sprintf(baseInsert, typeName, strings.Join(fields.Fields(), ", "), strings.Join(fields.Values(), ", "))
}

// CraftUpdate will take conditions and fields and will craft an update with
// them.
func CraftUpdate(typeName string, conditions, fields *FieldsWithValue) string {
	fieldPairs := fields.Pairs("=")
	conditionalPairs := conditions.Pairs("=")
	return fmt.Sprintf(baseUpdate, typeName, strings.Join(fieldPairs, ", "), strings.Join(conditionalPairs, " AND "))
}
