/**
 * System prompts for AI dashboard generation.
 * These prompts provide context for LLMs to generate valid dashboard configurations.
 */

import { getSchemaForPrompt } from './schema'
import type { CubeSchema } from '../api/cube'

/**
 * Base system prompt for dashboard generation
 */
export const DASHBOARD_SYSTEM_PROMPT = `You are a dashboard design assistant. You help users create data dashboards by generating JSON configurations.

You output JSON that follows the DashboardIR specification. Key rules:
- Use a 12-column grid layout
- Position widgets using x, y, w, h coordinates (x and y start at 0)
- Common widget sizes: metrics (2x2), charts (4x3 or 6x3), tables (6x4)
- Align widgets to avoid overlap
- Use descriptive titles for widgets
- Connect widgets to appropriate data sources

Chart types available: line, bar, pie, scatter, area
Widget types available: chart, metric, table, text, image

When creating charts:
- For time series: use line or area charts with date on x-axis
- For comparisons: use bar charts
- For distributions: use pie charts
- For correlations: use scatter plots

Always respond with valid JSON only. No explanations or markdown.`

/**
 * Generate a prompt for creating a full dashboard
 */
export function generateDashboardPrompt(
  userRequest: string,
  schema?: CubeSchema
): string {
  let prompt = `${DASHBOARD_SYSTEM_PROMPT}

Dashboard Schema:
${getSchemaForPrompt('dashboard')}

`

  // Add available data context if schema is provided
  if (schema && schema.cubes.length > 0) {
    prompt += `Available data cubes:
${schema.cubes.map(cube => `- ${cube.name}: ${cube.title}
  Measures: ${cube.measures.map(m => m.name).join(', ')}
  Dimensions: ${cube.dimensions.map(d => d.name).join(', ')}`).join('\n')}

`
  }

  prompt += `User request: "${userRequest}"

Generate a complete dashboard JSON configuration. Include appropriate widgets, layout, and data source references.`

  return prompt
}

/**
 * Generate a prompt for adding a single widget
 */
export function generateWidgetPrompt(
  userRequest: string,
  existingWidgets: { type: string; title: string; position: { x: number; y: number; w: number; h: number } }[],
  schema?: CubeSchema
): string {
  let prompt = `${DASHBOARD_SYSTEM_PROMPT}

Widget Schema:
${getSchemaForPrompt('widget')}

`

  // Add existing widgets context
  if (existingWidgets.length > 0) {
    const occupiedPositions = existingWidgets.map(w =>
      `${w.title}: x=${w.position.x}, y=${w.position.y}, w=${w.position.w}, h=${w.position.h}`
    )
    prompt += `Existing widgets in dashboard:
${occupiedPositions.join('\n')}

Avoid overlapping with existing widgets.

`
  }

  // Add available data context
  if (schema && schema.cubes.length > 0) {
    prompt += `Available data:
${schema.cubes.map(cube =>
  `${cube.name}: measures=[${cube.measures.slice(0, 5).map(m => m.shortTitle).join(', ')}], dimensions=[${cube.dimensions.slice(0, 5).map(d => d.shortTitle).join(', ')}]`
).join('\n')}

`
  }

  prompt += `User request: "${userRequest}"

Generate a single widget JSON configuration with suggested position that doesn't overlap existing widgets.`

  return prompt
}

/**
 * Generate a prompt for modifying a widget
 */
export function generateWidgetModificationPrompt(
  userRequest: string,
  currentWidget: unknown
): string {
  return `${DASHBOARD_SYSTEM_PROMPT}

Current widget configuration:
${JSON.stringify(currentWidget, null, 2)}

User modification request: "${userRequest}"

Generate the modified widget JSON. Only change what the user requested, keep other settings the same.`
}

/**
 * Generate a prompt for natural language querying
 */
export function generateQueryPrompt(
  userRequest: string,
  schema: CubeSchema
): string {
  const schemaContext = schema.cubes.map(cube => `
Cube: ${cube.name} (${cube.title})
${cube.description ? `Description: ${cube.description}` : ''}
Measures:
${cube.measures.map(m => `  - ${m.name}: ${m.title}${m.description ? ` - ${m.description}` : ''} [${m.aggType || 'count'}]`).join('\n')}
Dimensions:
${cube.dimensions.map(d => `  - ${d.name}: ${d.title}${d.description ? ` - ${d.description}` : ''} [${d.type}]`).join('\n')}
`).join('\n---\n')

  return `You are a query assistant that translates natural language questions into Cube.js queries.

Available Schema:
${schemaContext}

Query JSON format:
{
  "measures": ["CubeName.measureName"],
  "dimensions": ["CubeName.dimensionName"],
  "filters": [{"member": "CubeName.dimension", "operator": "equals", "values": ["value"]}],
  "timeDimensions": [{"dimension": "CubeName.date", "granularity": "month", "dateRange": "last 6 months"}],
  "order": {"CubeName.measure": "desc"},
  "limit": 100
}

User question: "${userRequest}"

Generate a Cube.js query JSON to answer this question. Only output valid JSON.`
}

/**
 * Example prompts for testing
 */
export const EXAMPLE_PROMPTS = {
  dashboard: [
    'Create a sales dashboard with revenue by region and monthly trends',
    'Build an executive dashboard showing KPIs for revenue, customers, and orders',
    'Design a marketing analytics dashboard with campaign performance metrics'
  ],
  widget: [
    'Add a line chart showing revenue over the last 6 months',
    'Add a metric card showing total customers',
    'Add a pie chart breaking down sales by category',
    'Add a table listing top 10 products by revenue'
  ],
  modify: [
    'Change the bar chart to show percentages instead of absolute values',
    'Make the chart horizontal instead of vertical',
    'Add a comparison to previous period'
  ],
  query: [
    'Show me total revenue by month for this year',
    'Which products have the highest profit margin?',
    'How many new customers did we get last quarter?'
  ]
}
