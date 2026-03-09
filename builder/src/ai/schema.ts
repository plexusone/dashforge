/**
 * JSON Schema definitions for AI-generated dashboard configurations.
 * These schemas provide structure for LLMs to generate valid DashboardIR and ChartIR.
 */

export const DashboardSchema = {
  $schema: 'https://json-schema.org/draft/2020-12/schema',
  title: 'Dashboard',
  description: 'A dashboard configuration in DashboardIR format',
  type: 'object',
  required: ['title', 'widgets'],
  properties: {
    title: {
      type: 'string',
      description: 'The display title of the dashboard'
    },
    description: {
      type: 'string',
      description: 'Optional description of the dashboard purpose'
    },
    layout: {
      type: 'object',
      properties: {
        type: { enum: ['grid', 'flex', 'free'], default: 'grid' },
        columns: { type: 'integer', default: 12, minimum: 1, maximum: 24 },
        rowHeight: { type: 'integer', default: 80, minimum: 20 },
        gap: { type: 'integer', default: 8 },
        padding: { type: 'integer', default: 16 }
      }
    },
    widgets: {
      type: 'array',
      items: { $ref: '#/$defs/Widget' }
    },
    dataSources: {
      type: 'array',
      items: { $ref: '#/$defs/DataSource' }
    },
    variables: {
      type: 'array',
      items: { $ref: '#/$defs/Variable' }
    }
  },
  $defs: {
    Widget: {
      type: 'object',
      required: ['type', 'position'],
      properties: {
        id: { type: 'string' },
        type: { enum: ['chart', 'metric', 'table', 'text', 'image'] },
        title: { type: 'string' },
        position: { $ref: '#/$defs/Position' },
        datasourceId: { type: 'string' },
        config: { type: 'object' }
      }
    },
    Position: {
      type: 'object',
      required: ['x', 'y', 'w', 'h'],
      properties: {
        x: { type: 'integer', minimum: 0 },
        y: { type: 'integer', minimum: 0 },
        w: { type: 'integer', minimum: 1, maximum: 12 },
        h: { type: 'integer', minimum: 1 }
      }
    },
    DataSource: {
      type: 'object',
      required: ['id', 'type'],
      properties: {
        id: { type: 'string' },
        type: { enum: ['url', 'inline', 'postgres', 'mysql', 'derived', 'cube'] },
        url: { type: 'string' },
        query: { type: 'string' },
        data: { type: 'array' }
      }
    },
    Variable: {
      type: 'object',
      required: ['id', 'name', 'type'],
      properties: {
        id: { type: 'string' },
        name: { type: 'string' },
        label: { type: 'string' },
        type: { enum: ['select', 'text', 'date', 'daterange'] },
        defaultValue: { type: 'string' }
      }
    }
  }
}

export const ChartConfigSchema = {
  $schema: 'https://json-schema.org/draft/2020-12/schema',
  title: 'ChartConfig',
  description: 'Chart configuration using echartify ChartIR format',
  type: 'object',
  required: ['geometry'],
  properties: {
    geometry: {
      enum: ['line', 'bar', 'pie', 'scatter', 'area', 'radar', 'funnel', 'gauge'],
      description: 'The type of chart to render'
    },
    encodings: {
      type: 'object',
      description: 'Data field mappings',
      properties: {
        x: { type: 'string', description: 'Field for X axis (cartesian charts)' },
        y: { type: 'string', description: 'Field for Y axis (cartesian charts)' },
        color: { type: 'string', description: 'Field for color encoding' },
        size: { type: 'string', description: 'Field for size encoding' },
        value: { type: 'string', description: 'Field for value (pie/funnel)' },
        category: { type: 'string', description: 'Field for category (pie/funnel)' }
      }
    },
    style: {
      type: 'object',
      properties: {
        colors: { type: 'array', items: { type: 'string' } },
        showLegend: { type: 'boolean', default: true },
        legendPosition: { enum: ['top', 'bottom', 'left', 'right'] },
        showLabels: { type: 'boolean' },
        smooth: { type: 'boolean', description: 'Smooth lines (line/area charts)' },
        stack: { type: 'boolean', description: 'Stack series (bar/area charts)' },
        horizontal: { type: 'boolean', description: 'Horizontal bars' }
      }
    }
  }
}

export const WidgetGenerationSchema = {
  $schema: 'https://json-schema.org/draft/2020-12/schema',
  title: 'WidgetGeneration',
  description: 'Schema for AI-generated widget',
  type: 'object',
  required: ['type', 'title'],
  properties: {
    type: {
      enum: ['chart', 'metric', 'table', 'text'],
      description: 'Widget type to create'
    },
    title: {
      type: 'string',
      description: 'Widget title'
    },
    description: {
      type: 'string',
      description: 'Brief description of what the widget shows'
    },
    suggestedPosition: {
      type: 'object',
      properties: {
        w: { type: 'integer', minimum: 1, maximum: 12, description: 'Width in grid columns' },
        h: { type: 'integer', minimum: 1, maximum: 8, description: 'Height in grid rows' }
      }
    },
    chartConfig: {
      $ref: '#/$defs/ChartConfig',
      description: 'Chart configuration (for chart widgets)'
    },
    metricConfig: {
      type: 'object',
      properties: {
        valueField: { type: 'string' },
        format: { enum: ['number', 'currency', 'percent', 'compact'] },
        showComparison: { type: 'boolean' },
        showSparkline: { type: 'boolean' }
      }
    },
    tableConfig: {
      type: 'object',
      properties: {
        columns: {
          type: 'array',
          items: {
            type: 'object',
            properties: {
              field: { type: 'string' },
              header: { type: 'string' }
            }
          }
        },
        sortable: { type: 'boolean' },
        pagination: { type: 'boolean' }
      }
    },
    textConfig: {
      type: 'object',
      properties: {
        content: { type: 'string' },
        format: { enum: ['plain', 'markdown'] }
      }
    },
    suggestedQuery: {
      type: 'object',
      description: 'Suggested Cube.js query for this widget',
      properties: {
        measures: { type: 'array', items: { type: 'string' } },
        dimensions: { type: 'array', items: { type: 'string' } },
        filters: { type: 'array' }
      }
    }
  },
  $defs: {
    ChartConfig: ChartConfigSchema
  }
}

/**
 * Get schema as a compact string for embedding in prompts
 */
export function getSchemaForPrompt(schemaName: 'dashboard' | 'chart' | 'widget'): string {
  const schema = {
    dashboard: DashboardSchema,
    chart: ChartConfigSchema,
    widget: WidgetGenerationSchema
  }[schemaName]

  return JSON.stringify(schema, null, 2)
}
