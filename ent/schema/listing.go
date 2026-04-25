package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	cfmixin "github.com/grokify/coreforge/identity/ent/mixin"
)

// Listing holds the schema definition for marketplace listings.
// A listing represents a dashboard template available for purchase on the marketplace.
type Listing struct {
	ent.Schema
}

// Mixin returns the mixins for the Listing schema.
func (Listing) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.ListingMixin{},
	}
}

// Fields of the Listing (app-specific fields in addition to mixin).
func (Listing) Fields() []ent.Field {
	return []ent.Field{
		// DashForge-specific listing fields
		field.Bool("preview_enabled").
			Default(true).
			Comment("Allow preview of template before purchase"),

		field.Strings("tags").
			Optional().
			Comment("Searchable tags for marketplace discovery"),

		field.String("category").
			Optional().
			Comment("Listing category"),

		field.Int("install_count").
			Default(0).
			Comment("Number of times this template has been installed"),
	}
}

// Edges of the Listing.
func (Listing) Edges() []ent.Edge {
	return []ent.Edge{
		// Publisher (creator organization)
		edge.From("publisher", Publisher.Type).
			Ref("listings").
			Unique().
			Required().
			Field("creator_org_id"),

		// Owner principal
		edge.From("owner", Principal.Type).
			Ref("owned_listings").
			Unique().
			Required().
			Field("owner_id"),

		// The template being listed
		edge.From("template", DashboardTemplate.Type).
			Ref("listing").
			Unique().
			Field("product_id"),

		// Licenses granted for this listing
		edge.To("licenses", License.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Listing.
func (Listing) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("creator_org_id", "status"),
		index.Fields("product_type", "product_id").Unique(),
		index.Fields("status"),
		index.Fields("pricing_model"),
		index.Fields("category"),
	}
}
