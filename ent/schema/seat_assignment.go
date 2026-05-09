package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/index"
	cfmixin "github.com/grokify/systemforge/identity/ent/mixin"
)

// SeatAssignment holds the schema definition for license seat assignments.
// A seat assignment grants a specific principal access to a licensed template.
type SeatAssignment struct {
	ent.Schema
}

// Mixin returns the mixins for the SeatAssignment schema.
func (SeatAssignment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.SeatAssignmentMixin{},
	}
}

// Fields of the SeatAssignment (app-specific fields in addition to mixin).
func (SeatAssignment) Fields() []ent.Field {
	return []ent.Field{
		// No additional app-specific fields needed for DashForge
	}
}

// Edges of the SeatAssignment.
func (SeatAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		// The license this seat belongs to
		edge.From("license", License.Type).
			Ref("seat_assignments").
			Unique().
			Required().
			Field("license_id"),

		// The principal assigned to this seat
		edge.From("principal", Principal.Type).
			Ref("seat_assignments").
			Unique().
			Required().
			Field("principal_id"),

		// The principal who made this assignment
		edge.From("assigner", Principal.Type).
			Ref("assigned_seats").
			Unique().
			Required().
			Field("assigned_by"),
	}
}

// Indexes of the SeatAssignment.
func (SeatAssignment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("license_id", "principal_id").Unique(),
		index.Fields("principal_id"),
	}
}
