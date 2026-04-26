# Getting Started

This guide walks you through installing Dashforge and creating your first dashboard.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/plexusone/dashforge.git
cd dashforge

# Build the binaries
go build -o dashforge ./cmd/dashforge
go build -o dashforge-server ./cmd/dashforge-server

# Build the dashboard builder (optional)
cd builder && npm install && npm run build && cd ..
```

### Go Install

```bash
go install github.com/plexusone/dashforge/cmd/dashforge@latest
go install github.com/plexusone/dashforge/cmd/dashforge-server@latest
```

## Visual Builder Quick Start

The fastest way to create dashboards is with the visual builder.

### 1. Build the Builder

```bash
cd builder
npm install
npm run build
cd ..
```

### 2. Start the Server

```bash
go run ./cmd/dashforge-server serve --port 8080
```

### 3. Open the Builder

Navigate to `http://localhost:8080/builder/` in your browser.

### 4. Create Your Dashboard

1. Drag widgets from the left palette onto the canvas
2. Click a widget to edit its properties in the right panel
3. Configure chart types, data mappings, and styles
4. Click "Save" to save your dashboard

See [Dashboard Builder](builder.md) for detailed documentation.

## Static Mode Quick Start

For simple dashboards without a server.

### 1. Create a Dashboard JSON File

Create `my-dashboard.json`:

```json
{
  "id": "my-first-dashboard",
  "title": "My First Dashboard",
  "layout": {
    "type": "grid",
    "columns": 12,
    "rowHeight": 80
  },
  "dataSources": [
    {
      "id": "sample-data",
      "type": "inline",
      "data": [
        { "month": "Jan", "sales": 100, "profit": 20 },
        { "month": "Feb", "sales": 150, "profit": 35 },
        { "month": "Mar", "sales": 200, "profit": 50 },
        { "month": "Apr", "sales": 180, "profit": 40 }
      ]
    }
  ],
  "widgets": [
    {
      "id": "sales-metric",
      "type": "metric",
      "title": "Total Sales",
      "position": { "x": 0, "y": 0, "w": 3, "h": 2 },
      "datasourceId": "sample-data",
      "config": {
        "valueField": "sales",
        "format": "currency"
      }
    },
    {
      "id": "sales-chart",
      "type": "chart",
      "title": "Monthly Sales",
      "position": { "x": 3, "y": 0, "w": 9, "h": 4 },
      "datasourceId": "sample-data",
      "config": {
        "geometry": "bar",
        "encodings": { "x": "month", "y": "sales" },
        "style": { "showLegend": true }
      }
    }
  ]
}
```

### 2. Serve Locally

```bash
# Using Python's built-in server
cd dashforge
python3 -m http.server 8080

# Open in browser
open http://localhost:8080/viewer/?dashboard=../my-dashboard.json
```

## Server Mode Quick Start

For dynamic data sources, authentication, and multi-tenancy.

### 1. Set Up PostgreSQL

```bash
# Create database
createdb dashforge

# Set environment variables
export DATABASE_URL="postgres://localhost:5432/dashforge?sslmode=disable"
export JWT_SECRET="your-secure-secret-key-at-least-32-chars"
```

### 2. Start the Server

```bash
./dashforge-server serve \
  --port 8080 \
  --database-url "$DATABASE_URL" \
  --jwt-secret "$JWT_SECRET" \
  --auto-migrate
```

### 3. Access Dashforge

- **Builder**: `http://localhost:8080/builder/` - Visual dashboard editor
- **Viewer**: `http://localhost:8080/viewer/` - Dashboard viewer
- **API**: `http://localhost:8080/api/v1/` - REST API

## Adding Cube.js (Optional)

For semantic queries with business-friendly field names:

### 1. Configure Cube.js

```bash
cd cube
npm install

# Create .env file
cat > .env << EOF
CUBEJS_DB_TYPE=postgres
CUBEJS_DB_HOST=localhost
CUBEJS_DB_PORT=5432
CUBEJS_DB_NAME=your_database
CUBEJS_DB_USER=your_user
CUBEJS_DB_PASS=your_password
CUBEJS_API_SECRET=your-api-secret
EOF
```

### 2. Start Cube.js

```bash
npm run dev
# Cube.js API available at http://localhost:4000
```

### 3. Use in Builder

The builder automatically connects to Cube.js for schema introspection and query building.

See [Cube.js Integration](cube-integration.md) for details.

## Adding OAuth (Optional)

To enable GitHub and Google login:

```bash
export GITHUB_CLIENT_ID="your-github-client-id"
export GITHUB_CLIENT_SECRET="your-github-client-secret"
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
export BASE_URL="http://localhost:8080"

./dashforge-server serve \
  --port 8080 \
  --database-url "$DATABASE_URL" \
  --jwt-secret "$JWT_SECRET" \
  --auto-migrate
```

## Next Steps

- [Dashboard Builder](builder.md) - Visual editor guide
- [Cube.js Integration](cube-integration.md) - Semantic data layer
- [AI Features](ai-features.md) - LLM-powered generation
- [Dashboard IR Reference](dashboard-ir.md) - Learn the full JSON schema
- [Widgets](widgets.md) - Available widget types
- [Data Sources](data-sources.md) - Connect to your data
- [Server Configuration](server-config.md) - Production deployment
