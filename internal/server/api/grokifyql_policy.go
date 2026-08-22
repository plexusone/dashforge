package api

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/grokify/grokifyql"
	"github.com/grokify/grokifyql/authzsystemforge"
	systemauthz "github.com/grokify/systemforge/authz"
	"github.com/plexusone/uiforge/dashboardir"
	localauthz "github.com/plexusone/uiforge/internal/authz"
	serveranalytics "github.com/plexusone/uiforge/internal/server/analytics"
	serverauth "github.com/plexusone/uiforge/internal/server/auth"
)

// GrokifyQLPolicyProvider builds request-scoped GrokifyQL policies.
type GrokifyQLPolicyProvider interface {
	Policy(ctx context.Context, sourceID string) (grokifyql.Policy, error)
}

type staticGrokifyQLPolicyProvider struct{}

func (staticGrokifyQLPolicyProvider) Policy(context.Context, string) (grokifyql.Policy, error) {
	return defaultGrokifyQLPolicy(), nil
}

// SystemForgeGrokifyQLPolicyProvider converts UIForge analytics metadata into a
// GrokifyQL schema, then compiles SystemForge authorization checks into an AST
// policy. It is intentionally request-scoped because SpiceDB decisions depend on
// the authenticated principal.
type SystemForgeGrokifyQLPolicyProvider struct {
	Authorizer systemauthz.Authorizer
	Analytics  *serveranalytics.Service
}

// Policy implements GrokifyQLPolicyProvider.
func (p SystemForgeGrokifyQLPolicyProvider) Policy(ctx context.Context, sourceID string) (grokifyql.Policy, error) {
	principalID := serverauth.PrincipalIDFromContext(ctx)
	if p.Authorizer == nil || p.Analytics == nil || principalID == uuid.Nil {
		return defaultGrokifyQLPolicy(), nil
	}
	catalog, err := p.Analytics.Catalog(ctx)
	if err != nil {
		return grokifyql.Policy{}, err
	}
	schema, err := grokifyQLSchemaForSource(catalog, sourceID)
	if err != nil {
		return grokifyql.Policy{}, err
	}
	return authzsystemforge.PolicyBuilder{
		Authorizer:      p.Authorizer,
		Principal:       systemauthz.NewUserPrincipal(principalID),
		Schema:          schema,
		ResourceBuilder: analyticsResourceBuilder{sourceID: sourceID},
		MaxDepth:        defaultGrokifyQLMaxDepth,
		MaxNodes:        defaultGrokifyQLMaxNodes,
		MaxInValues:     defaultGrokifyQLMaxInValues,
	}.Build(ctx)
}

const (
	defaultGrokifyQLMaxDepth    = 8
	defaultGrokifyQLMaxNodes    = 80
	defaultGrokifyQLMaxInValues = 100
)

func defaultGrokifyQLPolicy() grokifyql.Policy {
	return grokifyql.Policy{
		AllowedOps:  []grokifyql.Operation{grokifyql.OperationRead},
		MaxDepth:    defaultGrokifyQLMaxDepth,
		MaxNodes:    defaultGrokifyQLMaxNodes,
		MaxInValues: defaultGrokifyQLMaxInValues,
	}
}

func grokifyQLSchemaForSource(catalog dashboardir.AnalyticsCatalog, sourceID string) (grokifyql.Schema, error) {
	for _, source := range catalog.Sources {
		if !strings.EqualFold(source.ID, sourceID) {
			continue
		}
		schema := grokifyql.Schema{Entities: map[string]grokifyql.Entity{}}
		for _, dataset := range source.Datasets {
			queryName := strings.TrimSpace(dataset.QueryName)
			if queryName == "" {
				queryName = dataset.ID
			}
			entity := grokifyql.Entity{Name: queryName, Fields: map[string]grokifyql.Field{}}
			for _, field := range dataset.Fields {
				fieldName := strings.TrimSpace(field.QueryName)
				if fieldName == "" {
					fieldName = field.ID
				}
				entity.Fields[fieldName] = grokifyql.Field{
					Name:       fieldName,
					Type:       grokifyQLFieldType(field.Type),
					Selectable: field.Selectable,
					Filterable: field.Filterable,
					Sortable:   field.Sortable,
				}
			}
			schema.Entities[queryName] = entity
		}
		return schema, nil
	}
	return grokifyql.Schema{}, fmt.Errorf("analytics source %q not found", sourceID)
}

func grokifyQLFieldType(fieldType string) grokifyql.FieldType {
	switch strings.ToLower(strings.TrimSpace(fieldType)) {
	case dashboardir.AnalyticsFieldTypeNumber:
		return grokifyql.FieldNumber
	case dashboardir.AnalyticsFieldTypeBool:
		return grokifyql.FieldBool
	case dashboardir.AnalyticsFieldTypeDate:
		return grokifyql.FieldTime
	default:
		return grokifyql.FieldString
	}
}

type analyticsResourceBuilder struct {
	sourceID string
}

func (b analyticsResourceBuilder) EntityResource(entity string) systemauthz.Resource {
	id := analyticsResourceID(b.sourceID, entity, "")
	return systemauthz.NewResourceWithID(systemauthz.ResourceType(localauthz.ResourceTypeAnalyticsDataset), id)
}

func (b analyticsResourceBuilder) FieldResource(entity, field string) systemauthz.Resource {
	id := analyticsResourceID(b.sourceID, entity, field)
	return systemauthz.NewResourceWithID(systemauthz.ResourceType(localauthz.ResourceTypeAnalyticsField), id)
}

func analyticsResourceID(sourceID, entity, field string) uuid.UUID {
	key := "uiforge:analytics:" + normalizeResourcePart(sourceID) + ":" + normalizeResourcePart(entity)
	if strings.TrimSpace(field) != "" {
		key += ":" + normalizeResourcePart(field)
	}
	sum := sha1.Sum([]byte(key))
	return uuid.NewSHA1(uuid.NameSpaceURL, sum[:])
}

func normalizeResourcePart(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
