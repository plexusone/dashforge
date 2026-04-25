package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"

	cfmixin "github.com/grokify/coreforge/identity/ent/mixin"
)

// RefreshToken holds the schema definition for tracking refresh tokens.
type RefreshToken struct {
	ent.Schema
}

// Mixin of the RefreshToken.
func (RefreshToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.RefreshTokenMixin{},
	}
}

// Fields of the RefreshToken.
// RefreshTokenMixin provides: id, principal_id, token, family, expires_at, revoked, created_at.
func (RefreshToken) Fields() []ent.Field {
	// All core fields provided by RefreshTokenMixin
	return nil
}

// Edges of the RefreshToken.
func (RefreshToken) Edges() []ent.Edge {
	return []ent.Edge{
		// Migrated from User to Principal
		edge.From("principal", Principal.Type).
			Ref("refresh_tokens").
			Field("principal_id").
			Unique().
			Required(),
	}
}

// Indexes of the RefreshToken.
// RefreshTokenMixin provides: token (unique), principal_id, family, expires_at.
func (RefreshToken) Indexes() []ent.Index {
	// All indexes provided by RefreshTokenMixin
	return nil
}
