// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import "testing"

type dumbPtr struct {
	aField       int
	anotherField string
}

type dumbStruct struct {
	testInt    int
	testString string
	testFloat  float32
	testPtr    *dumbPtr
	testStruct dumbPtr
}

func (*dumbStruct) methodsAreIgnored() {
}

func TestCreate(t *testing.T) {
	d := dumbStruct{
		testInt:    1,
		testString: "some velvet string",
		testFloat:  2.0,
		testPtr: &dumbPtr{
			aField:       3,
			anotherField: "another string",
		},
		testStruct: dumbPtr{
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
}
