package registry

import "encoding/json"

// RegisterAssistantComponents adds the assistant namespace components to the registry:
// thread, composer, thread-list, tool-call, run-status.
func RegisterAssistantComponents(r *Registry) error {
	specs := []*ComponentSpec{
		{
			ID:       "assistant.thread",
			Version:  "1.0.0",
			Category: "assistant",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"showToolCalls":      {"type": "boolean"},
					"showTimestamps":     {"type": "boolean"},
					"markdownEnabled":    {"type": "boolean"},
					"streamingIndicator": {"type": "boolean"}
				}
			}`),
			DataInputs: map[string]DataInput{
				"messages": {Type: "message-stream", Description: "Message stream from runtime", Required: true},
			},
			Events: map[string]EventDef{
				"messageAdded": {Description: "New message added to thread"},
				"scrollEnd":    {Description: "User scrolled to end of thread"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "300px", MinHeight: "200px"},
		},
		{
			ID:       "assistant.composer",
			Version:  "1.0.0",
			Category: "assistant",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"placeholder":       {"type": "string"},
					"showAttachments":   {"type": "boolean"},
					"showModelSelector": {"type": "boolean"},
					"maxLength":         {"type": "integer"}
				}
			}`),
			Events: map[string]EventDef{
				"submit": {Description: "Message submitted"},
				"cancel": {Description: "Generation cancelled"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "300px"},
		},
		{
			ID:       "assistant.thread-list",
			Version:  "1.0.0",
			Category: "assistant",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"showSearch":    {"type": "boolean"},
					"showNewButton": {"type": "boolean"},
					"showFolders":   {"type": "boolean"},
					"showDelete":    {"type": "boolean"}
				}
			}`),
			DataInputs: map[string]DataInput{
				"conversations": {Type: "conversation-list", Description: "List of conversations", Required: true},
			},
			Events: map[string]EventDef{
				"select": {Description: "Thread selected"},
				"create": {Description: "New thread created"},
				"delete": {Description: "Thread deleted"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "200px"},
		},
		{
			ID:       "assistant.tool-call",
			Version:  "1.0.0",
			Category: "assistant",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"showArgs":    {"type": "boolean"},
					"showResult":  {"type": "boolean"},
					"collapsible": {"type": "boolean"}
				}
			}`),
			Events: map[string]EventDef{
				"toggle": {Description: "Expand/collapse toggled"},
			},
		},
		{
			ID:       "assistant.run-status",
			Version:  "1.0.0",
			Category: "assistant",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"showElapsed":    {"type": "boolean"},
					"showTokenCount": {"type": "boolean"},
					"compact":        {"type": "boolean"}
				}
			}`),
		},
	}

	for _, s := range specs {
		if err := r.Register(s); err != nil {
			return err
		}
	}
	return nil
}
