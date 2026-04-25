package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DashboardTemplate holds the schema definition for sellable dashboard templates.
// Templates are reusable dashboard configurations that can be sold on the marketplace.
type DashboardTemplate struct {
	ent.Schema
}

// Mixin of the DashboardTemplate.
func (DashboardTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the DashboardTemplate.
func (DashboardTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("publisher_id", uuid.UUID{}).
			Comment("Publisher who created this template"),

		field.String("slug").
			NotEmpty().
			Unique().
			Comment("URL-friendly identifier"),

		field.String("title").
			NotEmpty().
			MaxLen(200),

		field.Text("description").
			Optional(),

		// The actual template definition
		field.JSON("definition", map[string]any{}).
			Comment("Dashboard template definition (DashboardIR)"),

		// Template metadata
		field.Strings("tags").
			Optional().
			Comment("Searchable tags"),

		field.String("category").
			Optional().
			Comment("Template category (analytics, monitoring, sales, etc.)"),

		field.Strings("required_datasources").
			Optional().
			Comment("List of required datasource types"),

		// Preview assets
		field.String("thumbnail_url").
			Optional().
			Comment("Preview thumbnail"),

		field.Strings("screenshot_urls").
			Optional().
			Comment("Gallery screenshots"),

		// Status
		field.Enum("status").
			Values("draft", "published", "archived").
			Default("draft"),

		// Version tracking
		field.String("version").
			Default("1.0.0").
			Comment("Semantic version"),
	}
}

// Edges of the DashboardTemplate.
func (DashboardTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		// Publisher who owns this template
		edge.From("publisher", Publisher.Type).
			Ref("templates").
			Field("publisher_id").
			Unique().
			Required(),

		// Marketplace listing for this template
		edge.To("listing", Listing.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)).
			Comment("Marketplace listing for this template"),
	}
}

// Indexes of the DashboardTemplate.
func (DashboardTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("publisher_id"),
		index.Fields("status"),
		index.Fields("category"),
	}
}
