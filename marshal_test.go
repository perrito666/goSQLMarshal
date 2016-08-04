// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import (
	"fmt"
	"reflect"
	"testing"

	goyaml "gopkg.in/yaml.v2"
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

func TestTagged(t *testing.T) {
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
	m, err := NewTypeSQLMarshaller(d, "")
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

	c, err = m.Insert(d)
	if err != nil {
		t.Errorf("cannot marshall to INSERT statement: %v", err)
	}
	t.Log(c)
	expectedSQL = `INSERT INTO dumbStruct (testInt, testString, testFloat, testPtr_aField_fk, testStruct_aField_fk) VALUES (1, "some velvet string", 2.000000, 3, 4);`
	if c != expectedSQL {
		t.Errorf("unexpected INSERT statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

	c, err = m.UpdatePK(d)
	if err != nil {
		t.Errorf("cannot marshall to UPDATE statement: %v", err)
	}
	t.Log(c)
	expectedSQL = `UPDATE dumbStruct SET testString="some velvet string", testFloat=2.000000, testPtr_aField_fk=3, testStruct_aField_fk=4 WHERE testInt=1;`
	if c != expectedSQL {
		t.Errorf("unexpected UPDATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

func TestTaggedMultiPK(t *testing.T) {
	dm := dumbFKMulti{
		aField:       3,
		aField2:      5,
		anotherField: "another string",
	}
	d := dumbStructMulti{
		testInt:    1,
		testString: "some velvet string",
		testFloat:  2.0,
		testPtr:    &dm,
		testStruct: dumbFKMulti{
			aField:       4,
			aField2:      6,
			anotherField: "another non struct field",
		},
	}
	m, err := NewTypeSQLMarshaller(d, "")
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

	c, err = m.Insert(d)
	if err != nil {
		t.Errorf("cannot marshall to INSERT statement: %v", err)
	}
	t.Log(c)
	expectedSQL = `INSERT INTO dumbStructMulti (testInt, testString, testFloat, testPtr_aField_fk, testPtr_aField2_fk, testStruct_aField_fk, testStruct_aField2_fk) VALUES (1, "some velvet string", 2.000000, 3, 5, 4, 6);`
	if c != expectedSQL {
		t.Errorf("unexpected INSERT statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

	c, err = m.UpdatePK(d)
	if err != nil {
		t.Errorf("cannot marshall to UPDATE statement: %v", err)
	}
	t.Log(c)
	expectedSQL = `UPDATE dumbStructMulti SET testString="some velvet string", testFloat=2.000000, testPtr_aField_fk=3, testPtr_aField2_fk=5, testStruct_aField_fk=4, testStruct_aField2_fk=6 WHERE testInt=1;`
	if c != expectedSQL {
		t.Errorf("unexpected UPDATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

	m, err = NewTypeSQLMarshaller(dm, "")
	if err != nil {
		t.Errorf("cannot create marshaler: %v", err)
	}

	c, err = m.Create(dr)
	if err != nil {
		t.Errorf("cannot marshall to CREATE statement: %v", err)
	}
	t.Log(c)
	expectedSQL = `CREATE TABLE dumbFKMulti (aField SMALLINT, aField2 SMALLINT, anotherField VARCHAR, PRIMARY KEY (aField ,aField2));`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

	c, err = m.UpdatePK(dm)
	if err != nil {
		t.Errorf("cannot marshall to UPDATE statement: %v", err)
	}
	t.Log(c)
	expectedSQL = `UPDATE dumbFKMulti SET anotherField="another string" WHERE aField=3 AND aField2=5;`
	if c != expectedSQL {
		t.Errorf("unexpected UPDATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

func TestUnTagged(t *testing.T) {
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
	m, err := NewTypeSQLMarshaller(d, "")
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

	m, err := NewTypeSQLMarshaller(sample, "")
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

	m, err := NewTypeSQLMarshaller(sample, "")
	if err != nil {
		return "", fmt.Errorf("cannot create marshaler: %v", err)
	}

	c, err := m.Insert(sample)
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
		t.Errorf("unexpected INSERT statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

type ReferenceUpdate struct {
	DifferentNameID int `sql:"primary"`
	AnExtraID       int `sql:"primary"`
	Name            string
	AnotherField    string
}

func doSQLUpdate() (string, error) {
	sample := ReferenceUpdate{
		DifferentNameID: 1,
		AnExtraID:       3,
		Name:            "a reference name",
		AnotherField:    "just to show off",
	}

	m, err := NewTypeSQLMarshaller(sample, "")
	if err != nil {
		return "", fmt.Errorf("cannot create marshaler: %v", err)
	}

	c, err := m.UpdatePK(sample)
	if err != nil {
		return "", fmt.Errorf("cannot marshall to UPDATE statement: %v", err)
	}
	return c, nil
}

func TestDocSampleUpdate(t *testing.T) {
	c, err := doSQLUpdate()
	if err != nil {
		t.Errorf("could not run documentation sample for UPDATE: %v", err)
	}
	t.Log(c)
	expectedSQL := `UPDATE ReferenceUpdate SET Name="a reference name", AnotherField="just to show off" WHERE DifferentNameID=1 AND AnExtraID=3;`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}

const (
	ValidYAMLSchema string = `
tables:
  humans:
    fields:
      id:
        type: int
        unique: true
        primary: true
      created:
        type: int
  mamals:
    fields:
      id:
        type: int
      created:
        type: string
        unique: true`
)

type DatabaseSchema struct {
	Tables map[string]map[string]interface{}
}

func TestCreateFromYAML(t *testing.T) {
	var schema DatabaseSchema

	err := goyaml.Unmarshal([]byte(ValidYAMLSchema), &schema)
	if err != nil {
		t.Errorf("%v", err)
	}

	var obtained []string
	for table_name, table := range schema.Tables {
		if fields, ok := table["fields"]; ok {
			m, err := NewTypeSQLMarshaller(fields, table_name)
			if err != nil {
				t.Errorf("Cannot create marshaller: %v", err)
			}

			dr := &ANSISQLDriver{}

			c, err := m.Create(dr)
			if err != nil {
				t.Errorf("cannot marshall to CREATE statement: %v", err)
			}
			obtained = append(obtained, c)
		}
	}

	expected := []string{
		"CREATE TABLE humans (id SMALLINT, created SMALLINT, PRIMARY KEY (id));",
		"CREATE TABLE mamals (id SMALLINT, created VARCHAR);",
	}

	if !reflect.DeepEqual(obtained, expected) {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expected, obtained)
	}

	t.Log(obtained)
}
