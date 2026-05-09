package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/index"

	cfmixin "github.com/grokify/systemforge/identity/ent/mixin"
)

// Principal holds the schema definition for the Principal entity.
// Principal is the unified identity root representing any type of actor:
// human, application, agent, or service.
type Principal struct {
	ent.Schema
}

// Mixin of the Principal.
func (Principal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.PrincipalMixin{},
	}
}

// Fields of the Principal.
// PrincipalMixin provides: id, type, identifier, display_name, organization_id,
// active, capabilities, allowed_scopes, metadata, core_control_principal_id, timestamps.
func (Principal) Fields() []ent.Field {
	// All core fields provided by PrincipalMixin
	return nil
}

// Edges of the Principal.
func (Principal) Edges() []ent.Edge {
	return []ent.Edge{
		// Organization relationship
		edge.From("organization", Organization.Type).
			Ref("principals").
			Field("organization_id").
			Unique(),

		// Type-specific extensions (one-to-one)
		edge.To("human", Human.Type).
			Unique().
			Comment("Human extension (when type=human)"),

		// OAuth accounts (external provider credentials)
		edge.To("oauth_accounts", OAuthAccount.Type).
			Comment("External OAuth provider accounts"),

		// Refresh tokens
		edge.To("refresh_tokens", RefreshToken.Type).
			Comment("Refresh tokens for this principal"),

		// Memberships in organizations
		edge.To("principal_memberships", PrincipalMembership.Type).
			Comment("Organization memberships"),

		// Dashboard-related edges
		edge.To("dashboards", Dashboard.Type).
			Comment("Dashboards owned by this principal"),

		edge.To("queries", SavedQuery.Type).
			Comment("Saved queries owned by this principal"),

		// Marketplace edges
		edge.To("owned_listings", Listing.Type).
			Comment("Marketplace listings owned by this principal"),

		edge.To("purchased_licenses", License.Type).
			Comment("Licenses purchased by this principal"),

		edge.To("seat_assignments", SeatAssignment.Type).
			Comment("License seats assigned to this principal"),

		edge.To("assigned_seats", SeatAssignment.Type).
			Comment("License seats assigned by this principal"),
	}
}

// Indexes of the Principal.
// PrincipalMixin provides: type+identifier (unique), organization_id, core_control_principal_id, active.
func (Principal) Indexes() []ent.Index {
	// Additional app-specific indexes
	return []ent.Index{
		index.Fields("type", "active"),
		index.Fields("organization_id", "type"),
	}
}
