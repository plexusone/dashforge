// Package authz provides SpiceDB authorization for DashForge.
package authz

// DashForgeSchema defines the SpiceDB schema for DashForge resources.
// This extends CoreForge's BaseSchema with dashboard-specific permissions.
const DashForgeSchema = `
// Principal types (from CoreForge BaseSchema)
definition principal {}

// Organization with DashForge-specific roles
definition organization {
    relation owner: principal
    relation admin: principal
    relation editor: principal
    relation viewer: principal

    // Organization-level permissions
    permission delete = owner
    permission manage = owner + admin
    permission settings = manage
    permission edit = manage + editor
    permission view = edit + viewer

    // Member management
    permission invite_member = manage
    permission remove_member = manage
    permission change_role = manage

    // Resource management
    permission create_dashboard = edit
    permission create_datasource = manage
    permission create_integration = manage
}

// Platform for cross-org admin access
definition platform {
    relation admin: principal

    permission super_admin = admin
}

// Dashboard resource
definition dashboard {
    relation org: organization
    relation owner: principal
    relation editor: principal
    relation viewer: principal

    // Dashboard permissions
    permission manage = owner + org->admin
    permission edit = manage + editor + org->editor
    permission view = edit + viewer + org->viewer
    permission delete = owner + org->admin
    permission publish = manage
    permission share = manage
}

// Dashboard version
definition dashboard_version {
    relation dashboard: dashboard

    permission view = dashboard->view
    permission create = dashboard->edit
}

// Data source
definition data_source {
    relation org: organization
    relation owner: principal

    permission manage = owner + org->admin
    permission use = manage + org->editor
    permission view = use + org->viewer
}

// Saved query
definition saved_query {
    relation org: organization
    relation owner: principal
    relation shared_with: principal

    permission manage = owner
    permission execute = manage + shared_with + org->editor
    permission view = execute + org->viewer
}

// Integration (Slack, email, webhook)
definition integration {
    relation org: organization
    relation owner: principal

    permission manage = owner + org->admin
    permission use = manage + org->editor
    permission view = use + org->viewer
}

// Alert
definition alert {
    relation org: organization
    relation owner: principal
    relation dashboard: dashboard

    permission manage = owner + org->admin
    permission view = manage + org->viewer
}
`

// RoleToRelation maps DashForge role names to SpiceDB relation names.
var RoleToRelation = map[string]string{
	"owner":  "owner",
	"admin":  "admin",
	"editor": "editor",
	"viewer": "viewer",
}

// RelationToRole maps SpiceDB relation names back to DashForge roles.
var RelationToRole = map[string]string{
	"owner":  "owner",
	"admin":  "admin",
	"editor": "editor",
	"viewer": "viewer",
}
