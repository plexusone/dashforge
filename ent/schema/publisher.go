package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/index"
	cfmixin "github.com/grokify/systemforge/identity/ent/mixin"
)

// Publisher holds the schema definition for dashboard template publishers.
// A publisher is a creator organization that creates and sells dashboard templates
// on the DashForge marketplace.
type Publisher struct {
	ent.Schema
}

// Mixin returns the mixins for the Publisher schema.
func (Publisher) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.CreatorOrgMixin{},
	}
}

// Fields of the Publisher (app-specific fields in addition to mixin).
func (Publisher) Fields() []ent.Field {
	return []ent.Field{
		// DashForge-specific publisher fields can be added here
	}
}

// Edges of the Publisher.
func (Publisher) Edges() []ent.Edge {
	return []ent.Edge{
		// Templates created by this publisher
		edge.To("templates", DashboardTemplate.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)).
			Comment("Dashboard templates created by this publisher"),

		// Listings created by this publisher
		edge.To("listings", Listing.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)).
			Comment("Marketplace listings by this publisher"),
	}
}

// Indexes of the Publisher.
func (Publisher) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("active"),
		index.Fields("verified"),
	}
}
