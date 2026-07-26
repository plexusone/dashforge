package uispec

// NavigationSpec defines page-level navigation (routes, menus, breadcrumbs).
type NavigationSpec struct {
	Items      []NavItem `json:"items,omitempty"`
	Breadcrumb []NavItem `json:"breadcrumb,omitempty"`
	Position   string    `json:"position,omitempty"`
}

// NavItem is a single navigation entry.
type NavItem struct {
	ID       string    `json:"id"`
	Label    string    `json:"label"`
	Icon     string    `json:"icon,omitempty"`
	Href     string    `json:"href,omitempty"`
	Children []NavItem `json:"children,omitempty"`
	Badge    string    `json:"badge,omitempty"`
}
