package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertChannel holds the schema for the junction table between alerts and integrations.
type AlertChannel struct {
	ent.Schema
}

// Annotations of the AlertChannel.
func (AlertChannel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("alert_id", "integration_id"),
	}
}

// Fields of the AlertChannel.
func (AlertChannel) Fields() []ent.Field {
	return []ent.Field{
		field.Int("alert_id"),
		field.Int("integration_id"),
		field.JSON("channel_config", map[string]any{}).
			Optional().
			Comment("Per-alert channel overrides (e.g., specific Slack channel)"),
		field.Bool("enabled").
			Default(true).
			Comment("Whether this channel is enabled for this alert"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the AlertChannel.
func (AlertChannel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alert", Alert.Type).
			Required().
			Unique().
			Field("alert_id"),
		edge.To("integration", Integration.Type).
			Required().
			Unique().
			Field("integration_id"),
	}
}

// Indexes of the AlertChannel.
func (AlertChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("alert_id", "integration_id").Unique(),
		index.Fields("enabled"),
	}
}
