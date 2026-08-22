package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AnalyticsSource holds persisted analytics source configuration. It stores a
// secret reference (dsn_ref) such as env://VAR_NAME — never a resolved DSN.
type AnalyticsSource struct {
	ent.Schema
}

// Mixin of the AnalyticsSource.
func (AnalyticsSource) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

// Fields of the AnalyticsSource.
func (AnalyticsSource) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").
			Unique().
			NotEmpty().
			Comment("Stable source ID used in catalogs, queries, and authz resource IDs"),
		field.String("name").
			NotEmpty(),
		field.String("connector").
			NotEmpty().
			Comment("Connector registry name, e.g. omniroadmap"),
		field.String("dsn_ref").
			NotEmpty().
			Comment("OmniVault secret reference (env://..., file://...); never a raw DSN"),
		field.Bool("enabled").
			Default(true),
	}
}

// Indexes of the AnalyticsSource.
func (AnalyticsSource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("connector"),
		index.Fields("enabled"),
	}
}
