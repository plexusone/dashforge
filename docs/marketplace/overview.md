# Template Marketplace

Dashforge includes a marketplace for publishing and purchasing dashboard templates.

## Overview

The marketplace enables:

- **Publishers** to create and sell dashboard templates
- **Consumers** to discover and license templates for their organizations
- **Seat-based licensing** for team access management

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Template Marketplace                     │
├─────────────────────────────────────────────────────────────┤
│  Publishers                    │  Consumers                  │
│  ├── Create Templates          │  ├── Browse Marketplace     │
│  ├── Version Control           │  ├── Purchase Licenses      │
│  ├── Set Pricing               │  ├── Assign Seats           │
│  └── Track Analytics           │  └── Install Templates      │
└─────────────────────────────────────────────────────────────┘
```

## Key Concepts

### Dashboard Templates

Reusable dashboard configurations that can be:

- Published to the marketplace
- Versioned with semantic versioning
- Customized after installation
- Auto-updated when new versions release

### Publishers

Organizations that create templates. Publishers have:

- Creator roles for template development
- Reviewer roles for quality control
- Analytics on template performance
- Revenue from template sales

### Listings

Marketplace entries for templates including:

- Pricing (free, one-time, subscription, per-seat)
- Preview screenshots and descriptions
- Tags and categories for discovery
- Install counts and ratings

### Licenses

Access grants for purchased templates:

- Seat-based for team distribution
- Auto-renewal options
- Usage tracking

## Getting Started

### As a Publisher

1. **Become a Publisher** - Contact platform admin or self-register
2. **Create Templates** - Build dashboard templates in the builder
3. **Publish Listing** - Submit for marketplace approval
4. **Track Performance** - Monitor installs and revenue

### As a Consumer

1. **Browse Marketplace** - Discover templates by category/tag
2. **Preview Templates** - View screenshots and descriptions
3. **Purchase License** - Select license type and complete checkout
4. **Install Template** - Deploy to your organization
5. **Manage Seats** - Assign access to team members

## Template Lifecycle

```
Draft → Under Review → Published → Archived
         ↓
      Rejected (with feedback)
```

| Status | Description |
|--------|-------------|
| `draft` | Template in development |
| `under_review` | Submitted for marketplace approval |
| `published` | Live in marketplace |
| `archived` | Removed from marketplace |

## Pricing Models

| Model | Description | Use Case |
|-------|-------------|----------|
| `free` | No charge | Community templates, lead gen |
| `one_time` | Single purchase | Simple templates |
| `subscription` | Monthly/annual fee | Premium with updates |
| `per_seat` | Per-user pricing | Team templates |

## Data Flow

```
Publisher                    Marketplace                   Consumer
    │                            │                            │
    │ 1. Create Template         │                            │
    │ ─────────────────────────> │                            │
    │                            │                            │
    │ 2. Publish Listing         │                            │
    │ ─────────────────────────> │                            │
    │                            │                            │
    │                            │ 3. Browse/Search           │
    │                            │ <─────────────────────────  │
    │                            │                            │
    │                            │ 4. Purchase License        │
    │                            │ <─────────────────────────  │
    │                            │                            │
    │ 5. Revenue Payout          │ 6. License + Template      │
    │ <───────────────────────── │ ─────────────────────────> │
    │                            │                            │
```

## Next Steps

- [Publishing Templates](publishing.md) - Create and publish templates
- [Licensing](licensing.md) - Manage template licenses and seats
