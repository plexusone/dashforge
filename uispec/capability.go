package uispec

// CapabilitySpec declares permissions a component requires.
type CapabilitySpec struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}
