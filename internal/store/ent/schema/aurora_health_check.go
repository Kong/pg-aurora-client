package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
)

type AuroraHealthCheck struct {
	ent.Schema
}

func (AuroraHealthCheck) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.Time("ts").Default(time.Now),
	}
}

func (AuroraHealthCheck) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "aurora_health_check"},
	}
}
