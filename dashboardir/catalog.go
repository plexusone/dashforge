package dashboardir

// AnalyticsCatalog describes queryable data in a neutral shape that can be
// supplied by application-specific systems such as OmniRoadmap or VisionStudio,
// or by generic SQL/database introspection.
type AnalyticsCatalog struct {
	// ID is a stable catalog identifier, e.g. "omniroadmap".
	ID string `json:"id"`

	// Name is the display name for this catalog.
	Name string `json:"name"`

	// Description provides optional context for humans and AI-assisted builders.
	Description string `json:"description,omitempty"`

	// Sources are the connected analytical sources represented by the catalog.
	Sources []AnalyticsSource `json:"sources"`
}

// AnalyticsSource is one connected analytical source, such as an OmniRoadmap
// database, a VisionStudio database, or a PostgreSQL connection.
type AnalyticsSource struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	Description string             `json:"description,omitempty"`
	Datasets    []AnalyticsDataset `json:"datasets"`
}

// AnalyticsSourceConfig is the persisted configuration for one analytics
// source. It is separate from AnalyticsSource because the runtime catalog is
// safe to return to browsers, while source configuration may include secret
// references and operational settings.
type AnalyticsSourceConfig struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Connector     string            `json:"connector"`
	Description   string            `json:"description,omitempty"`
	CredentialRef string            `json:"credentialRef,omitempty"`
	DSNRef        string            `json:"dsnRef,omitempty"`
	Enabled       bool              `json:"enabled"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// CredentialReference returns the configured secret reference for this source.
func (s AnalyticsSourceConfig) CredentialReference() string {
	if s.CredentialRef != "" {
		return s.CredentialRef
	}
	return s.DSNRef
}

// AnalyticsDataset is a queryable entity/table/model within a source.
type AnalyticsDataset struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	QueryName   string           `json:"queryName"`
	Description string           `json:"description,omitempty"`
	Fields      []AnalyticsField `json:"fields"`
}

// AnalyticsField is a queryable field/column. It intentionally carries both
// database-style and semantic metadata so DashForge can power table builders,
// chart builders, query linting, and AI-assisted dashboard generation.
type AnalyticsField struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	QueryName    string   `json:"queryName"`
	Type         string   `json:"type"`
	Source       string   `json:"source"`
	Role         string   `json:"role,omitempty"`
	Description  string   `json:"description,omitempty"`
	Selectable   bool     `json:"selectable"`
	Filterable   bool     `json:"filterable"`
	Sortable     bool     `json:"sortable"`
	Nullable     bool     `json:"nullable,omitempty"`
	Count        int      `json:"count,omitempty"`
	Coverage     float64  `json:"coverage,omitempty"`
	SampleValues []string `json:"sampleValues,omitempty"`
}

// Analytics source type constants.
const (
	AnalyticsSourceTypeApplication = "application"
	AnalyticsSourceTypeSQL         = "sql"
)

// Analytics field source constants.
const (
	AnalyticsFieldSourceStandard = "standard"
	AnalyticsFieldSourceCustom   = "custom"
	AnalyticsFieldSourceDerived  = "derived"
	AnalyticsFieldSourceMetadata = "metadata"
)

// Analytics field role constants.
const (
	AnalyticsFieldRoleDimension = "dimension"
	AnalyticsFieldRoleMeasure   = "measure"
	AnalyticsFieldRoleTime      = "time"
	AnalyticsFieldRoleLink      = "link"
)

// Analytics field type constants.
const (
	AnalyticsFieldTypeString = "string"
	AnalyticsFieldTypeNumber = "number"
	AnalyticsFieldTypeBool   = "bool"
	AnalyticsFieldTypeDate   = "date"
	AnalyticsFieldTypeJSON   = "json"
)
