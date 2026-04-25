package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	cfmixin "github.com/grokify/coreforge/identity/ent/mixin"
)

// License holds the schema definition for marketplace licenses.
// A license grants an organization access to a dashboard template.
type License struct {
	ent.Schema
}

// Mixin returns the mixins for the License schema.
func (License) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.LicenseMixin{},
	}
}

// Fields of the License (app-specific fields in addition to mixin).
func (License) Fields() []ent.Field {
	return []ent.Field{
		// DashForge-specific license fields
		field.Bool("auto_update").
			Default(true).
			Comment("Automatically update to new template versions"),

		field.Int("max_dashboards").
			Optional().
			Nillable().
			Comment("Maximum dashboards that can be created from this template (nil = unlimited)"),

		field.Int("current_dashboards").
			Default(0).
			Comment("Current number of dashboards created from this template"),
	}
}

// Edges of the License.
func (License) Edges() []ent.Edge {
	return []ent.Edge{
		// The listing this license grants access to
		edge.From("listing", Listing.Type).
			Ref("licenses").
			Unique().
			Required().
			Field("listing_id"),

		// The organization that holds this license
		edge.From("organization", Organization.Type).
			Ref("licenses").
			Unique().
			Required().
			Field("organization_id"),

		// The principal who purchased this license
		edge.From("purchaser", Principal.Type).
			Ref("purchased_licenses").
			Unique().
			Required().
			Field("purchased_by"),

		// Seat assignments for this license
		edge.To("seat_assignments", SeatAssignment.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the License.
func (License) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("listing_id", "organization_id").Unique(),
		index.Fields("organization_id"),
		index.Fields("purchased_by"),
		index.Fields("valid_until"),
	}
}
