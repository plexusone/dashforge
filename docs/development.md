# Development

This guide covers setting up a development environment and contributing to Dashforge.

## Prerequisites

- Go 1.22 or later
- Node.js 18+ (for builder and viewer development)
- PostgreSQL 14+ (for server mode)
- Docker (optional, for containerized development)

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/grokify/dashforge.git
cd dashforge
```

### Build

```bash
# Build all Go binaries
go build ./...

# Build specific binary
go build -o dashforge ./cmd/dashforge
go build -o dashforge-server ./cmd/dashforge-server

# Build the dashboard builder
cd builder && npm install && npm run build && cd ..
```

### Run Tests

```bash
# Go tests
go test -v ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Builder tests
cd builder && npm run typecheck && cd ..
```

### Lint

```bash
# Go
golangci-lint run

# Builder
cd builder && npm run lint && cd ..
```

## Project Structure

```
dashforge/
├── cmd/
│   ├── dashforge/           # Static CLI
│   └── dashforge-server/    # Full server
├── builder/                 # Dashboard builder (React)
│   ├── src/
│   │   ├── components/      # React components
│   │   │   ├── canvas/      # Grid layout canvas
│   │   │   ├── palette/     # Widget palette
│   │   │   ├── properties/  # Properties panel
│   │   │   ├── chart-builder/   # Chart configuration
│   │   │   ├── query-builder/   # Cube.js queries
│   │   │   ├── schema-browser/  # Schema explorer
│   │   │   └── widgets/     # Widget renderers
│   │   ├── stores/          # Zustand state
│   │   ├── api/             # API clients
│   │   ├── ai/              # AI generation
│   │   └── types/           # TypeScript types
│   ├── package.json
│   └── vite.config.ts
├── cube/                    # Cube.js semantic layer
│   ├── model/cubes/         # Data models (YAML)
│   └── cube.js              # Cube configuration
├── dashboardir/             # Dashboard JSON types
├── datasource/              # Data source providers
│   ├── providers/           # PostgreSQL, MySQL
│   ├── manager.go           # Connection pool
│   └── query.go             # Query executor
├── ent/                     # Ent ORM
│   └── schema/              # Entity schemas
├── internal/server/         # Server implementation
│   ├── api/                 # REST API handlers
│   ├── auth/                # JWT & OAuth
│   └── db/                  # Database layer
├── viewer/                  # Static HTML/JS viewer
├── docs/                    # MkDocs documentation
└── examples/                # Example dashboards
```

## Builder Development

The dashboard builder is a React/TypeScript application.

### Development Server

```bash
cd builder
npm install
npm run dev
# Opens http://localhost:5173 with hot reload
```

The dev server proxies `/api` requests to `http://localhost:8080`, so start the Dashforge server:

```bash
# In another terminal
go run ./cmd/dashforge-server serve --port 8080
```

### Building for Production

```bash
cd builder
npm run build
# Output in builder/dist/
```

The built files are embedded in the Go binary via `builder/embed.go`.

### Type Checking

```bash
cd builder
npm run typecheck
```

### Key Technologies

| Technology | Purpose |
|------------|---------|
| React 18 | UI framework |
| TypeScript | Type safety |
| Vite | Build tool |
| Tailwind CSS | Styling |
| Zustand | State management |
| React Query | Data fetching |
| react-grid-layout | Drag-and-drop grid |
| ECharts | Chart rendering |
| Cube.js Client | Semantic queries |

### Adding a New Widget Type

1. Add type to `builder/src/types/dashboard.ts`
2. Create widget component in `builder/src/components/widgets/`
3. Add to `WidgetContainer.tsx` switch statement
4. Add palette entry in `WidgetPalette.tsx`
5. Add config editor in `PropertiesPanel.tsx`

### Adding a New Chart Type

1. Add geometry to `ChartConfig` in `types/dashboard.ts`
2. Add option to `GeometryPicker.tsx`
3. Add rendering case in `ChartWidget.tsx`
4. Update `DataMapping.tsx` for encoding requirements

## Cube.js Development

### Setup

```bash
cd cube
npm install

# Create .env
cat > .env << EOF
CUBEJS_DB_TYPE=postgres
CUBEJS_DB_HOST=localhost
CUBEJS_DB_NAME=your_database
CUBEJS_DB_USER=your_user
CUBEJS_DB_PASS=your_password
EOF
```

### Running

```bash
cd cube
npm run dev
# API at http://localhost:4000
```

### Adding Data Models

Create YAML files in `cube/model/cubes/`:

```yaml
cubes:
  - name: MyTable
    sql: SELECT * FROM my_table

    measures:
      - name: count
        type: count

    dimensions:
      - name: id
        type: string
        primaryKey: true
```

## Server Development

### Making Changes

1. Create a feature branch
2. Make changes
3. Run tests and linter
4. Submit pull request

### Ent Schema Changes

When modifying Ent schemas:

```bash
# Edit schema files in ent/schema/

# Regenerate code
go generate ./ent

# Run migrations (dev database)
./dashforge-server serve --database-url "$DEV_DATABASE_URL" --auto-migrate
```

### Adding a New API Endpoint

1. Add handler in `internal/server/api/`
2. Register route in handler setup
3. Add documentation in `docs/api-reference.md`
4. Add tests

### Adding a New Data Source Provider

1. Create provider package in `datasource/providers/<name>/`
2. Implement the `Provider` interface
3. Register via `init()`
4. Import in `internal/server/server.go`

**Example: ClickHouse Provider**

```go
// datasource/providers/clickhouse/provider.go
package clickhouse

import (
    "context"
    "database/sql"

    _ "github.com/ClickHouse/clickhouse-go/v2"
    "github.com/grokify/dashforge/datasource"
)

func init() {
    datasource.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() string { return "clickhouse" }

func (p *Provider) Connect(ctx context.Context, cfg datasource.ConnectionConfig) (datasource.Connection, error) {
    db, err := sql.Open("clickhouse", cfg.ConnectionURL)
    if err != nil {
        return nil, err
    }
    return &Connection{db: db}, nil
}
```

## Local Development Setup

### Database

```bash
# Start PostgreSQL with Docker
docker run -d \
  --name dashforge-db \
  -e POSTGRES_USER=dashforge \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=dashforge \
  -p 5432:5432 \
  postgres:16-alpine

# Connection string
export DATABASE_URL="postgres://dashforge:password@localhost:5432/dashforge?sslmode=disable"
```

### Full Stack

```bash
# Terminal 1: Dashforge server
export JWT_SECRET="dev-secret-key-at-least-32-chars"
export DATABASE_URL="postgres://dashforge:password@localhost:5432/dashforge?sslmode=disable"
go run ./cmd/dashforge-server serve --auto-migrate

# Terminal 2: Builder dev server (optional, for hot reload)
cd builder && npm run dev

# Terminal 3: Cube.js (optional, for semantic queries)
cd cube && npm run dev
```

Access:

- Builder (dev): `http://localhost:5173`
- Builder (embedded): `http://localhost:8080/builder/`
- Viewer: `http://localhost:8080/viewer/`
- Cube.js Playground: `http://localhost:4000`

## Testing

### Unit Tests

```bash
go test -v ./internal/...
```

### Integration Tests

```bash
export TEST_DATABASE_URL="postgres://..."
go test -v -tags=integration ./...
```

### Builder Type Check

```bash
cd builder && npm run typecheck
```

## Code Style

### Go

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Keep functions focused and small
- Prefer returning errors over logging

### TypeScript

- Use TypeScript strict mode
- Prefer functional components
- Use React Query for data fetching
- Use Zustand for state management

### Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(builder): add scatter plot support
fix(api): handle null values in query results
docs: update builder documentation
chore(deps): update dependencies
```

## Documentation

### Building Docs

```bash
# Install MkDocs
pip install mkdocs-material

# Serve locally
mkdocs serve

# Build static site
mkdocs build
```

### Deploying Docs

```bash
mkdocs gh-deploy
```

## Release Process

1. Update version in relevant files
2. Update CHANGELOG.md
3. Build and test:
   ```bash
   go build ./...
   go test ./...
   cd builder && npm run build && npm run typecheck && cd ..
   ```
4. Create release commit and tag
5. Push to trigger release workflow

```bash
git add -A
git commit -m "chore: release v1.2.0"
git tag v1.2.0
git push origin main --tags
```

## Debugging

### Server Logging

```bash
export LOG_LEVEL=debug
./dashforge-server serve
```

### Builder DevTools

React DevTools and Redux DevTools (for Zustand) are available in development mode.

### Cube.js Playground

Access `http://localhost:4000` for the Cube.js development playground with query builder and schema explorer.

## Common Issues

### Builder Build Fails

```bash
# Clear node_modules and reinstall
cd builder
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Embed Error (no matching files)

The `dist/` directory must exist for Go embed to work:

```bash
cd builder && npm run build && cd ..
```

### Ent Generation Fails

```bash
rm -rf ent/*.go ent/*/
go generate ./ent
```

### Database Connection Issues

```bash
psql "$DATABASE_URL" -c "SELECT 1"
docker logs dashforge-db
```

## Getting Help

- [GitHub Issues](https://github.com/grokify/dashforge/issues)
- [Discussions](https://github.com/grokify/dashforge/discussions)
