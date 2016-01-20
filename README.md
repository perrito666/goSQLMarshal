# goSQLMarshal
A go package that allows struct serialization into SQL statements.

The idea behind this package is to allow the serialization of structure into SQL, ideally we will provide:
 * [CREATE](#create)
 * INSERT
 * SELECT
 * UPDATE
 * DELETE

# CREATE

Generates the **CREATE** *SQL* statement for the given structure.
In this example we can see the usage of basic types, Foreign and Primary keys.

 * Basic Types are generated from the builtin types in go, there are equivalent for most basics.
 * Foreign Keys are generated from Structs or Pointer to structs, no other pointer is supported so far.
 * Primary Keys are generated from the fields tagged as such.

To help improve the SQL generated, tags can be used in the format: `sql:"tag,tag,tag"`
Currently the supported are:
 * *primary* : it will make the tagged field the (or one of the, in case of multiple,  primary key)

If no primary key is tagged, there will be none, the Foreign Key pointing to a non primary key structure 
will assume that the key name is the same as the field in the referencing struct.

***Note:** there is no consistency checking at present:*

```go
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
```
```sql
CREATE TABLE Sample ID SMALLINT, Name VARCHAR, Reference FOREIGN KEY (DifferentNameID) REFERENCES Reference (DifferentNameID) ON DELETE CASCADE ON UPDATE CASCADE, ConcreteReference FOREIGN KEY (DifferentNameID) REFERENCES Reference (DifferentNameID) ON DELETE CASCADE ON UPDATE CASCADE, PRIMARY KEY (ID);
```
