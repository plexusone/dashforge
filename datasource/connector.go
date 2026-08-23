package datasource

import (
	"context"
	"fmt"
	"sync"

	"github.com/plexusone/dashforge/pkg/expression"
	"github.com/plexusone/dashforge/uispec"
)

// Connector is a high-level interface for UISpec data bindings.
// Unlike Provider/Connection (which are SQL-oriented), a Connector
// maps named operations to results — suitable for REST APIs, RPC, etc.
type Connector interface {
	// ConnectorID returns the data source identifier used in bindings.
	ConnectorID() string

	// Execute runs an operation and returns the result data.
	Execute(ctx context.Context, operation string, params map[string]any) (any, error)
}

// ConnectorRegistry manages Connector instances for UISpec binding resolution.
type ConnectorRegistry struct {
	mu         sync.RWMutex
	connectors map[string]Connector
}

// NewConnectorRegistry creates an empty connector registry.
func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: make(map[string]Connector),
	}
}

// RegisterConnector adds a connector. Returns an error if the ID is empty.
func (r *ConnectorRegistry) RegisterConnector(c Connector) error {
	if c.ConnectorID() == "" {
		return fmt.Errorf("datasource: connector ID is required")
	}
	r.mu.Lock()
	r.connectors[c.ConnectorID()] = c
	r.mu.Unlock()
	return nil
}

// GetConnector retrieves a connector by ID.
func (r *ConnectorRegistry) GetConnector(id string) (Connector, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.connectors[id]
	return c, ok
}

// Resolve evaluates all UISpec bindings in a component's data map,
// using the expression evaluator for parameter interpolation and
// registered connectors for data fetching.
func (r *ConnectorRegistry) Resolve(ctx context.Context, bindings map[string]uispec.Binding, exprCtx map[string]any) (map[string]any, error) {
	results := make(map[string]any, len(bindings))

	for name, binding := range bindings {
		resolvedParams, err := resolveBindingParams(binding.Parameters, exprCtx)
		if err != nil {
			return nil, fmt.Errorf("datasource: binding %q: resolve params: %w", name, err)
		}

		c, ok := r.GetConnector(binding.Source)
		if !ok {
			return nil, fmt.Errorf("datasource: binding %q: unknown source %q", name, binding.Source)
		}

		result, err := c.Execute(ctx, binding.Operation, resolvedParams)
		if err != nil {
			return nil, fmt.Errorf("datasource: binding %q: execute %q: %w", name, binding.Operation, err)
		}

		if binding.Transform != "" {
			transformed, err := applyTransform(result, binding.Transform, exprCtx)
			if err != nil {
				return nil, fmt.Errorf("datasource: binding %q: transform: %w", name, err)
			}
			result = transformed
		}

		if result == nil && binding.Default != nil {
			result = binding.Default
		}

		results[name] = result
	}

	return results, nil
}

func resolveBindingParams(params map[string]any, exprCtx map[string]any) (map[string]any, error) {
	if len(params) == 0 {
		return nil, nil
	}
	resolved := make(map[string]any, len(params))
	for k, v := range params {
		s, ok := v.(string)
		if ok && expression.ContainsExpression(s) {
			val, err := expression.Evaluate(s, exprCtx)
			if err != nil {
				return nil, fmt.Errorf("param %q: %w", k, err)
			}
			resolved[k] = val
		} else {
			resolved[k] = v
		}
	}
	return resolved, nil
}

func applyTransform(data any, transformExpr string, exprCtx map[string]any) (any, error) {
	ctx := make(map[string]any, len(exprCtx)+1)
	for k, v := range exprCtx {
		ctx[k] = v
	}
	ctx["data"] = data

	if expression.ContainsExpression(transformExpr) {
		return expression.Evaluate(transformExpr, ctx)
	}
	return data, nil
}
