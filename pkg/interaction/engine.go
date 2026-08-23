// Package interaction implements a declarative event→condition→action engine
// for UISpec interactions. It dispatches events through matching interaction
// rules, evaluates conditions via the expression evaluator, and executes
// registered action handlers.
package interaction

import (
	"context"
	"fmt"

	"github.com/plexusone/dashforge/pkg/expression"
	"github.com/plexusone/dashforge/pkg/state"
	"github.com/plexusone/dashforge/uispec"
)

// ActionHandler executes a single interaction action.
type ActionHandler func(ctx context.Context, action uispec.InteractionAction, eventData map[string]any) error

// Event represents a component event to be dispatched.
type Event struct {
	ComponentID string
	EventName   string
	Data        map[string]any
}

// Engine routes events through interaction rules and executes actions.
type Engine struct {
	store    *state.Store
	handlers map[string]ActionHandler
}

// New creates an Engine with built-in handlers (state.set, state.toggle,
// component.refresh, navigate) wired to the provided state store.
func New(store *state.Store) *Engine {
	e := &Engine{
		store:    store,
		handlers: make(map[string]ActionHandler),
	}
	e.registerBuiltins()
	return e
}

// RegisterHandler adds or replaces a named action handler.
func (e *Engine) RegisterHandler(action string, handler ActionHandler) {
	e.handlers[action] = handler
}

// Dispatch finds all interactions matching the event and executes their actions.
// Actions within a single interaction run sequentially; all matching
// interactions are evaluated.
func (e *Engine) Dispatch(ctx context.Context, interactions []uispec.Interaction, event Event) error {
	for _, ix := range interactions {
		if ix.When.Component != event.ComponentID || ix.When.Event != event.EventName {
			continue
		}

		exprCtx := e.buildExprContext(event)

		for _, action := range ix.Then {
			if action.Condition != "" {
				val, err := expression.Evaluate(action.Condition, exprCtx)
				if err != nil {
					return fmt.Errorf("interaction: evaluate condition %q: %w", action.Condition, err)
				}
				if !isTruthy(val) {
					continue
				}
			}

			resolved, err := e.resolveParams(action, exprCtx)
			if err != nil {
				return fmt.Errorf("interaction: resolve params for action %q: %w", action.Action, err)
			}

			handler, ok := e.handlers[action.Action]
			if !ok {
				return fmt.Errorf("interaction: unknown action %q", action.Action)
			}
			if err := handler(ctx, resolved, event.Data); err != nil {
				return fmt.Errorf("interaction: action %q: %w", action.Action, err)
			}

			// Refresh expression context after state-mutating actions.
			exprCtx = e.buildExprContext(event)
		}
	}
	return nil
}

func (e *Engine) buildExprContext(event Event) map[string]any {
	ctx := e.store.AsContext()
	ctx["event"] = event.Data
	return ctx
}

func (e *Engine) resolveParams(action uispec.InteractionAction, exprCtx map[string]any) (uispec.InteractionAction, error) {
	resolved := uispec.InteractionAction{
		Target:    action.Target,
		Action:    action.Action,
		Value:     action.Value,
		Condition: action.Condition,
	}

	if action.Params != nil {
		resolved.Params = make(map[string]any, len(action.Params))
		for k, v := range action.Params {
			s, ok := v.(string)
			if ok && expression.ContainsExpression(s) {
				val, err := expression.Evaluate(s, exprCtx)
				if err != nil {
					return resolved, err
				}
				resolved.Params[k] = val
			} else {
				resolved.Params[k] = v
			}
		}
	}

	if s, ok := action.Value.(string); ok && expression.ContainsExpression(s) {
		val, err := expression.Evaluate(s, exprCtx)
		if err != nil {
			return resolved, err
		}
		resolved.Value = val
	}

	return resolved, nil
}

func (e *Engine) registerBuiltins() {
	e.handlers["state.set"] = func(_ context.Context, action uispec.InteractionAction, _ map[string]any) error {
		path, ok := action.Params["path"].(string)
		if !ok {
			return fmt.Errorf("state.set: 'path' parameter must be a string")
		}
		value := action.Params["value"]
		if value == nil {
			value = action.Value
		}
		e.store.Set(path, value)
		return nil
	}

	e.handlers["state.toggle"] = func(_ context.Context, action uispec.InteractionAction, _ map[string]any) error {
		path, ok := action.Params["path"].(string)
		if !ok {
			path = action.Target
		}
		return e.store.Toggle(path)
	}

	e.handlers["component.refresh"] = func(_ context.Context, _ uispec.InteractionAction, _ map[string]any) error {
		// Refresh is a client-side concern; the Go engine records the intent.
		// In a full runtime, this would trigger a re-render of the target component.
		return nil
	}

	e.handlers["navigate"] = func(_ context.Context, _ uispec.InteractionAction, _ map[string]any) error {
		// Navigation is a client-side concern; the Go engine records the intent.
		return nil
	}
}

func isTruthy(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "" && val != "false"
	case int:
		return val != 0
	case float64:
		return val != 0
	case nil:
		return false
	default:
		return true
	}
}
