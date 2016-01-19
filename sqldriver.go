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
	DefineFK(string, string, []string) string
}

var ansiTypes = map[ANSISQLFieldKind]string{
	sqlFK:         "FOREIGN KEY",
	sqlChar:       "CHAR",
	sqlVarchar:    "VARCHAR",
	sqlNchar:      "NCHAR",
	sqlNVarchar:   "NVARCHAR",
	sqlBit:        "BIT",
	sqlBitVarying: "BIT VARYING",
	sqlInt:        "INT",
	sqlSmallInt:   "SMALLINT",
	sqlBigInt:     "BIGINT",
	sqlFloat:      "FLOAT",
	sqlReal:       "REAL",
	sqlDouble:     "DOUBLE",
	sqlNumeric:    "NUMERIC",
	sqlDecimal:    "DECIMAL",
}

// ANSISQLDriver is the reference implementation of SQLDriver
// it provides the ANSI SQL types.
type ANSISQLDriver struct {
}

// customers_services_fk FOREIGN KEY (service_id) REFERENCES services (service_id) ON DELETE CASCADE ON UPDATE CASCADE
const (
	fkTemplate   = `%s FOREIGN KEY (%s) REFERENCES %s (%s) ON DELETE CASCADE ON UPDATE CASCADE`
	baseTempalte = `%s %s`
)

// Type implements SQLDriver.
// TODO(perrito666) support sizes for things like varchar.
func (*ANSISQLDriver) Define(k ANSISQLFieldKind, name string) (string, bool) {
	v, ok := ansiTypes[k]
	if ok {
		v = fmt.Sprintf(baseTempalte, name, v)
	}
	return v, ok
}

func (*ANSISQLDriver) DefineFK(fieldName, referenceName string, referenceFields []string) string {
	referenceField := strings.Join(referenceFields, ", ")
	return fmt.Sprintf(fkTemplate, fieldName, referenceField, referenceName, referenceField)
}
