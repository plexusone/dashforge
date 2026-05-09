// Package authz provides SpiceDB authorization for DashForge.
package authz

import "github.com/grokify/systemforge/authz"

// DashForgeSchema defines the SpiceDB schema for DashForge resources.
// Supports a two-sided marketplace: publishers (template creators) and
// consumers (organizations using dashboards).
const DashForgeSchema = `
// =============================================================================
// PRINCIPALS
// =============================================================================

definition principal {}

// =============================================================================
// PLATFORM LEVEL
// =============================================================================

definition platform {
    relation admin: principal
    relation marketplace_moderator: principal

    permission super_admin = admin
    permission moderate_marketplace = admin + marketplace_moderator
    permission view_analytics = admin
}

// =============================================================================
// ORGANIZATIONS
// =============================================================================

// Publisher organization - creates and sells dashboard templates
definition publisher {
    relation owner: principal
    relation admin: principal
    relation creator: principal
    relation reviewer: principal

    // Membership hierarchy
    permission manage = owner + admin
    permission create = manage + creator
    permission review = manage + reviewer

    // Publisher operations
    permission delete = owner
    permission settings = manage
    permission billing = owner + admin
    permission view_revenue = manage

    // Template operations
    permission create_template = create
    permission publish_template = manage
    permission view_analytics = manage

    // Connector operations
    permission create_connector = create
    permission publish_connector = manage

    // Member management
    permission invite_member = manage
    permission remove_member = manage
    permission change_role = manage
}

// Consumer organization - subscribes to DashForge, builds dashboards
definition organization {
    relation owner: principal
    relation admin: principal
    relation editor: principal
    relation viewer: principal

    // Membership hierarchy
    permission manage = owner + admin
    permission edit = manage + editor
    permission view = edit + viewer

    // Organization operations
    permission delete = owner
    permission settings = manage
    permission billing = owner + admin

    // Resource creation
    permission create_dashboard = edit
    permission create_datasource = manage
    permission create_alert = edit
    permission create_integration = manage

    // Marketplace
    permission purchase_template = manage
    permission view_purchases = view

    // Member management
    permission invite_member = manage
    permission remove_member = manage
    permission change_role = manage
}

// =============================================================================
// DASHBOARDS (Consumer-owned)
// =============================================================================

definition dashboard {
    relation org: organization
    relation owner: principal
    relation editor: principal
    relation viewer: principal
    relation from_template: dashboard_template

    // Permissions
    permission manage = owner + org->admin
    permission edit = manage + editor + org->editor
    permission view = edit + viewer + org->viewer
    permission delete = manage
    permission share = manage
    permission export = edit
    permission publish = manage
}

definition dashboard_version {
    relation dashboard: dashboard

    permission view = dashboard->view
    permission create = dashboard->edit
    permission restore = dashboard->manage
}

// =============================================================================
// DASHBOARD TEMPLATES (Publisher-owned, sold on marketplace)
// =============================================================================

definition dashboard_template {
    relation publisher: publisher
    relation owner: principal
    relation reviewer: principal
    relation licensed_org: organization

    // Publisher permissions
    permission manage = owner + publisher->manage
    permission edit = manage + publisher->creator
    permission review = publisher->review
    permission publish = publisher->publish_template
    permission view_analytics = manage

    // Access via license
    permission use = licensed_org->edit
    permission view = use + edit
}

// =============================================================================
// DATA SOURCES & CONNECTORS
// =============================================================================

definition data_source {
    relation org: organization
    relation owner: principal
    relation connector: data_connector

    permission manage = owner + org->admin
    permission use = manage + org->editor
    permission view = use + org->viewer
    permission test_connection = manage
}

definition data_connector {
    relation publisher: publisher
    relation owner: principal
    relation licensed_org: organization

    permission manage = owner + publisher->manage
    permission edit = manage + publisher->creator
    permission publish = publisher->publish_connector

    permission use = licensed_org->manage
    permission view = use + edit
}

// =============================================================================
// SAVED QUERIES & ALERTS
// =============================================================================

definition saved_query {
    relation org: organization
    relation owner: principal
    relation shared_with: principal

    permission manage = owner
    permission execute = manage + shared_with + org->editor
    permission view = execute + org->viewer
}

definition alert {
    relation org: organization
    relation owner: principal
    relation dashboard: dashboard
    relation subscribers: principal

    permission manage = owner + org->admin
    permission subscribe = org->viewer
    permission view = manage + subscribers + org->viewer
}

// =============================================================================
// INTEGRATIONS (Notification channels)
// =============================================================================

definition integration {
    relation org: organization
    relation owner: principal

    permission manage = owner + org->admin
    permission use = manage + org->editor
    permission view = use + org->viewer
}

// =============================================================================
// MARKETPLACE
// =============================================================================

definition marketplace_listing {
    relation template: dashboard_template
    relation connector: data_connector
    relation publisher: publisher
    relation moderator: principal

    permission manage = publisher->manage
    permission moderate = moderator + platform->moderate_marketplace
    permission view = publisher->creator
    permission purchase = organization->purchase_template
}

definition template_license {
    relation template: dashboard_template
    relation organization: organization
    relation purchased_by: principal

    permission view = organization->manage + purchased_by
    permission use = organization->edit
    permission transfer = organization->manage
}

definition connector_license {
    relation connector: data_connector
    relation organization: organization
    relation purchased_by: principal

    permission view = organization->manage + purchased_by
    permission use = organization->manage
}
`

// =============================================================================
// Role Constants
// =============================================================================

// Publisher roles
const (
	PublisherRoleOwner    = "owner"
	PublisherRoleAdmin    = "admin"
	PublisherRoleCreator  = "creator"
	PublisherRoleReviewer = "reviewer"
)

// Organization (consumer) roles
const (
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleEditor = "editor"
	OrgRoleViewer = "viewer"
)

// =============================================================================
// Role Mappings
// =============================================================================

// PublisherRoleToRelation maps publisher role names to SpiceDB relation names.
var PublisherRoleToRelation = map[string]string{
	"owner":    "owner",
	"admin":    "admin",
	"creator":  "creator",
	"reviewer": "reviewer",
}

// OrgRoleToRelation maps organization role names to SpiceDB relation names.
var OrgRoleToRelation = map[string]string{
	"owner":  "owner",
	"admin":  "admin",
	"editor": "editor",
	"viewer": "viewer",
}

// RoleToRelation is the default mapping (for backward compatibility).
var RoleToRelation = OrgRoleToRelation

// RelationToRole maps SpiceDB relation names back to DashForge roles.
var RelationToRole = map[string]string{
	"owner":    "owner",
	"admin":    "admin",
	"editor":   "editor",
	"viewer":   "viewer",
	"creator":  "creator",
	"reviewer": "reviewer",
}

// =============================================================================
// Role Hierarchies
// =============================================================================

// PublisherRoleHierarchy defines role hierarchy for publishers.
var PublisherRoleHierarchy = authz.RoleHierarchy{
	PublisherRoleOwner:    100,
	PublisherRoleAdmin:    80,
	PublisherRoleCreator:  60,
	PublisherRoleReviewer: 40,
}

// OrgRoleHierarchy defines role hierarchy for consumer organizations.
var OrgRoleHierarchy = authz.RoleHierarchy{
	OrgRoleOwner:  100,
	OrgRoleAdmin:  80,
	OrgRoleEditor: 60,
	OrgRoleViewer: 40,
}

// RoleHierarchy is the default hierarchy (backward compatibility).
var RoleHierarchy = OrgRoleHierarchy

// =============================================================================
// Permission Constants
// =============================================================================

const (
	PermManage  = "manage"
	PermEdit    = "edit"
	PermView    = "view"
	PermCreate  = "create"
	PermDelete  = "delete"
	PermPublish = "publish"
	PermShare   = "share"
	PermExport  = "export"
	PermUse     = "use"
)

// =============================================================================
// Resource Types
// =============================================================================

const (
	ResourceTypePlatform           = "platform"
	ResourceTypePublisher          = "publisher"
	ResourceTypeOrganization       = "organization"
	ResourceTypeDashboard          = "dashboard"
	ResourceTypeDashboardVersion   = "dashboard_version"
	ResourceTypeDashboardTemplate  = "dashboard_template"
	ResourceTypeDataSource         = "data_source"
	ResourceTypeDataConnector      = "data_connector"
	ResourceTypeSavedQuery         = "saved_query"
	ResourceTypeAlert              = "alert"
	ResourceTypeIntegration        = "integration"
	ResourceTypeMarketplaceListing = "marketplace_listing"
	ResourceTypeTemplateLicense    = "template_license"
	ResourceTypeConnectorLicense   = "connector_license"
)
