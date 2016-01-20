// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import "testing"

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
	// TODO(perrito666) check statement without depending on
	// the order of the map.
	t.Log(c)
	expectedSQL := `CREATE TABLE dumbStruct testInt SMALLINT, testString VARCHAR, testFloat FLOAT, testPtr FOREIGN KEY (aField) REFERENCES dumbFK (aField) ON DELETE CASCADE ON UPDATE CASCADE, testStruct FOREIGN KEY (aField) REFERENCES dumbFK (aField) ON DELETE CASCADE ON UPDATE CASCADE, PRIMARY KEY (testInt);`
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
	// TODO(perrito666) check statement without depending on
	// the order of the map.
	t.Log(c)
	expectedSQL := `CREATE TABLE untaggedDumbStruct testInt SMALLINT, testString VARCHAR, testFloat FLOAT, testPtr FOREIGN KEY (testPtr) REFERENCES untaggedDumbFK (testPtr) ON DELETE CASCADE ON UPDATE CASCADE, testStruct FOREIGN KEY (testStruct) REFERENCES untaggedDumbFK (testStruct) ON DELETE CASCADE ON UPDATE CASCADE;`
	if c != expectedSQL {
		t.Errorf("unexpected CREATE statement: \nexpected: %q\nobtained: %q", expectedSQL, c)
	}

}
