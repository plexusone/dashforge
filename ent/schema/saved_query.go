package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SavedQuery holds the schema definition for saved SQL queries.
type SavedQuery struct {
	ent.Schema
}

// Fields of the SavedQuery.
func (SavedQuery) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Text("query").
			NotEmpty().
			Comment("SQL query text"),
		field.String("datasource_slug").
			NotEmpty().
			Comment("Reference to DataSource"),
		field.JSON("parameters", []QueryParameter{}).
			Optional().
			Comment("Query parameters definition"),
		field.JSON("result_columns", []ResultColumn{}).
			Optional().
			Comment("Expected result columns"),
		field.Enum("visibility").
			Values("private", "team", "public").
			Default("private"),
		field.Int("cache_ttl_seconds").
			Default(0).
			Comment("Cache TTL in seconds, 0 = no cache"),
		field.Time("last_run_at").
			Optional().
			Nillable(),
		field.Int64("last_run_duration_ms").
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

// QueryParameter defines a parameter for a saved query.
type QueryParameter struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // string, int, float, date, bool
	Required     bool   `json:"required"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Description  string `json:"description,omitempty"`
}

// ResultColumn defines expected result column metadata.
type ResultColumn struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

// Edges of the SavedQuery.
func (SavedQuery) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("queries").
			Unique().
			Required(),
		edge.From("owner", User.Type).
			Ref("queries").
			Unique(),
	}
}

// Indexes of the SavedQuery.
func (SavedQuery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("datasource_slug"),
		index.Fields("visibility"),
	}
}
