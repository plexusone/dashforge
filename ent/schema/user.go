package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("name").
			Optional(),
		field.String("password_hash").
			Optional().
			Sensitive(),
		field.String("avatar_url").
			Optional(),
		field.Bool("is_platform_admin").
			Default(false).
			Comment("Cross-org admin access"),
		field.UUID("core_control_principal_id", uuid.UUID{}).
			Optional().
			Nillable().
			Unique().
			Comment("CoreControl Principal ID for SSO"),
		field.Bool("active").
			Default(true),
		field.Time("last_login_at").
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

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", Membership.Type),
		edge.To("dashboards", Dashboard.Type),
		edge.To("queries", SavedQuery.Type),
		edge.To("oauth_accounts", OAuthAccount.Type),
		edge.To("refresh_tokens", RefreshToken.Type),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
		index.Fields("active"),
		index.Fields("core_control_principal_id"),
	}
}
