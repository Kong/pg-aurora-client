// Code generated by ent, DO NOT EDIT.

package aurorahealthcheck

import (
	"time"
)

const (
	// Label holds the string label denoting the aurorahealthcheck type in the database.
	Label = "aurora_health_check"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldTs holds the string denoting the ts field in the database.
	FieldTs = "ts"
	// Table holds the table name of the aurorahealthcheck in the database.
	Table = "aurora_health_check"
)

// Columns holds all SQL columns for aurorahealthcheck fields.
var Columns = []string{
	FieldID,
	FieldTs,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultTs holds the default value on creation for the "ts" field.
	DefaultTs func() time.Time
)
