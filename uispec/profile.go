package uispec

const (
	ProfileDashboard   = "dashboard"
	ProfileApplication = "application"
	ProfileAgent       = "agent"
	ProfilePortal      = "portal"
	ProfileEmbedded    = "embedded"
)

// ProfileConstraints defines what a profile permits.
type ProfileConstraints struct {
	Name              string   `json:"name"`
	AllowedLayouts    []string `json:"allowedLayouts"`
	AllowedNamespaces []string `json:"allowedNamespaces"`
	RequiredSlots     []string `json:"requiredSlots,omitempty"`
	MaxDepth          int      `json:"maxDepth,omitempty"`
}
