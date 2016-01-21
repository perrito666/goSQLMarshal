// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"testing"
)

type dumbFK struct {
	aField       int `sql:"primary"`
	anotherField string
}

type dumbStruct struct {
	testInt    int `sql:"primary"`
	testString string
	testFloat  float32
	testPtr    *dumbFK
	testStruct dumbFK
}

type dumbFKMulti struct {
	aField       int `sql:"primary"`
	aField2      int `sql:"primary"`
	anotherField string
}

type dumbStructMulti struct {
	testInt    int `sql:"primary"`
	testString string
	testFloat  float32
	testPtr    *dumbFKMulti
	testStruct dumbFKMulti
}

type untaggedDumbFK struct {
	aField       int
	anotherField string
}

type untaggedDumbStruct struct {
	testInt    int
	testString string
	testFloat  float32
	testPtr    *untaggedDumbFK
	testStruct untaggedDumbFK
}

func (*dumbStruct) methodsAreIgnored() {
}

func TestCreateTagged(t *testing.T) {
	d := dumbStruct{
		testInt:    1,
		testString: "some velvet string",
		testFloat:  2.0,
		testPtr: &dumbFK{
			aField:       3,
			anotherField: "another string",
		},
		testStruct: dumbFK{
			aField:       4,
			anotherField: "another non struct field",
		},
	}
	m, err := NewTypeSQLMarshaller(d)
	if err != nil {
		t.Errorf("cannot create marshaler: %v", err)
	}
	dr := &ANSISQLDriver{}
	c, err := m.Create(dr)
	if err != nil {
		t.Errorf("cannot marshall to CREATE statement: %v", err)
	}
	t.Log(c)
	expectedSQL := `CREATE TABLE dumbStruct (testInt SMALLINT, testString VARCHAR, testFloat FLOAT, testPtr_aField_fk SMALLINT, testStruct_aField_fk SMALLINT, FOREIGN KEY (testPtr_aField_fk) REFERENCES dumbFK (aField) ON DELETE CASCADE ON UPDATE CASCADE, FOREIGN KEY (testStruct_aField_fk) REFERENCES dumbFK (aField) ON DELETE CASCADE ON UPDATE CASCADE, PRIMARY KEY (testInt));`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

	c, err = m.Insert(dr, d)
	if err != nil {
		t.Errorf("cannot marshall to INSERT statement: %v", err)
	}
	t.Log(c)
	expectedSQL = `INSERT INTO dumbStruct (testInt, testString, testFloat, testPtr_aField_fk, testStruct_aField_fk) VALUES (1, "some velvet string", 2.000000, 3, 4);`
	if c != expectedSQL {
		t.Errorf("unexpected INSERT statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

func TestCreateTaggedMultiPK(t *testing.T) {
	d := dumbStructMulti{
		testInt:    1,
		testString: "some velvet string",
		testFloat:  2.0,
		testPtr: &dumbFKMulti{
			aField:       3,
			aField2:      5,
			anotherField: "another string",
		},
		testStruct: dumbFKMulti{
			aField:       4,
			aField2:      6,
			anotherField: "another non struct field",
		},
	}
	m, err := NewTypeSQLMarshaller(d)
	if err != nil {
		t.Errorf("cannot create marshaler: %v", err)
	}
	dr := &ANSISQLDriver{}
	c, err := m.Create(dr)
	if err != nil {
		t.Errorf("cannot marshall to CREATE statement: %v", err)
	}
	t.Log(c)
	expectedSQL := `CREATE TABLE dumbStructMulti (testInt SMALLINT, testString VARCHAR, testFloat FLOAT, testPtr_aField_fk SMALLINT, testPtr_aField2_fk SMALLINT, testStruct_aField_fk SMALLINT, testStruct_aField2_fk SMALLINT, FOREIGN KEY (testPtr_aField_fk, testPtr_aField2_fk) REFERENCES dumbFKMulti (aField, aField2) ON DELETE CASCADE ON UPDATE CASCADE, FOREIGN KEY (testStruct_aField_fk, testStruct_aField2_fk) REFERENCES dumbFKMulti (aField, aField2) ON DELETE CASCADE ON UPDATE CASCADE, PRIMARY KEY (testInt));`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

func TestCreateUnTagged(t *testing.T) {
	d := untaggedDumbStruct{
		testInt:    1,
		testString: "some velvet string",
		testFloat:  2.0,
		testPtr: &untaggedDumbFK{
			aField:       3,
			anotherField: "another string",
		},
		testStruct: untaggedDumbFK{
			aField:       4,
			anotherField: "another non struct field",
		},
	}
	m, err := NewTypeSQLMarshaller(d)
	if err != nil {
		t.Errorf("cannot create marshaler: %v", err)
	}
	dr := &ANSISQLDriver{}
	c, err := m.Create(dr)
	if err != nil {
		t.Errorf("cannot marshall to CREATE statement: %v", err)
	}
	t.Log(c)
	expectedSQL := `CREATE TABLE untaggedDumbStruct (testInt SMALLINT, testString VARCHAR, testFloat FLOAT, testPtr INT, testStruct INT, FOREIGN KEY (testPtr) REFERENCES untaggedDumbFK (_ID) ON DELETE CASCADE ON UPDATE CASCADE, FOREIGN KEY (testStruct) REFERENCES untaggedDumbFK (_ID) ON DELETE CASCADE ON UPDATE CASCADE);`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

// DOC Sample
type Sample struct {
	ID                int `sql:"primary"`
	Name              string
	Reference         *Reference
	ConcreteReference Reference
}

type Reference struct {
	DifferentNameID int `sql:"primary"`
	Name            string
}

func doSQLCreate() (string, error) {
	sample := Sample{
		ID:   1,
		Name: "a sample name",
		Reference: &Reference{
			DifferentNameID: 1,
			Name:            "a reference name",
		},
		ConcreteReference: Reference{
			DifferentNameID: 2,
			Name:            "another reference name",
		},
	}

	m, err := NewTypeSQLMarshaller(sample)
	if err != nil {
		return "", fmt.Errorf("cannot create marshaler: %v", err)
	}

	dr := &ANSISQLDriver{}

	c, err := m.Create(dr)
	if err != nil {
		return "", fmt.Errorf("cannot marshall to CREATE statement: %v", err)
	}
	return c, nil
}

func TestDocSampleCreate(t *testing.T) {
	c, err := doSQLCreate()
	if err != nil {
		t.Errorf("could not run documentation sample for CREATE: %v", err)
	}
	t.Log(c)
	expectedSQL := `CREATE TABLE Sample (ID SMALLINT, Name VARCHAR, Reference_DifferentNameID_fk SMALLINT, ConcreteReference_DifferentNameID_fk SMALLINT, FOREIGN KEY (Reference_DifferentNameID_fk) REFERENCES Reference (DifferentNameID) ON DELETE CASCADE ON UPDATE CASCADE, FOREIGN KEY (ConcreteReference_DifferentNameID_fk) REFERENCES Reference (DifferentNameID) ON DELETE CASCADE ON UPDATE CASCADE, PRIMARY KEY (ID));`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

type SampleInsert struct {
	ID                int `sql:"primary"`
	Name              string
	Reference         *ReferenceInsert
	ConcreteReference Reference
}

type ReferenceInsert struct {
	DifferentNameID int `sql:"primary"`
	AnExtraID       int `sql:"primary"`
	Name            string
}

func doSQLInsert() (string, error) {
	sample := SampleInsert{
		ID:   1,
		Name: "a sample name",
		Reference: &ReferenceInsert{
			DifferentNameID: 1,
			AnExtraID:       3,
			Name:            "a reference name",
		},
		ConcreteReference: Reference{
			DifferentNameID: 2,
			Name:            "another reference name",
		},
	}

	m, err := NewTypeSQLMarshaller(sample)
	if err != nil {
		return "", fmt.Errorf("cannot create marshaler: %v", err)
	}

	dr := &ANSISQLDriver{}

	c, err := m.Insert(dr, sample)
	if err != nil {
		return "", fmt.Errorf("cannot marshall to INSERT statement: %v", err)
	}
	return c, nil
}

func TestDocSampleInsert(t *testing.T) {
	c, err := doSQLInsert()
	if err != nil {
		t.Errorf("could not run documentation sample for INSERT: %v", err)
	}
	t.Log(c)
	expectedSQL := `INSERT INTO SampleInsert (ID, Name, Reference_DifferentNameID_fk, Reference_AnExtraID_fk, ConcreteReference_DifferentNameID_fk) VALUES (1, "a sample name", 1, 3, 2);`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}
