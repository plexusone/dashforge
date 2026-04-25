package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	cfmixin "github.com/grokify/coreforge/identity/ent/mixin"
)

// Subscription holds the schema definition for platform subscriptions.
// A subscription grants an organization access to DashForge platform features.
type Subscription struct {
	ent.Schema
}

// Mixin returns the mixins for the Subscription schema.
func (Subscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		cfmixin.SubscriptionMixin{},
	}
}

// Fields of the Subscription (app-specific fields in addition to mixin).
func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		// DashForge-specific subscription limits
		field.Int("max_dashboards").
			Optional().
			Nillable().
			Comment("Maximum dashboards allowed (nil = unlimited)"),

		field.Int("max_datasources").
			Optional().
			Nillable().
			Comment("Maximum datasources allowed (nil = unlimited)"),

		field.Int("max_queries").
			Optional().
			Nillable().
			Comment("Maximum saved queries allowed (nil = unlimited)"),

		field.Int("max_alerts").
			Optional().
			Nillable().
			Comment("Maximum alerts allowed (nil = unlimited)"),

		field.Int("max_team_members").
			Optional().
			Nillable().
			Comment("Maximum team members (nil = unlimited)"),

		field.Bool("custom_domain_enabled").
			Default(false).
			Comment("Whether custom domain is enabled"),

		field.Bool("white_label_enabled").
			Default(false).
			Comment("Whether white-labeling is enabled"),

		field.Bool("api_access_enabled").
			Default(false).
			Comment("Whether API access is enabled"),

		field.Bool("realtime_enabled").
			Default(false).
			Comment("Whether real-time dashboard updates are enabled"),

		field.Int("data_retention_days").
			Optional().
			Nillable().
			Comment("Data retention period in days (nil = unlimited)"),
	}
}

// Edges of the Subscription.
func (Subscription) Edges() []ent.Edge {
	return []ent.Edge{
		// The organization this subscription belongs to
		edge.From("organization", Organization.Type).
			Ref("subscription").
			Unique().
			Required().
			Field("organization_id"),
	}
}

// Indexes of the Subscription.
func (Subscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("plan_tier"),
		index.Fields("current_period_end"),
	}
}
