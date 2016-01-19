// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

type SQLDriver interface {
	// Type returns a given driver version of the provided
	// ANSI type and a bool indicating if said type was
	// defined.
	Type(ANSISQLFieldKind) (string, bool)
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

// Type implements SQLDriver.
// TODO(perrito666) support sizes for things like varchar.
func (*ANSISQLDriver) Type(k ANSISQLFieldKind) (string, bool) {
	v, ok := ansiTypes[k]
	return v, ok
}
