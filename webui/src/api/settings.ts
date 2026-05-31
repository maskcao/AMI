import { apiFetch, USE_API } from '@/api/http'
import {
  COLLECT_TYPE_LABELS,
  MOCK_COLLECT_SOURCES,
  MOCK_MODEL_PROVIDERS,
  MOCK_PARSER_PROFILES,
} from '@/mock/settings'
import type {
  CollectSourceSummary,
  ModelProviderSummary,
  ParserProfileSummary,
} from '@/types/settings'

export interface CollectSourcePagination {
  type?: 'none' | 'path' | 'scroll'
  path_template?: string
  start_page?: number
  max_pages?: number
  stop_on_empty_list?: boolean
  scroll_steps?: number
}

export interface CollectSourceLinkExtract {
  must_contain?: string
  exclude_contains?: string[]
  prefer_path_regex?: string
}

export interface CollectSourceConfig {
  article_urls?: string[]
  feed_url?: string
  list_url?: string
  entry_url?: string
  link_selector?: string
  article_base?: string
  link_extract?: CollectSourceLinkExtract
  pagination?: CollectSourcePagination
  max_items_per_run?: number
  timeout_sec?: number
  delay_ms?: number
}

export interface CollectSourceDetail extends CollectSourceSummary {
  config: CollectSourceConfig
}

function parseCollectConfig(raw: unknown): CollectSourceConfig {
  if (raw == null) return {}
  if (typeof raw === 'string') {
    try {
      return JSON.parse(raw) as CollectSourceConfig
    } catch {
      return {}
    }
  }
  return raw as CollectSourceConfig
}

export interface CollectRunResult {
  sourceId: string
  imported?: number
  skipped?: number
  failed?: number
  error?: string
}

export async function listCollectSources(): Promise<CollectSourceDetail[]> {
  if (USE_API) {
    const res = await apiFetch<{ items: CollectSourceDetail[] }>('/api/v1/settings/collect-sources')
    return res.items.map((r) => ({
      ...r,
      config: parseCollectConfig(r.config),
    }))
  }
  return MOCK_COLLECT_SOURCES.map((r) => ({ ...r, config: {} }))
}

export async function createCollectSource(body: {
  name: string
  type: CollectSourceDetail['type']
  enabled?: boolean
  scheduleCron?: string
  scheduleLabel?: string
  sourceLabel?: string
  config?: CollectSourceConfig
}): Promise<CollectSourceDetail> {
  return apiFetch<CollectSourceDetail>('/api/v1/settings/collect-sources', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function patchCollectSource(
  id: string,
  body: Partial<{
    name: string
    enabled: boolean
    scheduleCron: string
    scheduleLabel: string
    sourceLabel: string
    config: CollectSourceConfig
  }>,
): Promise<CollectSourceDetail> {
  return apiFetch<CollectSourceDetail>(`/api/v1/settings/collect-sources/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}

export async function runCollectSource(id: string): Promise<CollectRunResult> {
  return apiFetch<CollectRunResult>(`/api/v1/settings/collect-sources/${encodeURIComponent(id)}/run`, {
    method: 'POST',
  })
}

export async function deleteCollectSource(id: string): Promise<void> {
  await apiFetch<void>(`/api/v1/settings/collect-sources/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export async function listModelProviders(): Promise<ModelProviderSummary[]> {
  if (USE_API) {
    const res = await apiFetch<{ items: ModelProviderSummary[] }>('/api/v1/settings/model-providers')
    return res.items
  }
  return MOCK_MODEL_PROVIDERS
}

export async function createModelProvider(body: {
  name: string
  kind?: ModelProviderSummary['kind']
  modelId?: string
  baseUrl?: string
  apiKey?: string
  enabled?: boolean
  defaultFor?: string
}): Promise<ModelProviderSummary> {
  return apiFetch<ModelProviderSummary>('/api/v1/settings/model-providers', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function deleteModelProvider(id: string): Promise<void> {
  await apiFetch<void>(`/api/v1/settings/model-providers/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export async function testModelProvider(id: string): Promise<{ ok: boolean; error?: string }> {
  try {
    return await apiFetch<{ ok: boolean; error?: string }>(
      `/api/v1/settings/model-providers/${encodeURIComponent(id)}/test`,
      { method: 'POST' },
    )
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : 'test failed' }
  }
}

export async function patchModelProvider(
  id: string,
  body: Partial<{
    name: string
    modelId: string
    baseUrl: string
    apiKey: string
    enabled: boolean
    defaultFor: string
  }>,
): Promise<ModelProviderSummary> {
  return apiFetch<ModelProviderSummary>(`/api/v1/settings/model-providers/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}

export async function createParserProfile(body: {
  name: string
  description?: string
}): Promise<ParserProfileSummary> {
  return apiFetch<ParserProfileSummary>('/api/v1/settings/parser-profiles', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function patchParserProfile(
  id: string,
  body: Partial<{ name: string; description: string }>,
): Promise<ParserProfileSummary> {
  return apiFetch<ParserProfileSummary>(`/api/v1/settings/parser-profiles/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}

export async function purgeAllData(): Promise<void> {
  await apiFetch<{ status: string }>('/api/v1/admin/purge-data', { method: 'POST' })
}

export async function deleteParserProfile(id: string): Promise<void> {
  await apiFetch<void>(`/api/v1/settings/parser-profiles/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export async function listParserProfiles(): Promise<ParserProfileSummary[]> {
  if (USE_API) {
    const res = await apiFetch<{ items: ParserProfileSummary[] }>('/api/v1/settings/parser-profiles')
    return res.items
  }
  return MOCK_PARSER_PROFILES
}

export { COLLECT_TYPE_LABELS }
