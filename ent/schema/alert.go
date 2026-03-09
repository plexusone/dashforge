package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Alert holds the schema definition for alert definitions.
type Alert struct {
	ent.Schema
}

// Fields of the Alert.
func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").
			Unique().
			NotEmpty().
			Comment("URL-friendly identifier"),
		field.String("name").
			NotEmpty().
			Comment("Display name for this alert"),
		field.String("description").
			Optional().
			Comment("Description of what this alert monitors"),
		field.Enum("trigger_type").
			Values("threshold", "schedule", "data_change").
			Comment("Type of trigger that fires this alert"),
		field.JSON("trigger_config", map[string]any{}).
			Comment("Trigger-specific configuration"),
		field.Bool("enabled").
			Default(true).
			Comment("Whether this alert is active"),
		field.Int("cooldown_seconds").
			Default(300).
			Positive().
			Comment("Minimum seconds between alert firings"),
		field.Int("consecutive_failures").
			Default(0).
			Comment("Number of consecutive evaluation failures"),
		field.Time("last_triggered_at").
			Optional().
			Nillable().
			Comment("Last time this alert fired"),
		field.Time("last_evaluated_at").
			Optional().
			Nillable().
			Comment("Last time this alert was evaluated"),
		field.String("last_error").
			Optional().
			Comment("Last evaluation error if any"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("alerts").
			Unique().
			Required(),
		edge.From("dashboard", Dashboard.Type).
			Ref("alerts").
			Unique().
			Comment("Dashboard this alert is associated with (optional)"),
		edge.To("channels", Integration.Type).
			Through("alert_channels", AlertChannel.Type),
		edge.To("events", AlertEvent.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Alert.
func (Alert) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("trigger_type"),
		index.Fields("enabled"),
		index.Fields("last_triggered_at"),
	}
}
