package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/index"

	cfmixin "github.com/grokify/coreforge/identity/ent/mixin"
)

// PrincipalMembership holds the schema definition for principal-organization memberships.
// This replaces the previous Membership schema to support all principal types.
type PrincipalMembership struct {
	ent.Schema
}

// Mixin of the PrincipalMembership.
func (PrincipalMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.PrincipalMembershipMixin{},
	}
}

// Fields of the PrincipalMembership.
// PrincipalMembershipMixin provides: id, principal_id, organization_id, role, active,
// permissions, joined_at, expires_at, timestamps.
func (PrincipalMembership) Fields() []ent.Field {
	// All core fields provided by PrincipalMembershipMixin
	return nil
}

// Edges of the PrincipalMembership.
func (PrincipalMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("principal", Principal.Type).
			Ref("principal_memberships").
			Field("principal_id").
			Required().
			Unique(),

		edge.From("organization", Organization.Type).
			Ref("principal_memberships").
			Field("organization_id").
			Required().
			Unique(),
	}
}

// Indexes of the PrincipalMembership.
// PrincipalMembershipMixin provides: principal_id+organization_id (unique), organization_id, active.
func (PrincipalMembership) Indexes() []ent.Index {
	// Additional app-specific indexes
	return []ent.Index{
		index.Fields("principal_id"),
		index.Fields("organization_id", "role"),
		index.Fields("role"),
	}
}
