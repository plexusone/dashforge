/**
 * Dashforge REST API Client
 * Handles dashboard CRUD operations and data source management.
 */

import type { AnalyticsCatalog, Dashboard, DataSource } from '../types/dashboard'

const API_BASE = '/api/v1'

interface ApiError {
  message: string
  error?: string
  code?: string
  details?: unknown
}

class DashforgeApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public apiError?: ApiError,
  ) {
    super(message)
    this.name = 'DashforgeApiError'
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
    throw new DashforgeApiError(
      apiError?.message || apiError?.error || `HTTP ${response.status}: ${response.statusText}`,
      response.status,
      apiError,
    )
  }
  return response.json()
}

// Dashboard API

export interface ListDashboardsParams {
  page?: number
  limit?: number
  search?: string
}

export interface ListDashboardsResponse {
  dashboards: Dashboard[]
  total: number
  page: number
  limit: number
}

export async function listDashboards(
  params?: ListDashboardsParams,
): Promise<ListDashboardsResponse> {
  const searchParams = new URLSearchParams()
  if (params?.page) searchParams.set('page', String(params.page))
  if (params?.limit) searchParams.set('limit', String(params.limit))
  if (params?.search) searchParams.set('search', params.search)

  const url = `${API_BASE}/dashboards${searchParams.toString() ? `?${searchParams}` : ''}`
  const response = await fetch(url)
  return handleResponse(response)
}

export async function getDashboard(id: string): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/dashboards/${id}`)
  return handleResponse(response)
}

export async function createDashboard(
  dashboard: Omit<Dashboard, 'id' | 'createdAt' | 'updatedAt'>,
): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/dashboards`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(dashboard),
  })
  return handleResponse(response)
}

export async function updateDashboard(
  id: string,
  dashboard: Partial<Dashboard>,
): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/dashboards/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(dashboard),
  })
  return handleResponse(response)
}

export async function deleteDashboard(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/dashboards/${id}`, {
    method: 'DELETE',
  })
  if (!response.ok) {
    throw new DashforgeApiError(
      `Failed to delete dashboard: ${response.statusText}`,
      response.status,
    )
  }
}

export async function duplicateDashboard(id: string): Promise<Dashboard> {
  const response = await fetch(`${API_BASE}/dashboards/${id}/duplicate`, {
    method: 'POST',
  })
  return handleResponse(response)
}

// DataSource API

export interface ListDataSourcesResponse {
  dataSources: DataSource[]
  total: number
}

export async function listDataSources(): Promise<ListDataSourcesResponse> {
  const response = await fetch(`${API_BASE}/datasources`)
  return handleResponse(response)
}

export async function getDataSource(id: string): Promise<DataSource> {
  const response = await fetch(`${API_BASE}/datasources/${id}`)
  return handleResponse(response)
}

export async function createDataSource(dataSource: Omit<DataSource, 'id'>): Promise<DataSource> {
  const response = await fetch(`${API_BASE}/datasources`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(dataSource),
  })
  return handleResponse(response)
}

export async function updateDataSource(
  id: string,
  dataSource: Partial<DataSource>,
): Promise<DataSource> {
  const response = await fetch(`${API_BASE}/datasources/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(dataSource),
  })
  return handleResponse(response)
}

export async function deleteDataSource(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/datasources/${id}`, {
    method: 'DELETE',
  })
  if (!response.ok) {
    throw new DashforgeApiError(
      `Failed to delete data source: ${response.statusText}`,
      response.status,
    )
  }
}

export interface QueryParams {
  dataSourceId: string
  query?: string
  variables?: Record<string, unknown>
}

export interface QueryResult {
  data: unknown[]
  columns: string[]
  rowCount: number
  executionTime: number
}

export async function executeQuery(params: QueryParams): Promise<QueryResult> {
  const response = await fetch(`${API_BASE}/datasources/${params.dataSourceId}/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query: params.query,
      variables: params.variables,
    }),
  })
  return handleResponse(response)
}

export interface SchemaInfo {
  tables: {
    name: string
    columns: {
      name: string
      type: string
      nullable: boolean
    }[]
  }[]
}

export async function getDataSourceSchema(id: string): Promise<SchemaInfo> {
  const response = await fetch(`${API_BASE}/datasources/${id}/schema`)
  return handleResponse(response)
}

export async function getAnalyticsCatalog(): Promise<AnalyticsCatalog> {
  const response = await fetch(`${API_BASE}/analytics/catalog`)
  return handleResponse(response)
}

export interface AnalyticsQueryParams {
  sourceId: string
  query: string
  dialect?: 'grokifyql' | 'sql' | string
  params?: Record<string, unknown>
  limit?: number
}

export interface AnalyticsQueryResult {
  columns: { name: string; type?: string }[]
  rows: Record<string, unknown>[]
  rowCount: number
  executionTimeMs?: number
}

export async function executeAnalyticsQuery(
  params: AnalyticsQueryParams,
): Promise<AnalyticsQueryResult> {
  const response = await fetch(`${API_BASE}/analytics/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
  return handleResponse(response)
}

export interface AnalyticsSourceConfig {
  id: string
  name: string
  connector: string
  dsnRef: string
  enabled: boolean
  createdAt?: string
  updatedAt?: string
}

export interface AnalyticsSourceStatus extends AnalyticsSourceConfig {
  status: 'connected' | 'error' | 'disabled'
  error?: string
}

export async function listAnalyticsSources(): Promise<AnalyticsSourceStatus[]> {
  const response = await fetch(`${API_BASE}/analytics/sources`)
  const data = await handleResponse<{ sources: AnalyticsSourceStatus[] }>(response)
  return data.sources ?? []
}

export async function createAnalyticsSource(
  config: AnalyticsSourceConfig,
): Promise<AnalyticsSourceConfig> {
  const response = await fetch(`${API_BASE}/analytics/sources`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  return handleResponse(response)
}

export async function updateAnalyticsSource(
  config: AnalyticsSourceConfig,
): Promise<AnalyticsSourceConfig> {
  const response = await fetch(`${API_BASE}/analytics/sources/${encodeURIComponent(config.id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  return handleResponse(response)
}

export async function deleteAnalyticsSource(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/analytics/sources/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
  if (!response.ok && response.status !== 204) {
    await handleResponse(response)
  }
}

export interface AnalyticsSourceTestResult {
  ok: boolean
  error?: string
}

export async function testAnalyticsSource(
  config: AnalyticsSourceConfig,
): Promise<AnalyticsSourceTestResult> {
  const response = await fetch(`${API_BASE}/analytics/sources/test`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  return handleResponse(response)
}

export async function listAnalyticsConnectors(): Promise<string[]> {
  const response = await fetch(`${API_BASE}/analytics/connectors`)
  const data = await handleResponse<{ connectors: string[] }>(response)
  return data.connectors ?? []
}

export interface SavedQuestion {
  id: string
  name: string
  description?: string
  sourceId: string
  datasetId: string
  dialect?: string
  query: string
  compiledQuery?: {
    version: string
    ast?: unknown
    fingerprint: string
    datasets?: string[]
    fields?: string[]
    readOnly: boolean
    limit?: number
  }
  visualization?: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface ListSavedQuestionsResponse {
  questions: SavedQuestion[]
  total: number
}

export async function listSavedQuestions(): Promise<ListSavedQuestionsResponse> {
  const response = await fetch(`${API_BASE}/questions`)
  return handleResponse(response)
}

export async function getSavedQuestion(id: string): Promise<SavedQuestion> {
  const response = await fetch(`${API_BASE}/questions/${id}`)
  return handleResponse(response)
}

export async function saveSavedQuestion(
  question: Omit<SavedQuestion, 'id' | 'createdAt' | 'updatedAt'> &
    Partial<Pick<SavedQuestion, 'id'>>,
): Promise<SavedQuestion> {
  const response = await fetch(`${API_BASE}/questions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(question),
  })
  return handleResponse(response)
}

export async function updateSavedQuestion(
  id: string,
  question: Partial<SavedQuestion>,
): Promise<SavedQuestion> {
  const response = await fetch(`${API_BASE}/questions/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(question),
  })
  return handleResponse(response)
}

export async function deleteSavedQuestion(id: string): Promise<void> {
  const response = await fetch(`${API_BASE}/questions/${id}`, {
    method: 'DELETE',
  })
  if (!response.ok) {
    throw new DashforgeApiError(
      `Failed to delete question: ${response.statusText}`,
      response.status,
    )
  }
}

// Export types
export { DashforgeApiError }
