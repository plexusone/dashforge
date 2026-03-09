package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Organization holds the schema definition for multi-tenant organizations.
type Organization struct {
	ent.Schema
}

// Fields of the Organization.
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("slug").
			Unique().
			NotEmpty().
			Comment("URL-friendly identifier"),
		field.String("name").
			NotEmpty(),
		field.String("logo_url").
			Optional(),
		field.String("domain").
			Optional().
			Comment("Custom domain for organization"),
		field.Enum("plan").
			Values("free", "starter", "pro", "enterprise").
			Default("free"),
		field.Bool("active").
			Default(true),
		field.JSON("settings", map[string]any{}).
			Optional().
			Comment("Organization-specific settings"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Organization.
func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", Membership.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("dashboards", Dashboard.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("datasources", DataSource.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("queries", SavedQuery.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("integrations", Integration.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("alerts", Alert.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Organization.
func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("domain").Unique(),
		index.Fields("active"),
	}
}
