package api

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grokify/grokifyql"
	systemauthz "github.com/grokify/systemforge/authz"
	sessionjwt "github.com/grokify/systemforge/session/jwt"
	sessionmw "github.com/grokify/systemforge/session/middleware"
	"github.com/plexusone/uiforge/dashboardir"
	localauthz "github.com/plexusone/uiforge/internal/authz"
	serveranalytics "github.com/plexusone/uiforge/internal/server/analytics"
)

func TestCompileGrokifyQLQuestionDefaultPolicyAllowsNoLimit(t *testing.T) {
	if _, err := compileGrokifyQLQuestion("SELECT name FROM items", defaultGrokifyQLPolicy()); err != nil {
		t.Fatalf("compileGrokifyQLQuestion() error = %v", err)
	}
}

func TestSystemForgeGrokifyQLPolicyProviderDeniesUnauthorizedField(t *testing.T) {
	sourceID := "omniroadmap"
	provider := SystemForgeGrokifyQLPolicyProvider{
		Authorizer: mockSystemForgeAuthorizer{allowed: map[string]bool{
			resourceKey(localauthz.ResourceTypeAnalyticsDataset, analyticsResourceID(sourceID, "items", ""), "read"):   true,
			resourceKey(localauthz.ResourceTypeAnalyticsField, analyticsResourceID(sourceID, "items", "name"), "read"): true,
			resourceKey(localauthz.ResourceTypeAnalyticsField, analyticsResourceID(sourceID, "items", "name"), "list"): true,
			resourceKey(localauthz.ResourceTypeAnalyticsField, analyticsResourceID(sourceID, "items", "name"), "sort"): true,
		}},
		Analytics: serveranalytics.NewService(staticCatalogProvider{catalog: dashboardir.AnalyticsCatalog{
			Sources: []dashboardir.AnalyticsSource{{
				ID: sourceID,
				Datasets: []dashboardir.AnalyticsDataset{{
					ID:        "items",
					QueryName: "items",
					Fields: []dashboardir.AnalyticsField{
						{ID: "name", QueryName: "name", Type: dashboardir.AnalyticsFieldTypeString, Selectable: true, Filterable: true, Sortable: true},
						{ID: "score", QueryName: "score", Type: dashboardir.AnalyticsFieldTypeNumber, Selectable: true, Filterable: true, Sortable: true},
					},
				}},
			}},
		}}),
	}
	ctx := sessionmw.ContextWithClaims(context.Background(), &sessionjwt.Claims{PrincipalID: uuid.New()})
	policy, err := provider.Policy(ctx, sourceID)
	if err != nil {
		t.Fatalf("Policy() error = %v", err)
	}
	query, err := grokifyql.Parse("SELECT score FROM items")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if issues := grokifyql.CheckPolicy(query, policy); len(issues) == 0 {
		t.Fatalf("CheckPolicy() issues = 0, want unauthorized score field")
	}
}

type staticCatalogProvider struct {
	catalog dashboardir.AnalyticsCatalog
}

func (p staticCatalogProvider) Catalog(context.Context) (dashboardir.AnalyticsCatalog, error) {
	return p.catalog, nil
}

func (p staticCatalogProvider) Close() error {
	return nil
}

type mockSystemForgeAuthorizer struct {
	allowed map[string]bool
}

func (m mockSystemForgeAuthorizer) Can(_ context.Context, _ systemauthz.Principal, action systemauthz.Action, resource systemauthz.Resource) (bool, error) {
	if resource.ID == nil {
		return false, nil
	}
	return m.allowed[resourceKey(string(resource.Type), *resource.ID, string(action))], nil
}

func (m mockSystemForgeAuthorizer) CanAll(ctx context.Context, principal systemauthz.Principal, actions []systemauthz.Action, resource systemauthz.Resource) (bool, error) {
	for _, action := range actions {
		ok, err := m.Can(ctx, principal, action, resource)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (m mockSystemForgeAuthorizer) CanAny(ctx context.Context, principal systemauthz.Principal, actions []systemauthz.Action, resource systemauthz.Resource) (bool, error) {
	for _, action := range actions {
		ok, err := m.Can(ctx, principal, action, resource)
		if err != nil || ok {
			return ok, err
		}
	}
	return false, nil
}

func (m mockSystemForgeAuthorizer) Filter(ctx context.Context, principal systemauthz.Principal, action systemauthz.Action, resources []systemauthz.Resource) ([]systemauthz.Resource, error) {
	var out []systemauthz.Resource
	for _, resource := range resources {
		ok, err := m.Can(ctx, principal, action, resource)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, resource)
		}
	}
	return out, nil
}

func resourceKey(resourceType string, id uuid.UUID, action string) string {
	return resourceType + ":" + id.String() + ":" + action
}
