# Publishing Templates

Guide for creating and publishing dashboard templates to the marketplace.

## Prerequisites

- Publisher organization membership
- Creator or Admin role within the publisher org

## Creating a Template

### 1. Build Your Dashboard

Create a dashboard using the visual builder or Dashboard IR:

```json
{
  "id": "sales-template",
  "title": "Sales Analytics Dashboard",
  "layout": { "type": "grid", "columns": 12, "rowHeight": 80 },
  "dataSources": [
    {
      "id": "sales",
      "type": "cube",
      "query": {
        "measures": ["Orders.totalRevenue", "Orders.count"],
        "dimensions": ["Orders.createdAt"],
        "timeDimensions": [{
          "dimension": "Orders.createdAt",
          "granularity": "month"
        }]
      }
    }
  ],
  "widgets": [
    {
      "id": "revenue-chart",
      "type": "chart",
      "position": { "x": 0, "y": 0, "w": 8, "h": 4 },
      "dataSourceId": "sales",
      "config": {
        "geometry": "line",
        "encodings": { "x": "createdAt", "y": "totalRevenue" }
      }
    }
  ]
}
```

### 2. Create Template Record

```http
POST /api/v1/publishers/{publisher_id}/templates
Content-Type: application/json
Authorization: Bearer {token}

{
  "name": "Sales Analytics Dashboard",
  "description": "Comprehensive sales analytics with revenue tracking, order trends, and customer insights",
  "dashboard_ir": { ... },
  "version": "1.0.0",
  "required_datasources": ["cube"],
  "tags": ["sales", "analytics", "revenue"],
  "category": "business"
}
```

### 3. Add Preview Assets

Upload thumbnail and screenshots:

```http
POST /api/v1/publishers/{publisher_id}/templates/{template_id}/assets
Content-Type: multipart/form-data

thumbnail: (file)
screenshots[]: (files)
```

## Template Configuration

### Required Data Sources

Specify what data sources the template needs:

| Type | Description |
|------|-------------|
| `cube` | Requires Cube.js semantic layer |
| `postgres` | Direct PostgreSQL connection |
| `mysql` | Direct MySQL connection |
| `api` | REST API data source |

### Variables

Make templates customizable with variables:

```json
{
  "variables": [
    {
      "name": "company_name",
      "type": "string",
      "label": "Company Name",
      "default": "Acme Corp"
    },
    {
      "name": "primary_color",
      "type": "color",
      "label": "Primary Color",
      "default": "#6366f1"
    },
    {
      "name": "date_range_days",
      "type": "number",
      "label": "Default Date Range (days)",
      "default": 30
    }
  ]
}
```

## Versioning

### Semantic Versioning

Templates use semver (MAJOR.MINOR.PATCH):

- **MAJOR** - Breaking changes (new required data sources)
- **MINOR** - New features (new widgets, optional features)
- **PATCH** - Bug fixes (layout adjustments, typos)

### Creating a New Version

```http
POST /api/v1/publishers/{publisher_id}/templates/{template_id}/versions
Content-Type: application/json

{
  "version": "1.1.0",
  "dashboard_ir": { ... },
  "changelog": "Added customer segmentation widget"
}
```

## Creating a Listing

### 1. Prepare Listing Details

```http
POST /api/v1/publishers/{publisher_id}/listings
Content-Type: application/json

{
  "template_id": "tmpl_xxx",
  "title": "Sales Analytics Dashboard",
  "description": "Track revenue, orders, and customer trends with this comprehensive sales dashboard",
  "pricing_model": "per_seat",
  "price": 4900,
  "currency": "USD",
  "preview_enabled": true,
  "tags": ["sales", "analytics", "revenue", "ecommerce"],
  "category": "business"
}
```

### 2. Submit for Review

```http
POST /api/v1/publishers/{publisher_id}/listings/{listing_id}/submit
```

### 3. Marketplace Review

The marketplace team reviews for:

- Template functionality
- Description accuracy
- Appropriate pricing
- No malicious code

### 4. Publication

Once approved, the listing goes live:

```http
POST /api/v1/publishers/{publisher_id}/listings/{listing_id}/publish
```

## Analytics

Track your template performance:

```http
GET /api/v1/publishers/{publisher_id}/analytics
```

Response:

```json
{
  "period": "30d",
  "templates": [
    {
      "template_id": "tmpl_xxx",
      "name": "Sales Analytics Dashboard",
      "views": 1250,
      "installs": 89,
      "licenses_sold": 34,
      "revenue": 166600,
      "conversion_rate": 0.027
    }
  ],
  "totals": {
    "views": 3400,
    "installs": 245,
    "revenue": 485000
  }
}
```

## Best Practices

### Design

1. **Mobile Responsive** - Test on various screen sizes
2. **Performance** - Optimize queries for speed
3. **Accessibility** - Use proper color contrast and labels
4. **Documentation** - Include setup instructions

### Pricing

1. **Research Competition** - Check similar templates
2. **Value-Based** - Price based on value delivered
3. **Tier Options** - Consider multiple pricing tiers
4. **Free Tier** - Offer a free version for discovery

### Marketing

1. **Clear Title** - Descriptive, searchable name
2. **Compelling Screenshots** - Show key features
3. **Detailed Description** - Explain benefits clearly
4. **Tags** - Use relevant, searchable tags
5. **Updates** - Keep templates current

## Revenue

### Platform Fees

| Publisher Tier | Platform Fee |
|----------------|--------------|
| Standard | 15% |
| Pro | 10% |
| Enterprise | Negotiated |

### Payouts

- Monthly payouts via Stripe Connect
- Minimum payout threshold: $50
- Payout reports available in dashboard
