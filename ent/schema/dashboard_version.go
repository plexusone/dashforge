package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DashboardVersion holds the schema definition for dashboard version history.
type DashboardVersion struct {
	ent.Schema
}

// Fields of the DashboardVersion.
func (DashboardVersion) Fields() []ent.Field {
	return []ent.Field{
		field.Int("version").
			Positive(),
		field.JSON("definition", map[string]any{}).
			Comment("DashboardIR JSON at this version"),
		field.String("change_summary").
			Optional().
			Comment("Description of changes"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the DashboardVersion.
func (DashboardVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dashboard", Dashboard.Type).
			Ref("versions").
			Unique().
			Required(),
	}
}

// Indexes of the DashboardVersion.
func (DashboardVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version"),
	}
}
