import { apiFetch, USE_API } from '@/api/http'
import { INTEL_DETAILS, getIntelById } from '@/mock/intel'
import type { IntelDetail, IntelListParams, IntelListResult, IntelSummary } from '@/types/intel'

const MOCK_DELAY_MS = 180

function delay<T>(value: T): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), MOCK_DELAY_MS))
}

function asStringArray(v: unknown): string[] {
  return Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string') : []
}

function normalizeAiExtract(raw: unknown): IntelDetail['aiExtract'] {
  if (!raw || typeof raw !== 'object') return null
  const o = raw as Record<string, unknown>
  const summary = typeof o.summary === 'string' ? o.summary : ''
  if (!summary) return null
  return {
    summary,
    suggestedTechniques: asStringArray(o.suggestedTechniques),
    suggestedMalware: asStringArray(o.suggestedMalware),
    suggestedCves: asStringArray(o.suggestedCves),
    tools: asStringArray(o.tools),
    infrastructure: asStringArray(o.infrastructure),
    malwareClasses: asStringArray(o.malwareClasses),
    qualityHints: asStringArray(o.qualityHints),
    qualityOk: typeof o.qualityOk === 'boolean' ? o.qualityOk : undefined,
  }
}

function normalizeIntelDetail(row: IntelDetail): IntelDetail {
  return {
    ...row,
    sourceUrl: row.sourceUrl ?? '',
    hasRaw: !!row.hasRaw,
    hasParsed: !!row.hasParsed,
    knowledgeIds: Array.isArray(row.knowledgeIds) ? row.knowledgeIds : [],
    figures: Array.isArray(row.figures) ? row.figures : [],
    iocs: Array.isArray(row.iocs) ? row.iocs : [],
    aiExtract: normalizeAiExtract(row.aiExtract),
    parseVersion: row.parseVersion ?? '',
    enrichVersion: row.enrichVersion ?? '',
    originalHtml: row.originalHtml ?? '',
  }
}

function normalizeIntelSummary(r: IntelSummary): IntelSummary {
  return {
    ...r,
    sourceUrl: r.sourceUrl ?? '',
    knowledgeIds: Array.isArray(r.knowledgeIds) ? r.knowledgeIds : [],
    extractConfidence: r.extractConfidence ?? null,
    iocCount: r.iocCount ?? 0,
  }
}

function toSummary(d: IntelDetail): IntelSummary {
  const {
    id,
    title,
    sourceUrl,
    disclosedAt,
    sourceName,
    collectSourceName,
    ingestStatus,
    collectedAt,
    knowledgeIds,
    knowledgeAction,
    extractConfidence,
    iocCount,
    pipelineError,
  } = d
  return {
    id,
    title,
    sourceUrl,
    disclosedAt,
    sourceName,
    collectSourceName,
    ingestStatus,
    collectedAt,
    knowledgeIds,
    knowledgeAction,
    extractConfidence,
    iocCount,
    pipelineError,
  }
}

export async function listIntel(params: IntelListParams = {}): Promise<IntelListResult> {
  if (USE_API) {
    const q = new URLSearchParams()
    if (params.sourceName) q.set('sourceName', params.sourceName)
    if (params.knowledgeId != null) q.set('knowledgeId', String(params.knowledgeId))
    if (params.ingestStatus) q.set('ingestStatus', params.ingestStatus)
    if (params.sortOrder) q.set('sortOrder', params.sortOrder)
    const qs = q.toString()
    const res = await apiFetch<IntelListResult>(`/api/v1/intel${qs ? `?${qs}` : ''}`)
    return {
      total: res.total ?? res.items.length,
      items: res.items.map(normalizeIntelSummary),
    }
  }

  let items = INTEL_DETAILS.map(toSummary)

  if (params.sourceName) {
    items = items.filter((item) => item.sourceName === params.sourceName)
  }
  if (params.knowledgeId != null) {
    items = items.filter((item) => item.knowledgeIds.includes(params.knowledgeId!))
  }
  if (params.ingestStatus) {
    items = items.filter((item) => item.ingestStatus === params.ingestStatus)
  }

  const order = params.sortOrder === 'asc' ? 1 : -1
  items.sort((a, b) => (a.id - b.id) * order)

  return delay({ items, total: items.length })
}

export async function getIntel(id: number): Promise<IntelDetail | null> {
  if (USE_API) {
    try {
      const row = await apiFetch<IntelDetail>(`/api/v1/intel/${id}`)
      return row ? normalizeIntelDetail(row) : null
    } catch {
      return null
    }
  }
  return delay(getIntelById(id) ?? null)
}

/** Re-queue parse → extract pipeline for one intel report. */
export async function retryIntel(id: number): Promise<IntelDetail | null> {
  if (USE_API) {
    const row = await apiFetch<IntelDetail>(`/api/v1/intel/${id}/retry`, { method: 'POST' })
    return row ? normalizeIntelDetail(row) : null
  }
  const d = getIntelById(id)
  if (!d) return null
  return delay({
    ...d,
    ingestStatus: 'queued',
    pipelineError: undefined,
    knowledgeAction: 'pending',
  })
}

/** Load full original HTML from file store (lazy). */
export async function getIntelOriginal(id: number): Promise<string | null> {
  if (!USE_API) return getIntel(id).then((d) => d?.originalHtml ?? null)
  const res = await fetch(`/api/v1/intel/${id}/original`)
  if (!res.ok) return null
  return res.text()
}

/** Parsed markdown body (LLM input / body.md); alias of markdown endpoint. */
export async function getIntelParsed(id: number): Promise<string | null> {
  if (!USE_API) return getIntel(id).then((d) => d?.originalHtml ?? null)
  const res = await fetch(`/api/v1/intel/${id}/parsed`)
  if (!res.ok) return null
  return res.text()
}

/** Markdown converted from archived HTML (body.md). */
export async function getIntelMarkdown(id: number): Promise<string | null> {
  if (!USE_API) return null
  const res = await fetch(`/api/v1/intel/${id}/markdown`)
  if (!res.ok) return null
  return res.text()
}

/** IOC helper search for /ioc demo page */
export async function searchIoc(query: string): Promise<
  {
    ioc: { type: string; value: string; confidence?: string }
    intelId: number
    intelTitle: string
    knowledgeIds: number[]
  }[]
> {
  if (USE_API) {
    const q = query.trim()
    if (!q) return []
    return apiFetch(`/api/v1/ioc?q=${encodeURIComponent(q)}`)
  }

  const q = query.trim().toLowerCase()
  if (!q) return delay([])

  const hits: {
    ioc: { type: string; value: string; confidence?: string }
    intelId: number
    intelTitle: string
    knowledgeIds: number[]
  }[] = []

  for (const intel of INTEL_DETAILS) {
    for (const ioc of intel.iocs) {
      if (ioc.value.toLowerCase().includes(q) || ioc.type.toLowerCase().includes(q)) {
        hits.push({
          ioc,
          intelId: intel.id,
          intelTitle: intel.title,
          knowledgeIds: intel.knowledgeIds,
        })
      }
    }
  }
  return delay(hits)
}
