package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// RefreshToken holds the schema definition for tracking refresh tokens.
type RefreshToken struct {
	ent.Schema
}

// Fields of the RefreshToken.
func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.String("token").
			Unique().
			NotEmpty().
			Sensitive(),
		field.String("family").
			Optional().
			Comment("Token family for rotation tracking"),
		field.Time("expires_at").
			Comment("When the refresh token expires"),
		field.Bool("revoked").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the RefreshToken.
func (RefreshToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("refresh_tokens").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the RefreshToken.
func (RefreshToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token").Unique(),
		index.Fields("user_id"),
		index.Fields("family"),
		index.Fields("expires_at"),
	}
}
