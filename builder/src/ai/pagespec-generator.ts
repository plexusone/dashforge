import type { PageSpec } from '../stores/pagespec'

export interface ValidationResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

export interface PageSpecGenerationRequest {
  prompt: string
  profile?: string
  availableComponents?: string[]
  existingPage?: PageSpec
}

export interface PageSpecGenerationResult {
  page: PageSpec | null
  validation: ValidationResult
  retryCount: number
}

const VALID_LAYOUT_TYPES = [
  'responsive-grid',
  'stack',
  'split-pane',
  'tabs',
  'application-shell',
]

export function validatePageSpec(data: unknown): ValidationResult {
  const errors: string[] = []
  const warnings: string[] = []

  if (!data || typeof data !== 'object') {
    return { valid: false, errors: ['PageSpec must be a JSON object'], warnings: [] }
  }

  const page = data as Record<string, unknown>

  if (!page.apiVersion || typeof page.apiVersion !== 'string') {
    errors.push('Missing or invalid apiVersion (expected string like "ui.plexusone.dev/v1")')
  }

  if (!page.kind || page.kind !== 'Page') {
    errors.push('Missing or invalid kind (must be "Page")')
  }

  if (!page.metadata || typeof page.metadata !== 'object') {
    errors.push('Missing metadata object')
  } else {
    const meta = page.metadata as Record<string, unknown>
    if (!meta.id || typeof meta.id !== 'string') {
      errors.push('metadata.id is required (string)')
    }
    if (!meta.name || typeof meta.name !== 'string') {
      errors.push('metadata.name is required (string)')
    }
  }

  if (!page.layout || typeof page.layout !== 'object') {
    errors.push('Missing layout object')
  } else {
    const layout = page.layout as Record<string, unknown>
    if (!layout.type || typeof layout.type !== 'string') {
      errors.push('layout.type is required (string)')
    } else if (!VALID_LAYOUT_TYPES.includes(layout.type as string)) {
      errors.push(
        `Invalid layout.type "${layout.type}". Must be one of: ${VALID_LAYOUT_TYPES.join(', ')}`
      )
    }
  }

  if (!Array.isArray(page.components)) {
    errors.push('components must be an array')
  } else {
    const ids = new Set<string>()
    for (let i = 0; i < page.components.length; i++) {
      const comp = page.components[i] as Record<string, unknown>
      if (!comp || typeof comp !== 'object') {
        errors.push(`components[${i}] must be an object`)
        continue
      }
      if (!comp.id || typeof comp.id !== 'string') {
        errors.push(`components[${i}].id is required (string)`)
      } else if (ids.has(comp.id as string)) {
        errors.push(`Duplicate component id: "${comp.id}"`)
      } else {
        ids.add(comp.id as string)
      }
      if (!comp.type || typeof comp.type !== 'string') {
        errors.push(`components[${i}].type is required (string)`)
      }
    }

    if (Array.isArray(page.interactions)) {
      for (let i = 0; i < page.interactions.length; i++) {
        const inter = page.interactions[i] as Record<string, unknown>
        if (!inter || typeof inter !== 'object') {
          warnings.push(`interactions[${i}] should be an object`)
          continue
        }
        const when = inter.when as Record<string, unknown> | undefined
        if (when?.component && !ids.has(when.component as string)) {
          warnings.push(
            `interactions[${i}].when.component "${when.component}" does not match any component id`
          )
        }
      }
    }
  }

  if (page.profile && typeof page.profile !== 'string') {
    warnings.push('profile should be a string')
  }

  return { valid: errors.length === 0, errors, warnings }
}

export function buildPageSpecPrompt(request: PageSpecGenerationRequest): string {
  let prompt = `You are a UISpec page design assistant. You generate valid UISpec PageSpec JSON documents.

A PageSpec describes a composable UI page with this structure:
{
  "apiVersion": "ui.plexusone.dev/v1",
  "kind": "Page",
  "metadata": { "id": "unique-id", "name": "slug-name", "title": "Display Title" },
  "profile": "${request.profile || 'dashboard'}",
  "layout": {
    "type": "responsive-grid",
    "config": { "columns": 12, "gap": "16px" }
  },
  "components": [
    {
      "id": "unique-component-id",
      "type": "namespace.component-name",
      "position": { "row": 0, "col": 0, "colSpan": 6, "rowSpan": 3 },
      "properties": { "title": "Example" },
      "data": {
        "binding-name": {
          "source": "data-source-id",
          "operation": "operationName",
          "parameters": {}
        }
      }
    }
  ],
  "interactions": [
    {
      "when": { "component": "component-id", "event": "eventName" },
      "then": [
        { "target": "other-component-id", "action": "actionName", "params": {} }
      ]
    }
  ]
}

Layout types: responsive-grid, stack, split-pane, tabs, application-shell.
For responsive-grid: use position with row, col, colSpan, rowSpan (12-column grid).
For split-pane/application-shell: use regions and slot on components.
For stack: components render in order.

`

  if (request.availableComponents && request.availableComponents.length > 0) {
    prompt += `Available component types:
${request.availableComponents.map((c) => `- ${c}`).join('\n')}

Only use components from this list.

`
  }

  if (request.existingPage) {
    prompt += `Current page to modify:
${JSON.stringify(request.existingPage, null, 2)}

Modify this page based on the user request. Keep existing components unless asked to remove them.

`
  }

  prompt += `User request: "${request.prompt}"

Generate a complete, valid PageSpec JSON. Output only the JSON, no explanations or markdown.`

  return prompt
}

function extractJSON(response: string): string {
  let jsonString = response.trim()
  const jsonMatch = jsonString.match(/```(?:json)?\s*([\s\S]*?)```/)
  if (jsonMatch) {
    jsonString = jsonMatch[1].trim()
  }
  return jsonString
}

export async function generatePageSpec(
  request: PageSpecGenerationRequest,
  callAI: (prompt: string) => Promise<string>,
  maxRetries: number = 3
): Promise<PageSpecGenerationResult> {
  let prompt = buildPageSpecPrompt(request)
  let lastValidation: ValidationResult = { valid: false, errors: ['No response yet'], warnings: [] }
  let retryCount = 0

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    let response: string
    try {
      response = await callAI(prompt)
    } catch (err) {
      lastValidation = {
        valid: false,
        errors: [`AI call failed: ${(err as Error).message}`],
        warnings: [],
      }
      retryCount = attempt
      continue
    }

    const jsonString = extractJSON(response)

    let parsed: unknown
    try {
      parsed = JSON.parse(jsonString)
    } catch (err) {
      lastValidation = {
        valid: false,
        errors: [`Invalid JSON: ${(err as Error).message}`],
        warnings: [],
      }
      retryCount = attempt

      if (attempt < maxRetries) {
        prompt += `\n\nYour previous response was invalid JSON. Error: ${(err as Error).message}\nPlease output ONLY valid JSON, no markdown or text.`
      }
      continue
    }

    lastValidation = validatePageSpec(parsed)
    retryCount = attempt

    if (lastValidation.valid) {
      return {
        page: parsed as PageSpec,
        validation: lastValidation,
        retryCount,
      }
    }

    if (attempt < maxRetries) {
      prompt += `\n\nYour previous response had validation errors:\n${lastValidation.errors.map((e) => `- ${e}`).join('\n')}\nPlease fix these issues and output corrected JSON only.`
    }
  }

  return {
    page: null,
    validation: lastValidation,
    retryCount,
  }
}
