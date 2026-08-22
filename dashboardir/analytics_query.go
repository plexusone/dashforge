package dashboardir

// AnalyticsQueryRequest is the neutral request shape for read-only analytical
// query execution. GrokifyQL providers can implement this over SQL, Ent, Dolt,
// application APIs, or another backend while keeping UIForge decoupled.
type AnalyticsQueryRequest struct {
	SourceID string         `json:"sourceId"`
	Query    string         `json:"query"`
	Dialect  string         `json:"dialect,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
	Limit    int            `json:"limit,omitempty"`
}

// AnalyticsQueryResult is the neutral tabular response for query previews and
// saved questions.
type AnalyticsQueryResult struct {
	Columns       []AnalyticsQueryColumn `json:"columns"`
	Rows          []map[string]any       `json:"rows"`
	RowCount      int                    `json:"rowCount"`
	ExecutionTime int64                  `json:"executionTimeMs,omitempty"`
}

// AnalyticsQueryColumn describes one returned result column.
type AnalyticsQueryColumn struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}
