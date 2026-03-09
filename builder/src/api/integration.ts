/**
 * Integration Platform API Client
 * Handles integration, alert, and marketplace operations.
 */

import type {
  Integration,
  Alert,
  IntegrationDefinition,
  CreateIntegrationRequest,
  UpdateIntegrationRequest,
  CreateAlertRequest,
  UpdateAlertRequest,
  InstallIntegrationRequest,
  TestConnectionResult,
  ListIntegrationsResponse,
  ListAlertsResponse,
  ListAlertEventsResponse,
  ListMarketplaceResponse,
  ListChannelsResponse,
} from '../types/integration'

const API_BASE = '/api/v1'

interface ApiError {
  error: string
}

class IntegrationApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public apiError?: ApiError
  ) {
    super(message)
    this.name = 'IntegrationApiError'
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let apiError: ApiError | undefined
    try {
      apiError = await response.json()
    } catch {
      // Response body is not JSON
    }
    throw new IntegrationApiError(
      apiError?.error || `HTTP ${response.status}: ${response.statusText}`,
      response.status,
      apiError
    )
  }
  return response.json()
}

// Integration API

export async function listIntegrations(params?: {
  channelType?: string
  status?: string
  source?: string
}): Promise<ListIntegrationsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.channelType) searchParams.set('channelType', params.channelType)
  if (params?.status) searchParams.set('status', params.status)
  if (params?.source) searchParams.set('source', params.source)

  const url = `${API_BASE}/integrations${searchParams.toString() ? `?${searchParams}` : ''}`
  const response = await fetch(url)
  return handleResponse(response)
}

export async function getIntegration(id: number): Promise<Integration> {
  const response = await fetch(`${API_BASE}/integrations/${id}`)
  return handleResponse(response)
}

export async function createIntegration(request: CreateIntegrationRequest): Promise<Integration> {
  const response = await fetch(`${API_BASE}/integrations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  })
  return handleResponse(response)
}

export async function updateIntegration(id: number, request: UpdateIntegrationRequest): Promise<Integration> {
  const response = await fetch(`${API_BASE}/integrations/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  })
  return handleResponse(response)
}

export async function deleteIntegration(id: number): Promise<void> {
  const response = await fetch(`${API_BASE}/integrations/${id}`, {
    method: 'DELETE'
  })
  if (!response.ok) {
    throw new IntegrationApiError(
      `Failed to delete integration: ${response.statusText}`,
      response.status
    )
  }
}

export async function testIntegration(id: number): Promise<TestConnectionResult> {
  const response = await fetch(`${API_BASE}/integrations/${id}/test`, {
    method: 'POST'
  })
  return handleResponse(response)
}

export async function listChannels(): Promise<ListChannelsResponse> {
  const response = await fetch(`${API_BASE}/integrations/channels`)
  return handleResponse(response)
}

// Alert API

export async function listAlerts(params?: {
  triggerType?: string
  enabled?: boolean
}): Promise<ListAlertsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.triggerType) searchParams.set('triggerType', params.triggerType)
  if (params?.enabled !== undefined) searchParams.set('enabled', String(params.enabled))

  const url = `${API_BASE}/alerts${searchParams.toString() ? `?${searchParams}` : ''}`
  const response = await fetch(url)
  return handleResponse(response)
}

export async function getAlert(id: number): Promise<Alert> {
  const response = await fetch(`${API_BASE}/alerts/${id}`)
  return handleResponse(response)
}

export async function createAlert(request: CreateAlertRequest): Promise<Alert> {
  const response = await fetch(`${API_BASE}/alerts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  })
  return handleResponse(response)
}

export async function updateAlert(id: number, request: UpdateAlertRequest): Promise<Alert> {
  const response = await fetch(`${API_BASE}/alerts/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  })
  return handleResponse(response)
}

export async function deleteAlert(id: number): Promise<void> {
  const response = await fetch(`${API_BASE}/alerts/${id}`, {
    method: 'DELETE'
  })
  if (!response.ok) {
    throw new IntegrationApiError(
      `Failed to delete alert: ${response.statusText}`,
      response.status
    )
  }
}

export async function enableAlert(id: number): Promise<Alert> {
  const response = await fetch(`${API_BASE}/alerts/${id}/enable`, {
    method: 'POST'
  })
  return handleResponse(response)
}

export async function disableAlert(id: number): Promise<Alert> {
  const response = await fetch(`${API_BASE}/alerts/${id}/disable`, {
    method: 'POST'
  })
  return handleResponse(response)
}

export async function getAlertEvents(id: number, limit?: number): Promise<ListAlertEventsResponse> {
  const searchParams = new URLSearchParams()
  if (limit) searchParams.set('limit', String(limit))

  const url = `${API_BASE}/alerts/${id}/events${searchParams.toString() ? `?${searchParams}` : ''}`
  const response = await fetch(url)
  return handleResponse(response)
}

export async function getDashboardAlerts(dashboardId: number): Promise<ListAlertsResponse> {
  const response = await fetch(`${API_BASE}/dashboards/${dashboardId}/alerts`)
  return handleResponse(response)
}

// Marketplace API

export async function listMarketplaceIntegrations(category?: string): Promise<ListMarketplaceResponse> {
  const searchParams = new URLSearchParams()
  if (category) searchParams.set('category', category)

  const url = `${API_BASE}/marketplace/integrations${searchParams.toString() ? `?${searchParams}` : ''}`
  const response = await fetch(url)
  return handleResponse(response)
}

export async function getMarketplaceIntegration(slug: string): Promise<IntegrationDefinition> {
  const response = await fetch(`${API_BASE}/marketplace/integrations/${slug}`)
  return handleResponse(response)
}

export async function installMarketplaceIntegration(
  marketplaceSlug: string,
  request: InstallIntegrationRequest
): Promise<Integration> {
  const response = await fetch(`${API_BASE}/marketplace/integrations/${marketplaceSlug}/install`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request)
  })
  return handleResponse(response)
}

// Export error class
export { IntegrationApiError }
