package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"

	cfmixin "github.com/grokify/systemforge/identity/ent/mixin"
)

// OAuthAccount holds the schema definition for external OAuth provider links.
type OAuthAccount struct {
	ent.Schema
}

// Mixin of the OAuthAccount.
func (OAuthAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.OAuthAccountMixin{},
	}
}

// Fields of the OAuthAccount.
// OAuthAccountMixin provides: id, principal_id, provider, provider_account_id,
// access_token, refresh_token, token_expires_at, scopes, raw_data, timestamps.
func (OAuthAccount) Fields() []ent.Field {
	// All core fields provided by OAuthAccountMixin
	return nil
}

// Edges of the OAuthAccount.
func (OAuthAccount) Edges() []ent.Edge {
	return []ent.Edge{
		// Migrated from User to Principal
		edge.From("principal", Principal.Type).
			Ref("oauth_accounts").
			Field("principal_id").
			Unique().
			Required(),
	}
}

// Indexes of the OAuthAccount.
// OAuthAccountMixin provides: provider+provider_account_id (unique), principal_id.
func (OAuthAccount) Indexes() []ent.Index {
	// All indexes provided by OAuthAccountMixin
	return nil
}
