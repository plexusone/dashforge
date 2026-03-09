/**
 * Integration Platform Types
 * Types for integrations, alerts, and the marketplace.
 */

// Channel types
export type ChannelType = 'slack' | 'whatsapp' | 'email' | 'webhook'

// Integration source
export type IntegrationSource = 'builtin' | 'marketplace' | 'custom'

// Integration status
export type IntegrationStatus = 'active' | 'inactive' | 'error'

// Alert trigger types
export type TriggerType = 'threshold' | 'schedule' | 'data_change'

// Severity levels
export type Severity = 'info' | 'warning' | 'error' | 'critical'

// Alert event types
export type AlertEventType = 'triggered' | 'resolved' | 'acknowledged' | 'error' | 'cooldown_skipped'

// Channel capabilities
export interface ChannelCapabilities {
  supportsRichText: boolean
  supportsAttachments: boolean
  supportsThreading: boolean
  supportsReactions: boolean
  maxMessageLength: number
  requiresRecipient: boolean
  supportsBatching: boolean
}

// Channel info from registry
export interface ChannelInfo {
  type: ChannelType
  name: string
  capabilities: ChannelCapabilities
}

// Integration entity
export interface Integration {
  id: number
  name: string
  slug: string
  channelType: ChannelType
  config?: Record<string, unknown>
  status: IntegrationStatus
  statusMessage?: string
  source: IntegrationSource
  marketplaceSlug?: string
  lastUsedAt?: string
  createdAt: string
  updatedAt: string
}

// Integration with credentials (for creation/update)
export interface IntegrationWithCredentials extends Omit<Integration, 'id' | 'createdAt' | 'updatedAt' | 'status' | 'statusMessage' | 'lastUsedAt'> {
  credentials?: Record<string, unknown>
}

// Alert entity
export interface Alert {
  id: number
  name: string
  slug: string
  description?: string
  triggerType: TriggerType
  triggerConfig: Record<string, unknown>
  enabled: boolean
  cooldownSeconds: number
  consecutiveFailures: number
  lastTriggeredAt?: string
  lastEvaluatedAt?: string
  lastError?: string
  channels?: IntegrationSummary[]
  createdAt: string
  updatedAt: string
}

// Slim integration info for alert responses
export interface IntegrationSummary {
  id: number
  name: string
  slug: string
  channelType: ChannelType
  status: IntegrationStatus
}

// Alert event (history)
export interface AlertEvent {
  id: number
  eventType: AlertEventType
  triggerData?: Record<string, unknown>
  channelsNotified?: string[]
  channelsSuccess: number
  channelsFailed: number
  errorMessage?: string
  acknowledgedBy?: string
  createdAt: string
}

// Trigger configuration types
export interface ThresholdTriggerConfig {
  metricField: string
  operator: 'gt' | 'gte' | 'lt' | 'lte' | 'eq' | 'neq'
  value: number
  datasourceSlug?: string
  query?: string
  severity?: Severity
}

export interface ScheduleTriggerConfig {
  cron: string
  timezone?: string
  message?: string
  severity?: Severity
}

export interface DataChangeTriggerConfig {
  datasourceSlug: string
  query: string
  changeType: 'any' | 'increase' | 'decrease' | 'new_rows' | 'deleted_rows'
  compareField?: string
  severity?: Severity
  message?: string
}

// Marketplace integration definition
export interface IntegrationDefinition {
  slug: string
  name: string
  description: string
  channelType: ChannelType
  category: string
  icon: string
  version: string
  author: string
  source: IntegrationSource
  configSchema: ConfigSchema
  capabilities: ChannelCapabilities
}

// JSON Schema for config fields
export interface ConfigSchema {
  type: 'object'
  properties: Record<string, ConfigSchemaProperty>
  required?: string[]
}

export interface ConfigSchemaProperty {
  type: 'string' | 'integer' | 'boolean' | 'object'
  title: string
  description?: string
  default?: unknown
  enum?: string[]
  secret?: boolean
}

// API request/response types

export interface CreateIntegrationRequest {
  name: string
  slug: string
  channelType: ChannelType
  config?: Record<string, unknown>
  credentials?: Record<string, unknown>
  source?: IntegrationSource
}

export interface UpdateIntegrationRequest {
  name?: string
  config?: Record<string, unknown>
  credentials?: Record<string, unknown>
  status?: IntegrationStatus
}

export interface CreateAlertRequest {
  name: string
  slug: string
  description?: string
  triggerType: TriggerType
  triggerConfig: Record<string, unknown>
  cooldownSeconds?: number
  enabled?: boolean
  dashboardId?: number
  channelIds?: number[]
}

export interface UpdateAlertRequest {
  name?: string
  description?: string
  triggerConfig?: Record<string, unknown>
  cooldownSeconds?: number
  enabled?: boolean
  channelIds?: number[]
}

export interface InstallIntegrationRequest {
  name?: string
  slug: string
  config?: Record<string, unknown>
  credentials?: Record<string, unknown>
}

export interface TestConnectionResult {
  success: boolean
  error?: string
  durationMs: number
}

export interface ListIntegrationsResponse {
  integrations: Integration[]
  total: number
}

export interface ListAlertsResponse {
  alerts: Alert[]
  total: number
}

export interface ListAlertEventsResponse {
  events: AlertEvent[]
  total: number
}

export interface ListMarketplaceResponse {
  integrations: IntegrationDefinition[]
  total: number
}

export interface ListChannelsResponse {
  channels: ChannelInfo[]
}
