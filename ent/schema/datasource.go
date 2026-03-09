package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DataSource holds the schema definition for database connections.
type DataSource struct {
	ent.Schema
}

// Fields of the DataSource.
func (DataSource) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),
		field.Enum("type").
			Values("postgres", "mysql", "sqlite", "clickhouse", "duckdb").
			Comment("Database type"),
		field.String("connection_url").
			Sensitive().
			Comment("Connection string (encrypted at rest)"),
		field.String("connection_url_env").
			Optional().
			Comment("Environment variable name for connection URL"),
		field.Int("max_connections").
			Default(10).
			Positive(),
		field.Int("query_timeout_seconds").
			Default(30).
			Positive(),
		field.Bool("read_only").
			Default(true).
			Comment("If true, only SELECT queries allowed"),
		field.Bool("ssl_enabled").
			Default(true),
		field.Bool("active").
			Default(true),
		field.Time("last_connected_at").
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

// Edges of the DataSource.
func (DataSource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("datasources").
			Unique().
			Required(),
	}
}

// Indexes of the DataSource.
func (DataSource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("type"),
		index.Fields("active"),
	}
}
