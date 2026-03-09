/**
 * AI API Client for dashboard generation.
 * Communicates with the Dashforge server's /api/v1/ai endpoints.
 */

import type { Dashboard, Widget } from '../types/dashboard'
import type { CubeSchema } from './cube'

const AI_API_BASE = '/api/v1/ai'

export interface AIGenerationOptions {
  model?: string
  temperature?: number
  maxTokens?: number
}

export interface AIGenerationResult<T> {
  success: boolean
  data?: T
  text?: string
  errors?: string[]
  warnings?: string[]
  usage?: {
    promptTokens: number
    completionTokens: number
    totalTokens: number
  }
  model?: string
  provider?: string
}

interface GenerateRequest {
  prompt: string
  systemPrompt?: string
  type?: 'dashboard' | 'widget' | 'query' | 'modify'
  context?: {
    existingWidgets?: {
      id: string
      type: string
      title: string
      position: { x: number; y: number; w: number; h: number }
    }[]
    schema?: CubeSchema
    currentWidget?: unknown
  }
  model?: string
  temperature?: number
  maxTokens?: number
}

interface GenerateResponse {
  success: boolean
  data?: unknown
  text?: string
  error?: string
  warnings?: string[]
  usage?: {
    promptTokens: number
    completionTokens: number
    totalTokens: number
  }
  model?: string
  provider?: string
}

/**
 * Check if the AI service is available
 */
export async function checkAIStatus(): Promise<{ enabled: boolean; providers?: string[] }> {
  try {
    const response = await fetch(`${AI_API_BASE}/status`)
    if (!response.ok) {
      return { enabled: false }
    }
    return response.json()
  } catch {
    return { enabled: false }
  }
}

/**
 * Get available AI models
 */
export async function getAvailableModels(): Promise<{
  models: Record<string, string[]>
  default: string
}> {
  const response = await fetch(`${AI_API_BASE}/models`)
  if (!response.ok) {
    throw new Error('Failed to fetch models')
  }
  return response.json()
}

/**
 * Generate a complete dashboard from natural language description
 */
export async function generateDashboard(
  prompt: string,
  schema?: CubeSchema,
  options?: AIGenerationOptions
): Promise<AIGenerationResult<Dashboard>> {
  const request: GenerateRequest = {
    prompt,
    type: 'dashboard',
    context: schema ? { schema } : undefined,
    ...options
  }

  try {
    const response = await fetch(`${AI_API_BASE}/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    })

    const result: GenerateResponse = await response.json()

    if (!result.success || !result.data) {
      return {
        success: false,
        errors: [result.error || 'Failed to generate dashboard'],
        warnings: result.warnings
      }
    }

    // Parse and validate the dashboard
    const dashboard = result.data as Dashboard

    // Ensure required fields
    if (!dashboard.title) {
      dashboard.title = 'Generated Dashboard'
    }
    if (!dashboard.id) {
      dashboard.id = crypto.randomUUID()
    }
    if (!dashboard.layout) {
      dashboard.layout = { type: 'grid', columns: 12, rowHeight: 80 }
    }
    if (!dashboard.widgets) {
      dashboard.widgets = []
    }
    if (!dashboard.dataSources) {
      dashboard.dataSources = []
    }

    // Ensure widget IDs
    dashboard.widgets = dashboard.widgets.map(w => ({
      ...w,
      id: w.id || crypto.randomUUID(),
      position: {
        ...w.position,
        minW: w.position.minW || 1,
        minH: w.position.minH || 1
      }
    }))

    return {
      success: true,
      data: dashboard,
      text: result.text,
      warnings: result.warnings,
      usage: result.usage,
      model: result.model,
      provider: result.provider
    }

  } catch (error) {
    return {
      success: false,
      errors: [error instanceof Error ? error.message : 'Unknown error']
    }
  }
}

/**
 * Generate a single widget from natural language description
 */
export async function generateWidget(
  prompt: string,
  existingWidgets: Widget[],
  schema?: CubeSchema,
  options?: AIGenerationOptions
): Promise<AIGenerationResult<Widget>> {
  const widgetSummaries = existingWidgets.map(w => ({
    id: w.id,
    type: w.type,
    title: w.title || 'Untitled',
    position: w.position
  }))

  const request: GenerateRequest = {
    prompt,
    type: 'widget',
    context: {
      existingWidgets: widgetSummaries,
      schema
    },
    ...options
  }

  try {
    const response = await fetch(`${AI_API_BASE}/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    })

    const result: GenerateResponse = await response.json()

    if (!result.success || !result.data) {
      return {
        success: false,
        errors: [result.error || 'Failed to generate widget'],
        warnings: result.warnings
      }
    }

    // Parse and fix the widget
    const widget = result.data as Partial<Widget>

    const fixedWidget: Widget = {
      id: widget.id || crypto.randomUUID(),
      type: widget.type || 'chart',
      title: widget.title || 'Generated Widget',
      position: {
        x: Math.max(0, widget.position?.x || 0),
        y: Math.max(0, widget.position?.y || findNextY(existingWidgets)),
        w: Math.max(1, Math.min(12, widget.position?.w || 4)),
        h: Math.max(1, widget.position?.h || 3),
        minW: 1,
        minH: 1
      },
      config: widget.config || {},
      datasourceId: widget.datasourceId,
      description: widget.description
    }

    return {
      success: true,
      data: fixedWidget,
      text: result.text,
      warnings: result.warnings,
      usage: result.usage,
      model: result.model,
      provider: result.provider
    }

  } catch (error) {
    return {
      success: false,
      errors: [error instanceof Error ? error.message : 'Unknown error']
    }
  }
}

/**
 * Modify an existing widget based on natural language instructions
 */
export async function modifyWidget(
  prompt: string,
  widget: Widget,
  options?: AIGenerationOptions
): Promise<AIGenerationResult<Widget>> {
  const request: GenerateRequest = {
    prompt,
    type: 'modify',
    context: {
      currentWidget: widget
    },
    ...options
  }

  try {
    const response = await fetch(`${AI_API_BASE}/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    })

    const result: GenerateResponse = await response.json()

    if (!result.success || !result.data) {
      return {
        success: false,
        errors: [result.error || 'Failed to modify widget'],
        warnings: result.warnings
      }
    }

    // Preserve original ID and merge changes
    const modifiedWidget: Widget = {
      ...widget,
      ...(result.data as Partial<Widget>),
      id: widget.id // Always preserve original ID
    }

    return {
      success: true,
      data: modifiedWidget,
      text: result.text,
      warnings: result.warnings,
      usage: result.usage,
      model: result.model,
      provider: result.provider
    }

  } catch (error) {
    return {
      success: false,
      errors: [error instanceof Error ? error.message : 'Unknown error']
    }
  }
}

/**
 * Generate a Cube.js query from natural language
 */
export async function generateQuery(
  prompt: string,
  schema: CubeSchema,
  options?: AIGenerationOptions
): Promise<AIGenerationResult<Record<string, unknown>>> {
  const request: GenerateRequest = {
    prompt,
    type: 'query',
    context: { schema },
    ...options
  }

  try {
    const response = await fetch(`${AI_API_BASE}/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request)
    })

    const result: GenerateResponse = await response.json()

    if (!result.success || !result.data) {
      return {
        success: false,
        errors: [result.error || 'Failed to generate query'],
        warnings: result.warnings
      }
    }

    return {
      success: true,
      data: result.data as Record<string, unknown>,
      text: result.text,
      warnings: result.warnings,
      usage: result.usage,
      model: result.model,
      provider: result.provider
    }

  } catch (error) {
    return {
      success: false,
      errors: [error instanceof Error ? error.message : 'Unknown error']
    }
  }
}

/**
 * Stream AI generation (for real-time updates)
 */
export async function* streamGenerate(
  prompt: string,
  type: 'dashboard' | 'widget' | 'query' | 'modify',
  context?: GenerateRequest['context'],
  options?: AIGenerationOptions
): AsyncGenerator<{ content?: string; error?: string; done?: boolean }> {
  const request: GenerateRequest = {
    prompt,
    type,
    context,
    ...options
  }

  const response = await fetch(`${AI_API_BASE}/stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  })

  if (!response.ok) {
    yield { error: `HTTP ${response.status}: ${response.statusText}` }
    return
  }

  const reader = response.body?.getReader()
  if (!reader) {
    yield { error: 'No response body' }
    return
  }

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = line.slice(6)
        if (data === '[DONE]') {
          yield { done: true }
          return
        }
        try {
          const parsed = JSON.parse(data)
          if (parsed.error) {
            yield { error: parsed.error }
          } else if (parsed.content) {
            yield { content: parsed.content }
          }
        } catch {
          // Ignore parse errors for partial data
        }
      }
    }
  }
}

/**
 * Find the next available Y position
 */
function findNextY(widgets: Widget[]): number {
  if (widgets.length === 0) return 0
  return Math.max(...widgets.map(w => w.position.y + w.position.h))
}

/**
 * Mock AI generation for development (when AI backend is not available)
 */
export async function mockGenerateWidget(
  prompt: string,
  existingWidgets: Widget[]
): Promise<AIGenerationResult<Widget>> {
  const lowerPrompt = prompt.toLowerCase()

  let widgetType: 'chart' | 'metric' | 'table' | 'text' = 'chart'
  let geometry: 'line' | 'bar' | 'pie' | 'scatter' | 'area' = 'bar'

  if (lowerPrompt.includes('metric') || lowerPrompt.includes('kpi') || lowerPrompt.includes('total')) {
    widgetType = 'metric'
  } else if (lowerPrompt.includes('table') || lowerPrompt.includes('list')) {
    widgetType = 'table'
  } else if (lowerPrompt.includes('text') || lowerPrompt.includes('note')) {
    widgetType = 'text'
  } else if (lowerPrompt.includes('line') || lowerPrompt.includes('trend')) {
    geometry = 'line'
  } else if (lowerPrompt.includes('pie') || lowerPrompt.includes('distribution')) {
    geometry = 'pie'
  } else if (lowerPrompt.includes('scatter') || lowerPrompt.includes('correlation')) {
    geometry = 'scatter'
  } else if (lowerPrompt.includes('area')) {
    geometry = 'area'
  }

  const nextY = findNextY(existingWidgets)

  const widget: Widget = {
    id: crypto.randomUUID(),
    type: widgetType,
    title: `Widget from: "${prompt.slice(0, 30)}..."`,
    position: {
      x: 0,
      y: nextY,
      w: widgetType === 'metric' ? 2 : widgetType === 'table' ? 6 : 4,
      h: widgetType === 'metric' ? 2 : widgetType === 'table' ? 4 : 3,
      minW: 1,
      minH: 1
    },
    config: widgetType === 'chart'
      ? {
          geometry,
          encodings: geometry === 'pie'
            ? { value: 'revenue', category: 'category' }
            : { x: 'date', y: 'revenue' },
          style: { showLegend: true }
        }
      : widgetType === 'metric'
      ? { valueField: 'total', format: 'number' }
      : widgetType === 'table'
      ? { columns: [], sortable: true, pagination: { enabled: true } }
      : { content: 'Enter text here...', format: 'markdown' }
  }

  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 500))

  return {
    success: true,
    data: widget,
    warnings: ['Using mock AI generation - connect AI backend for full functionality']
  }
}
