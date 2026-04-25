package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Dashboard holds the schema definition for the Dashboard entity.
type Dashboard struct {
	ent.Schema
}

// Fields of the Dashboard.
func (Dashboard) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").
			Unique().
			NotEmpty().
			Comment("URL-friendly identifier"),
		field.String("title").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.JSON("definition", map[string]any{}).
			Comment("Full DashboardIR JSON"),
		field.Enum("visibility").
			Values("private", "team", "public").
			Default("private"),
		field.Bool("archived").
			Default(false),
		field.Int("version").
			Default(1),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Dashboard.
func (Dashboard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("dashboards").
			Unique().
			Required(),
		// Migrated from User to Principal
		edge.From("owner", Principal.Type).
			Ref("dashboards").
			Unique(),
		edge.To("versions", DashboardVersion.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("alerts", Alert.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

// Indexes of the Dashboard.
func (Dashboard) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("visibility"),
		index.Fields("archived"),
	}
}
