// Copyright 2016 Horacio Duran.
// Licenced under the MIT licence, see LICENCE for details.
package sqlmarshal

import "fmt"

// FieldWithValue contains a field name and its value.
type FieldWithValue struct {
	Name  string
	Value string
}

// FieldsWithValue contains many FieldWithValue and has some
// convenience methods to handle them.
type FieldsWithValue struct {
	fields        []FieldWithValue
	innerRegistry map[string]int
}

// NewFieldsWithValue returns a pointer to a FieldsWithValue.
func NewFieldsWithValue() *FieldsWithValue {
	f := &FieldsWithValue{}
	f.innerRegistry = make(map[string]int)
	return f
}

// Append tries to concatenate this FieldsWithValue with the passed
// one and fails if there are repetitions.
func (f *FieldsWithValue) Append(field *FieldsWithValue) error {
	return f.Add(field.fields...)
}

// Add adds a new FieldWithValue or errors if it already exists.
func (f *FieldsWithValue) Add(fields ...FieldWithValue) error {
	for _, field := range fields {
		_, ok := f.innerRegistry[field.Name]
		if ok {
			return fmt.Errorf("field %q already present", field.Name)
		}
		f.fields = append(f.fields, field)
		f.innerRegistry[field.Name] = len(f.fields) - 1
	}
	return nil
}

// Pop removes and returns a Field and a bool indicating if
// the field exists.
func (f *FieldsWithValue) Pop(name string) (FieldWithValue, bool) {
	_, ok := f.innerRegistry[name]
	if !ok {
		return FieldWithValue{}, false
	}
	delete(f.innerRegistry, name)
	for i := range f.fields {
		field := f.fields[i]
		if field.Name == name {
			a := append(f.fields[:i], f.fields[i+1:]...)
			f.fields = a
			return field, true
		}
	}
	return FieldWithValue{}, false
}

// Pairs returns a string slice of the pairs key/value of each field
// joined by a passed separator.
func (f *FieldsWithValue) Pairs(separator string) []string {
	pairs := make([]string, len(f.fields))
	for i, v := range f.fields {
		pairs[i] = fmt.Sprintf("%s%s%s", v.Name, separator, v.Value)
	}
	return pairs
}

// Contains returns true if the passed field name corresponds to
// a FieldWithValue already in this FieldsWithValue.
func (f *FieldsWithValue) Contains(field string) bool {
	_, ok := f.innerRegistry[field]
	return ok
}

// Fields returns the field names for all the FieldWithValue.
func (f *FieldsWithValue) Fields() []string {
	fields := make([]string, len(f.fields))
	for i := range f.fields {
		fields[i] = f.fields[i].Name
	}
	return fields
}

// Values returns the field value for all the FieldWithValue.
func (f *FieldsWithValue) Values() []string {
	values := make([]string, len(f.fields))
	for i := range f.fields {
		values[i] = f.fields[i].Value
	}
	return values
}

// Len returns the amount of FieldWithValue in this
// FieldsWithValue.
func (f *FieldsWithValue) Len() int {
	return len(f.fields)
}
