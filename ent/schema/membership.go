package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Membership holds the schema definition for the user-organization relationship.
type Membership struct {
	ent.Schema
}

// Fields of the Membership.
func (Membership) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("organization_id", uuid.UUID{}),
		field.String("role").
			Default("viewer").
			Comment("Role within the organization (owner, admin, editor, viewer)"),
		field.JSON("permissions", []string{}).
			Optional().
			Comment("Fine-grained permissions"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Membership.
func (Membership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("memberships").
			Field("user_id").
			Unique().
			Required(),
		edge.From("organization", Organization.Type).
			Ref("memberships").
			Field("organization_id").
			Unique().
			Required(),
	}
}

// Indexes of the Membership.
func (Membership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").Unique(),
		index.Fields("organization_id"),
		index.Fields("role"),
	}
}
