/** Collect source type — see docs/system-config.md */
export type CollectSourceType =
  | 'direct_url'
  | 'paginated_list'
  | 'threat_portal'
  | 'scheduled_url'
  | 'wechat_feed'
  | 'rss'
  | 'manual_upload'
  | 'webhook'
  | 'stix_bundle'
  | 'email_imap'
  | 'api_pull'

export type CollectRunStatus = 'success' | 'partial' | 'failed' | 'skipped' | 'never'

export interface CollectSourceSummary {
  id: string
  name: string
  type: CollectSourceType
  typeLabel: string
  enabled: boolean
  scheduleCron: string
  scheduleLabel: string
  sourceLabel: string
  lastRunAt: string | null
  lastStatus: CollectRunStatus
}

export type ModelProviderKind = 'llm' | 'embedding'

export interface ModelProviderSummary {
  id: string
  name: string
  kind: ModelProviderKind
  modelId: string
  baseUrl: string
  enabled: boolean
  defaultFor: string
  apiKeyConfigured?: boolean
}

export interface ParserProfileSummary {
  id: string
  name: string
  description: string
  usedBySources: number
}
