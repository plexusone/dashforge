package uispec

// Interaction defines a declarative event→condition→action rule.
type Interaction struct {
	When InteractionTrigger  `json:"when"`
	Then []InteractionAction `json:"then"`
}

// InteractionTrigger identifies the source event.
type InteractionTrigger struct {
	Component string `json:"component"`
	Event     string `json:"event"`
}

// InteractionAction describes what happens when the trigger fires.
type InteractionAction struct {
	Target    string         `json:"target"`
	Action    string         `json:"action"`
	Value     any            `json:"value,omitempty"`
	Condition string         `json:"condition,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}
