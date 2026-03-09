package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Integration holds the schema definition for notification channel integrations.
type Integration struct {
	ent.Schema
}

// Fields of the Integration.
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").
			Unique().
			NotEmpty().
			Comment("URL-friendly identifier"),
		field.String("name").
			NotEmpty().
			Comment("Display name for this integration"),
		field.Enum("channel_type").
			Values("slack", "whatsapp", "email", "webhook").
			Comment("Type of notification channel"),
		field.JSON("config", map[string]any{}).
			Optional().
			Comment("Channel-specific configuration (non-sensitive)"),
		field.JSON("credentials", map[string]any{}).
			Optional().
			Sensitive().
			Comment("Encrypted credentials (API keys, tokens, etc.)"),
		field.Enum("status").
			Values("active", "inactive", "error").
			Default("inactive").
			Comment("Current status of the integration"),
		field.String("status_message").
			Optional().
			Comment("Status details (e.g., error message)"),
		field.Enum("source").
			Values("builtin", "marketplace", "custom").
			Default("builtin").
			Comment("Where this integration came from"),
		field.String("marketplace_slug").
			Optional().
			Comment("Slug from marketplace if source=marketplace"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("Last time a message was sent via this integration"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Integration.
func (Integration) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("integrations").
			Unique().
			Required(),
		edge.From("alerts", Alert.Type).
			Ref("channels").
			Through("alert_channels", AlertChannel.Type),
	}
}

// Indexes of the Integration.
func (Integration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("channel_type"),
		index.Fields("status"),
		index.Fields("source"),
	}
}
