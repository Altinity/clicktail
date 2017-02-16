package gonx

import (
	"fmt"
	"strconv"
	"strings"
)

// Shortcut for the map of strings
type Fieldmap map[string]string

// Parsed log record. Use Get method to retrieve a value by name instead of
// threating this as a map, because inner representation is in design.

// renaming "fields" to "Fields" and "Fields" to "Fieldmap" to export it
type Entry struct {
	Fields Fieldmap
}

// Creates an empty Entry to be filled later
func NewEmptyEntry() *Entry {
	return &Entry{make(Fieldmap)}
}

// Creates an Entry with fiven Fields
func NewEntry(Fields Fieldmap) *Entry {
	return &Entry{Fields}
}

// Return entry field value by name or empty string and error if it
// does not exist.
func (entry *Entry) Field(name string) (value string, err error) {
	value, ok := entry.Fields[name]
	if !ok {
		err = fmt.Errorf("field '%v' does not found in record %+v", name, *entry)
	}
	return
}

// Return entry field value as float64. Rutuen nil if field does not exists
// and convertion error if cannot cast a type.
func (entry *Entry) FloatField(name string) (value float64, err error) {
	tmp, err := entry.Field(name)
	if err == nil {
		value, err = strconv.ParseFloat(tmp, 64)
	}
	return
}

// Field value setter
func (entry *Entry) SetField(name string, value string) {
	entry.Fields[name] = value
}

// Float field value setter. It accepts float64, but still store it as a
// string in the same Fields map. The precision is 2, its enough for log
// parsing task
func (entry *Entry) SetFloatField(name string, value float64) {
	entry.SetField(name, strconv.FormatFloat(value, 'f', 2, 64))
}

// Integer field value setter. It accepts float64, but still store it as a
// string in the same Fields map.
func (entry *Entry) SetUintField(name string, value uint64) {
	entry.SetField(name, strconv.FormatUint(uint64(value), 10))
}

// Merge two entries by updating values for master entry with given.
func (master *Entry) Merge(entry *Entry) {
	for name, value := range entry.Fields {
		master.SetField(name, value)
	}
}
func (entry *Entry) FieldsHash(Fields []string) string {
	var key []string
	for _, name := range Fields {
		value, err := entry.Field(name)
		if err != nil {
			value = "NULL"
		}
		key = append(key, fmt.Sprintf("'%v'=%v", name, value))
	}
	return strings.Join(key, ";")
}

func (entry *Entry) Partial(Fields []string) *Entry {
	partial := NewEmptyEntry()
	for _, name := range Fields {
		value, _ := entry.Field(name)
		partial.SetField(name, value)
	}
	return partial
}
