package uispec

// ThemeRef references a design-system theme for the page.
type ThemeRef struct {
	ID      string            `json:"id"`
	Variant string            `json:"variant,omitempty"`
	Tokens  map[string]string `json:"tokens,omitempty"`
}
