// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// AuroraHealthCheckColumns holds the columns for the "aurora_health_check" table.
	AuroraHealthCheckColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "ts", Type: field.TypeTime},
	}
	// AuroraHealthCheckTable holds the schema information for the "aurora_health_check" table.
	AuroraHealthCheckTable = &schema.Table{
		Name:       "aurora_health_check",
		Columns:    AuroraHealthCheckColumns,
		PrimaryKey: []*schema.Column{AuroraHealthCheckColumns[0]},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		AuroraHealthCheckTable,
	}
)

func init() {
	AuroraHealthCheckTable.Annotation = &entsql.Annotation{
		Table: "aurora_health_check",
	}
}