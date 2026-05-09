package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/index"

	cfmixin "github.com/grokify/systemforge/identity/ent/mixin"
)

// Human holds the schema definition for human-specific principal data.
// This is a one-to-one extension of Principal where type="human".
type Human struct {
	ent.Schema
}

// Mixin of the Human.
func (Human) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.HumanMixin{},
	}
}

// Fields of the Human.
// HumanMixin provides: id, principal_id, email, name, password_hash, avatar_url,
// is_platform_admin, last_login_at, email_verified_at, timestamps.
func (Human) Fields() []ent.Field {
	// All core fields provided by HumanMixin
	return nil
}

// Edges of the Human.
func (Human) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("principal", Principal.Type).
			Ref("human").
			Field("principal_id").
			Required().
			Unique().
			Immutable(),
	}
}

// Indexes of the Human.
// HumanMixin provides: email (unique), principal_id (unique).
func (Human) Indexes() []ent.Index {
	// Additional app-specific indexes
	return []ent.Index{
		index.Fields("is_platform_admin"),
	}
}
