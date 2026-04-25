# Template Licensing

Managing licenses and seat assignments for purchased templates.

## Overview

When an organization purchases a template, they receive a license that grants access to the template. Seat-based licenses allow distributing access to team members.

## License Types

| Type | Description | Best For |
|------|-------------|----------|
| `individual` | Single user | Personal use |
| `seat_based` | Fixed seats | Teams (5-50) |
| `team` | Department access | Departments (50-200) |
| `unlimited` | Organization-wide | Enterprise |

## Purchasing a License

### Browse Marketplace

```http
GET /api/v1/marketplace/templates
```

### View Template Details

```http
GET /api/v1/marketplace/templates/{template_id}
```

### Purchase License

```http
POST /api/v1/marketplace/templates/{template_id}/licenses
Content-Type: application/json
Authorization: Bearer {token}

{
  "organization_id": "org_xxx",
  "license_type": "seat_based",
  "max_seats": 10,
  "auto_update": true
}
```

Response:

```json
{
  "id": "lic_xxx",
  "template_id": "tmpl_xxx",
  "organization_id": "org_xxx",
  "license_type": "seat_based",
  "max_seats": 10,
  "assigned_seats": 0,
  "auto_update": true,
  "valid_from": "2024-01-15T00:00:00Z",
  "valid_until": "2025-01-15T00:00:00Z",
  "status": "active"
}
```

## Managing Seats

### List Assigned Seats

```http
GET /api/v1/organizations/{org}/licenses/{license_id}/seats
Authorization: Bearer {token}
```

### Assign a Seat

```http
POST /api/v1/organizations/{org}/licenses/{license_id}/seats
Content-Type: application/json

{
  "principal_id": "user_xxx"
}
```

### Bulk Assign Seats

```http
POST /api/v1/organizations/{org}/licenses/{license_id}/seats/bulk
Content-Type: application/json

{
  "principal_ids": ["user_1", "user_2", "user_3"]
}
```

### Revoke a Seat

```http
DELETE /api/v1/organizations/{org}/licenses/{license_id}/seats/{seat_id}
```

## Installing Templates

Once licensed, install the template to create dashboards:

### Install Template

```http
POST /api/v1/organizations/{org}/templates/{template_id}/install
Content-Type: application/json

{
  "name": "My Sales Dashboard",
  "variables": {
    "company_name": "Acme Corp",
    "primary_color": "#6366f1"
  }
}
```

Response:

```json
{
  "dashboard_id": "dash_xxx",
  "name": "My Sales Dashboard",
  "template_id": "tmpl_xxx",
  "template_version": "1.2.0",
  "created_at": "2024-01-15T10:00:00Z"
}
```

### List Installed Templates

```http
GET /api/v1/organizations/{org}/installed-templates
```

## Auto-Updates

When `auto_update` is enabled:

1. New template version is published
2. System checks for compatible updates
3. Installed dashboards are updated automatically
4. Notification sent to org admins

### Disable Auto-Update

```http
PATCH /api/v1/organizations/{org}/licenses/{license_id}
Content-Type: application/json

{
  "auto_update": false
}
```

### Manual Update

```http
POST /api/v1/organizations/{org}/dashboards/{dashboard_id}/update-template
Content-Type: application/json

{
  "version": "1.3.0"
}
```

## License Limits

### Dashboard Limits

Some licenses limit how many dashboards can be created:

```json
{
  "max_dashboards": 5,
  "current_dashboards": 3
}
```

### Check Entitlement

```http
GET /api/v1/organizations/{org}/licenses/{license_id}/entitlements
```

Response:

```json
{
  "can_install": true,
  "remaining_dashboards": 2,
  "remaining_seats": 7
}
```

## License Lifecycle

### Status

| Status | Description |
|--------|-------------|
| `active` | License is valid |
| `expired` | Past expiration date |
| `suspended` | Payment issue |
| `revoked` | Manually revoked |

### Expiration Handling

When a license expires:

1. No new dashboards can be created from template
2. Existing dashboards continue to work (read-only)
3. 30-day grace period for renewal
4. After grace period, template features disabled

### Renewal

```http
POST /api/v1/organizations/{org}/licenses/{license_id}/renew
Content-Type: application/json

{
  "period_months": 12
}
```

## Usage Reports

### License Usage

```http
GET /api/v1/organizations/{org}/licenses/{license_id}/usage
```

Response:

```json
{
  "license_id": "lic_xxx",
  "template_name": "Sales Analytics Dashboard",
  "seats": {
    "total": 10,
    "assigned": 7,
    "active_30d": 5
  },
  "dashboards": {
    "total": 3,
    "limit": 5
  },
  "activity": {
    "views_30d": 245,
    "unique_users_30d": 6
  }
}
```

## SpiceDB Integration

License access is enforced via SpiceDB:

```zed
definition template_license {
    relation organization: organization
    relation seat_holder: principal

    permission use = seat_holder + organization->admin_access
    permission manage = organization->manage
}
```

Access checks:

```go
// Check if user can use licensed template
authz.Can(ctx, "template_license", licenseID, "use")
```

## Best Practices

1. **Right-Size Licenses** - Start with seats you need, expand later
2. **Monitor Usage** - Track seat utilization
3. **Enable Auto-Update** - Stay current with bug fixes
4. **Review Permissions** - Regularly audit seat assignments
5. **Plan Renewals** - Set renewal reminders
