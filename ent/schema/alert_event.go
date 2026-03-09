package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEvent holds the schema for alert history/audit log.
type AlertEvent struct {
	ent.Schema
}

// Fields of the AlertEvent.
func (AlertEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("event_type").
			Values("triggered", "resolved", "acknowledged", "error", "cooldown_skipped").
			Comment("Type of event"),
		field.JSON("trigger_data", map[string]any{}).
			Optional().
			Comment("Data that triggered the alert (metric value, etc.)"),
		field.JSON("channels_notified", []string{}).
			Optional().
			Comment("List of integration slugs that were notified"),
		field.Int("channels_success").
			Default(0).
			Comment("Number of channels that succeeded"),
		field.Int("channels_failed").
			Default(0).
			Comment("Number of channels that failed"),
		field.String("error_message").
			Optional().
			Comment("Error message if event_type=error"),
		field.String("acknowledged_by").
			Optional().
			Comment("User who acknowledged (if event_type=acknowledged)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("When this event occurred"),
	}
}

// Edges of the AlertEvent.
func (AlertEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert", Alert.Type).
			Ref("events").
			Unique().
			Required(),
	}
}

// Indexes of the AlertEvent.
func (AlertEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_type"),
		index.Fields("created_at"),
	}
}
