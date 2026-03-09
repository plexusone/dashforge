package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OAuthAccount holds the schema definition for external OAuth provider links.
type OAuthAccount struct {
	ent.Schema
}

// Fields of the OAuthAccount.
func (OAuthAccount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("provider").
			Values("github", "google", "corecontrol").
			Comment("OAuth provider name"),
		field.String("provider_user_id").
			NotEmpty().
			Comment("User ID from the OAuth provider"),
		field.String("access_token").
			Optional().
			Sensitive(),
		field.String("refresh_token").
			Optional().
			Sensitive(),
		field.Time("token_expires_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the OAuthAccount.
func (OAuthAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("oauth_accounts").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the OAuthAccount.
func (OAuthAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "provider_user_id").Unique(),
		index.Fields("user_id"),
	}
}
