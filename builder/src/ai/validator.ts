/**
 * Validator for AI-generated dashboard and widget configurations.
 * Ensures generated JSON is valid before applying to the dashboard.
 */

import type { Widget } from '../types/dashboard'

export interface ValidationResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

/**
 * Validate a complete dashboard configuration
 */
export function validateDashboard(data: unknown): ValidationResult {
  const errors: string[] = []
  const warnings: string[] = []

  if (!data || typeof data !== 'object') {
    return { valid: false, errors: ['Dashboard must be an object'], warnings: [] }
  }

  const dashboard = data as Record<string, unknown>

  // Required fields
  if (!dashboard.title || typeof dashboard.title !== 'string') {
    errors.push('Dashboard must have a title string')
  }

  if (!Array.isArray(dashboard.widgets)) {
    errors.push('Dashboard must have a widgets array')
  } else {
    // Validate each widget
    dashboard.widgets.forEach((widget: unknown, index: number) => {
      const widgetResult = validateWidget(widget)
      if (!widgetResult.valid) {
        errors.push(...widgetResult.errors.map(e => `Widget ${index}: ${e}`))
      }
      warnings.push(...widgetResult.warnings.map(w => `Widget ${index}: ${w}`))
    })

    // Check for overlapping widgets
    const widgets = dashboard.widgets as Widget[]
    const overlaps = findOverlappingWidgets(widgets)
    if (overlaps.length > 0) {
      warnings.push(`Overlapping widgets detected: ${overlaps.join(', ')}`)
    }
  }

  // Optional but validated fields
  if (dashboard.layout) {
    if (typeof dashboard.layout !== 'object') {
      errors.push('Layout must be an object')
    } else {
      const layout = dashboard.layout as Record<string, unknown>
      if (layout.columns && (typeof layout.columns !== 'number' || layout.columns < 1 || layout.columns > 24)) {
        errors.push('Layout columns must be between 1 and 24')
      }
      if (layout.rowHeight && (typeof layout.rowHeight !== 'number' || layout.rowHeight < 20)) {
        errors.push('Layout rowHeight must be at least 20')
      }
    }
  }

  if (dashboard.dataSources && !Array.isArray(dashboard.dataSources)) {
    errors.push('dataSources must be an array')
  }

  if (dashboard.variables && !Array.isArray(dashboard.variables)) {
    errors.push('variables must be an array')
  }

  return { valid: errors.length === 0, errors, warnings }
}

/**
 * Validate a single widget configuration
 */
export function validateWidget(data: unknown): ValidationResult {
  const errors: string[] = []
  const warnings: string[] = []

  if (!data || typeof data !== 'object') {
    return { valid: false, errors: ['Widget must be an object'], warnings: [] }
  }

  const widget = data as Record<string, unknown>

  // Required fields
  const validTypes = ['chart', 'metric', 'table', 'text', 'image']
  if (!widget.type || !validTypes.includes(widget.type as string)) {
    errors.push(`Widget type must be one of: ${validTypes.join(', ')}`)
  }

  // Position validation
  if (!widget.position || typeof widget.position !== 'object') {
    errors.push('Widget must have a position object')
  } else {
    const positionResult = validatePosition(widget.position as Record<string, unknown>)
    errors.push(...positionResult.errors)
    warnings.push(...positionResult.warnings)
  }

  // Type-specific validation
  if (widget.type === 'chart' && widget.config) {
    const chartResult = validateChartConfig(widget.config)
    errors.push(...chartResult.errors)
    warnings.push(...chartResult.warnings)
  }

  // Warnings for missing recommended fields
  if (!widget.title) {
    warnings.push('Widget has no title')
  }

  if (!widget.id) {
    warnings.push('Widget has no id (will be auto-generated)')
  }

  return { valid: errors.length === 0, errors, warnings }
}

/**
 * Validate widget position
 */
export function validatePosition(position: Record<string, unknown>): ValidationResult {
  const errors: string[] = []
  const warnings: string[] = []

  const requiredFields = ['x', 'y', 'w', 'h']
  for (const field of requiredFields) {
    if (typeof position[field] !== 'number') {
      errors.push(`Position must have numeric ${field}`)
    }
  }

  if (typeof position.x === 'number' && position.x < 0) {
    errors.push('Position x cannot be negative')
  }

  if (typeof position.y === 'number' && position.y < 0) {
    errors.push('Position y cannot be negative')
  }

  if (typeof position.w === 'number') {
    if (position.w < 1) {
      errors.push('Position width must be at least 1')
    }
    if (position.w > 12) {
      warnings.push('Position width exceeds 12 columns')
    }
  }

  if (typeof position.h === 'number' && position.h < 1) {
    errors.push('Position height must be at least 1')
  }

  return { valid: errors.length === 0, errors, warnings }
}

/**
 * Validate chart configuration
 */
export function validateChartConfig(config: unknown): ValidationResult {
  const errors: string[] = []
  const warnings: string[] = []

  if (!config || typeof config !== 'object') {
    return { valid: false, errors: ['Chart config must be an object'], warnings: [] }
  }

  const chartConfig = config as Record<string, unknown>

  // Geometry validation
  const validGeometries = ['line', 'bar', 'pie', 'scatter', 'area', 'radar', 'funnel', 'gauge']
  if (!chartConfig.geometry || !validGeometries.includes(chartConfig.geometry as string)) {
    errors.push(`Chart geometry must be one of: ${validGeometries.join(', ')}`)
  }

  // Encodings validation
  const geometry = chartConfig.geometry as string
  const encodings = chartConfig.encodings as Record<string, unknown> | undefined

  if (encodings) {
    const isPieType = geometry === 'pie' || geometry === 'funnel'
    if (isPieType) {
      if (!encodings.value && !encodings.category) {
        warnings.push('Pie/funnel chart should have value and category encodings')
      }
    } else if (geometry !== 'gauge') {
      if (!encodings.x && !encodings.y) {
        warnings.push('Cartesian chart should have x and y encodings')
      }
    }
  } else {
    warnings.push('Chart has no encodings defined')
  }

  return { valid: errors.length === 0, errors, warnings }
}

/**
 * Find overlapping widgets in a layout
 */
function findOverlappingWidgets(widgets: Widget[]): string[] {
  const overlaps: string[] = []

  for (let i = 0; i < widgets.length; i++) {
    for (let j = i + 1; j < widgets.length; j++) {
      const a = widgets[i].position
      const b = widgets[j].position

      // Check if rectangles overlap
      const xOverlap = a.x < b.x + b.w && a.x + a.w > b.x
      const yOverlap = a.y < b.y + b.h && a.y + a.h > b.y

      if (xOverlap && yOverlap) {
        overlaps.push(`${widgets[i].id || i} and ${widgets[j].id || j}`)
      }
    }
  }

  return overlaps
}

/**
 * Auto-fix common issues in generated configurations
 */
export function autoFixWidget(widget: Partial<Widget>): Widget {
  const fixed: Widget = {
    id: widget.id || crypto.randomUUID(),
    type: widget.type || 'chart',
    title: widget.title || 'Untitled Widget',
    position: {
      x: Math.max(0, widget.position?.x || 0),
      y: Math.max(0, widget.position?.y || 0),
      w: Math.max(1, Math.min(12, widget.position?.w || 4)),
      h: Math.max(1, widget.position?.h || 3),
      minW: 1,
      minH: 1
    },
    config: widget.config || {}
  }

  if (widget.datasourceId) {
    fixed.datasourceId = widget.datasourceId
  }

  if (widget.description) {
    fixed.description = widget.description
  }

  return fixed
}

/**
 * Parse and validate AI-generated JSON response
 */
export function parseAIResponse<T>(
  response: string,
  validator: (data: unknown) => ValidationResult
): { data: T | null; result: ValidationResult } {
  // Try to extract JSON from response (handle markdown code blocks)
  let jsonString = response.trim()

  // Remove markdown code block if present
  const jsonMatch = jsonString.match(/```(?:json)?\s*([\s\S]*?)```/)
  if (jsonMatch) {
    jsonString = jsonMatch[1].trim()
  }

  // Try to parse JSON
  let data: unknown
  try {
    data = JSON.parse(jsonString)
  } catch (e) {
    return {
      data: null,
      result: {
        valid: false,
        errors: [`Invalid JSON: ${(e as Error).message}`],
        warnings: []
      }
    }
  }

  // Validate the parsed data
  const result = validator(data)

  return {
    data: result.valid ? (data as T) : null,
    result
  }
}
