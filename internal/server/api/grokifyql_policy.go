package api

import (
	"context"
	"crypto/sha1" //nolint:gosec // G505: non-cryptographic deterministic ID derivation (see analyticsResourceID)
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/grokify/guardsql"
	systemauthz "github.com/grokify/systemforge/authz"
	"github.com/grokify/systemforge/authzguardsql"
	serveranalytics "github.com/plexusone/dashforge/analytics"
	"github.com/plexusone/dashforge/dashboardir"
	localauthz "github.com/plexusone/dashforge/internal/authz"
	serverauth "github.com/plexusone/dashforge/internal/server/auth"
)

// GrokifyQLPolicyProvider builds request-scoped GrokifyQL policies.
type GrokifyQLPolicyProvider interface {
	Policy(ctx context.Context, sourceID string) (guardsql.Policy, error)
}

type staticGrokifyQLPolicyProvider struct{}

func (staticGrokifyQLPolicyProvider) Policy(context.Context, string) (guardsql.Policy, error) {
	return defaultGrokifyQLPolicy(), nil
}

// SystemForgeGrokifyQLPolicyProvider converts DashForge analytics metadata into a
// GrokifyQL schema, then compiles SystemForge authorization checks into an AST
// policy. It is intentionally request-scoped because SpiceDB decisions depend on
// the authenticated principal.
type SystemForgeGrokifyQLPolicyProvider struct {
	Authorizer systemauthz.Authorizer
	Analytics  *serveranalytics.Service
}

// Policy implements GrokifyQLPolicyProvider.
func (p SystemForgeGrokifyQLPolicyProvider) Policy(ctx context.Context, sourceID string) (guardsql.Policy, error) {
	principalID := serverauth.PrincipalIDFromContext(ctx)
	if p.Authorizer == nil || p.Analytics == nil || principalID == uuid.Nil {
		return defaultGrokifyQLPolicy(), nil
	}
	catalog, err := p.Analytics.Catalog(ctx)
	if err != nil {
		return guardsql.Policy{}, err
	}
	schema, err := grokifyQLSchemaForSource(catalog, sourceID)
	if err != nil {
		return guardsql.Policy{}, err
	}
	return authzguardsql.PolicyBuilder{
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

func defaultGrokifyQLPolicy() guardsql.Policy {
	return guardsql.Policy{
		AllowedOps:  []guardsql.Operation{guardsql.OperationRead},
		MaxDepth:    defaultGrokifyQLMaxDepth,
		MaxNodes:    defaultGrokifyQLMaxNodes,
		MaxInValues: defaultGrokifyQLMaxInValues,
	}
}

func grokifyQLSchemaForSource(catalog dashboardir.AnalyticsCatalog, sourceID string) (guardsql.Schema, error) {
	for _, source := range catalog.Sources {
		if !strings.EqualFold(source.ID, sourceID) {
			continue
		}
		schema := guardsql.Schema{Entities: map[string]guardsql.Entity{}}
		for _, dataset := range source.Datasets {
			queryName := strings.TrimSpace(dataset.QueryName)
			if queryName == "" {
				queryName = dataset.ID
			}
			entity := guardsql.Entity{Name: queryName, Fields: map[string]guardsql.Field{}}
			for _, field := range dataset.Fields {
				fieldName := strings.TrimSpace(field.QueryName)
				if fieldName == "" {
					fieldName = field.ID
				}
				entity.Fields[fieldName] = guardsql.Field{
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
	return guardsql.Schema{}, fmt.Errorf("analytics source %q not found", sourceID)
}

func grokifyQLFieldType(fieldType string) guardsql.FieldType {
	switch strings.ToLower(strings.TrimSpace(fieldType)) {
	case dashboardir.AnalyticsFieldTypeNumber:
		return guardsql.FieldNumber
	case dashboardir.AnalyticsFieldTypeBool:
		return guardsql.FieldBool
	case dashboardir.AnalyticsFieldTypeDate:
		return guardsql.FieldTime
	default:
		return guardsql.FieldString
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
	key := "dashforge:analytics:" + normalizeResourcePart(sourceID) + ":" + normalizeResourcePart(entity)
	if strings.TrimSpace(field) != "" {
		key += ":" + normalizeResourcePart(field)
	}
	sum := sha1.Sum([]byte(key)) //nolint:gosec // G401: deterministic resource-ID fingerprint, not a security primitive; changing the derivation would break persisted authz resource IDs
	return uuid.NewSHA1(uuid.NameSpaceURL, sum[:])
}

func normalizeResourcePart(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
