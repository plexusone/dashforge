# Dashboard Builder

The Dashforge Dashboard Builder is a visual drag-and-drop editor for creating dashboards. It's built with React and TypeScript, and is optimized for both human users and LLM agents.

## Getting Started

### Building the Builder

```bash
cd builder
npm install
npm run build
cd ..
```

### Running in Development Mode

For hot-reloading during development:

```bash
cd builder
npm run dev
# Opens http://localhost:5173
```

The dev server proxies API requests to `http://localhost:8080`, so start the Dashforge server in another terminal:

```bash
go run ./cmd/dashforge-server serve --port 8080
```

### Running in Production

Build and embed in the Dashforge server:

```bash
cd builder && npm run build && cd ..
go run ./cmd/dashforge-server serve --port 8080
# Builder available at http://localhost:8080/builder/
```

## Interface Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  [DF] Sales Dashboard *                    [Undo] [Redo] [Save] │
├─────────────────────────────────────────────────────────────────┤
│ ┌──────────┐ ┌─────────────────────────┐ ┌───────────────────┐ │
│ │ Widgets  │ │                         │ │ Properties        │ │
│ │          │ │     Canvas              │ │                   │ │
│ │ ▢ Bar    │ │                         │ │ Title: [Revenue]  │ │
│ │ ▢ Line   │ │   ┌─────────────────┐   │ │                   │ │
│ │ ▢ Area   │ │   │  Revenue Chart  │   │ │ Chart Type:       │ │
│ │ ▢ Pie    │ │   │  [selected]     │   │ │ [Line] [Bar]      │ │
│ │ ▢ Scatter│ │   └─────────────────┘   │ │ [Pie] [Scatter]   │ │
│ │          │ │                         │ │                   │ │
│ │ ▢ Metric │ │   ┌──────┐ ┌──────┐    │ │ X Axis: [date]    │ │
│ │ ▢ Table  │ │   │KPI 1 │ │KPI 2 │    │ │ Y Axis: [revenue] │ │
│ │ ▢ Text   │ │   └──────┘ └──────┘    │ │                   │ │
│ │          │ │                         │ │ [✓] Show Legend   │ │
│ └──────────┘ └─────────────────────────┘ └───────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Widget Palette

Drag widgets from the left sidebar onto the canvas.

### Chart Widgets

| Widget | Description | Default Size |
|--------|-------------|--------------|
| Bar Chart | Categorical comparisons | 4x3 |
| Line Chart | Time series and trends | 6x3 |
| Area Chart | Stacked time series | 6x3 |
| Pie Chart | Part-to-whole relationships | 4x3 |
| Scatter Plot | Correlations between variables | 4x3 |

### Data Widgets

| Widget | Description | Default Size |
|--------|-------------|--------------|
| Metric | Single KPI with optional sparkline | 2x2 |
| Table | Tabular data with sorting/pagination | 6x4 |

### Content Widgets

| Widget | Description | Default Size |
|--------|-------------|--------------|
| Text | Markdown or plain text content | 4x2 |
| Image | Static images | 3x3 |

## Canvas

The canvas uses a 12-column grid layout, similar to Bootstrap:

- **Grid columns**: 12 (configurable)
- **Row height**: 80px (configurable)
- **Gap**: 8px between widgets
- **Snap-to-grid**: Widgets snap to grid positions

### Widget Operations

- **Select**: Click a widget to edit its properties
- **Move**: Drag the grip handle in the widget header
- **Resize**: Drag the bottom-right corner
- **Delete**: Click the trash icon in the header
- **Duplicate**: Click the copy icon in the header

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl+Z` | Undo |
| `Ctrl+Shift+Z` | Redo |
| `Delete` | Delete selected widget |
| `Ctrl+D` | Duplicate selected widget |

## Properties Panel

When a widget is selected, the properties panel shows configuration options.

### General Properties

All widgets have:

- **Title**: Display title shown in the widget header
- **Description**: Optional description (tooltip)
- **Data Source**: Which data source to use

### Chart Properties

Chart widgets have additional options:

- **Chart Type**: line, bar, pie, scatter, area
- **Data Mapping**:
    - X Axis / Category field
    - Y Axis / Value field
    - Color field (optional)
- **Style Options**:
    - Color palette selection
    - Show/hide legend
    - Legend position (top, bottom, left, right)
    - Smooth lines (line/area charts)
    - Stacked (bar/area charts)
    - Horizontal (bar charts)

### Metric Properties

- **Value Field**: The field containing the metric value
- **Format**: number, currency, percent, compact
- **Prefix/Suffix**: Text to display before/after the value
- **Comparison**: Show change vs previous period
- **Sparkline**: Show mini trend chart

### Table Properties

- **Columns**: Configure visible columns
- **Pagination**: Enable with page size
- **Sortable**: Allow column sorting
- **Filterable**: Allow column filtering

## Data Sources

The builder supports multiple data source types:

### Inline Data

Embed data directly in the dashboard JSON:

```json
{
  "id": "sample-data",
  "type": "inline",
  "data": [
    { "month": "Jan", "revenue": 12500 },
    { "month": "Feb", "revenue": 15700 }
  ]
}
```

### URL Data

Fetch data from an HTTP endpoint:

```json
{
  "id": "api-data",
  "type": "url",
  "url": "https://api.example.com/sales",
  "method": "GET"
}
```

### Cube.js Data

Query the semantic layer (requires Cube.js):

```json
{
  "id": "cube-data",
  "type": "cube",
  "query": {
    "measures": ["Orders.totalRevenue"],
    "dimensions": ["Orders.createdAt"],
    "timeDimensions": [{
      "dimension": "Orders.createdAt",
      "granularity": "month"
    }]
  }
}
```

See [Cube.js Integration](cube-integration.md) for details.

## Saving Dashboards

### Export to JSON

Click the settings menu and select "Export JSON" to download the dashboard as a JSON file.

### Save to Server

If connected to a Dashforge server, click "Save" to persist the dashboard to the database.

### Import from JSON

Click "Import JSON" to load a dashboard from a file.

## Undo/Redo

The builder maintains a history of changes:

- Up to 50 states are stored
- Use Ctrl+Z / Ctrl+Shift+Z or the toolbar buttons
- History is reset when loading a new dashboard

## Technology Stack

The builder is built with:

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **Zustand** - State management with undo/redo
- **React Query** - Data fetching
- **react-grid-layout** - Drag-and-drop grid
- **ECharts** - Chart rendering via echarts-for-react
- **Cube.js Client** - Semantic queries
- **Lucide React** - Icons
