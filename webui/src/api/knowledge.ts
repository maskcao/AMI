import { apiFetch, USE_API } from '@/api/http'
import { hydrateAnalytics, invalidateAnalyticsCache } from '@/api/analytics'
import {
  KNOWLEDGE_DETAILS,
  getKnowledgeById,
  mergeKnowledgeInStore,
  updateKnowledgeInStore,
} from '@/mock/knowledge'
import type {
  KnowledgeDetail,
  KnowledgeListParams,
  KnowledgeListResult,
  KnowledgeMergePayload,
  KnowledgeSummary,
  KnowledgeUpdatePayload,
} from '@/types/knowledge'
import { earliestDisclosedAt } from '@/utils/disclosed-at'
import { computeNeedsReview } from '@/utils/knowledge-review'
import { normalizeMalwareRefs } from '@/utils/malware-ref'

const MOCK_DELAY_MS = 180

function delay<T>(value: T): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), MOCK_DELAY_MS))
}

function toSummary(d: KnowledgeDetail): KnowledgeSummary {
  return {
    id: d.id,
    title: d.title,
    summary: d.summary,
    status: d.status,
    needsReview: d.needsReview,
    targetIndustries: d.targetIndustries,
    attackMatrix: d.attackMatrix,
    malware: d.malware,
    threatActors: d.threatActors,
    qualityScore: d.qualityScore,
    intelCount: d.intelCount,
    disclosedAt: earliestDisclosedAt(d.linkedIntel),
    updatedAt: d.updatedAt,
    dimensions: d.dimensions,
  }
}

function matchesQuery(d: KnowledgeDetail, q: string): boolean {
  const lower = q.toLowerCase()
  if (d.title.toLowerCase().includes(lower)) return true
  if (d.summary?.toLowerCase().includes(lower)) return true
  if (d.threatActors.some((a) => a.toLowerCase().includes(lower))) return true
  if (d.malware.some((m) => m.name.toLowerCase().includes(lower))) return true
  if (d.techniques.some((t) => t.techniqueId.toLowerCase().includes(lower))) return true
  if (d.techniques.some((t) => t.name.toLowerCase().includes(lower))) return true
  if (d.vulnerabilities.some((v) => v.cveId.toLowerCase().includes(lower))) return true
  return false
}

function applyReviewFilter(
  items: KnowledgeSummary[],
  review: KnowledgeListParams['review'],
): KnowledgeSummary[] {
  if (!review || review === 'all') return items
  if (review === 'needs_review') return items.filter((i) => i.needsReview)
  if (review === 'draft') return items.filter((i) => i.status === 'draft')
  if (review === 'published') return items.filter((i) => i.status === 'published')
  return items
}

export async function listKnowledge(
  params: KnowledgeListParams = {},
): Promise<KnowledgeListResult> {
  if (USE_API) {
    const q = new URLSearchParams()
    if (params.q?.trim()) q.set('q', params.q.trim())
    if (params.review === 'needs_review') q.set('needsReview', 'true')
    if (params.review === 'llm_failed') q.set('llmFailed', 'true')
    if (params.review === 'published') q.set('status', 'published')
    if (params.review === 'draft') q.set('status', 'draft')
    if (params.matrix) q.set('matrix', params.matrix)
    if (params.industries?.length) q.set('industry', params.industries[0])
    if (params.malwareClass) q.set('malwareClass', params.malwareClass)
    const qs = q.toString()
    const res = await apiFetch<KnowledgeListResult>(`/api/v1/knowledge${qs ? `?${qs}` : ''}`)
    const needsReviewCount = res.items.filter((i) => i.needsReview).length
    return {
      ...res,
      needsReviewCount,
      items: res.items.map((i) => ({ ...i, malware: normalizeMalwareRefs(i.malware) })),
    }
  }

  const needsReviewCount = KNOWLEDGE_DETAILS.filter((d) =>
    computeNeedsReview({
      status: d.status,
      qualityScore: d.qualityScore,
      intelCount: d.intelCount,
    }),
  ).length

  let details = [...KNOWLEDGE_DETAILS]

  if (params.industries?.length) {
    details = details.filter((d) =>
      params.industries!.some((ind) => d.targetIndustries.includes(ind)),
    )
  }
  if (params.matrix) {
    details = details.filter((d) => d.attackMatrix === params.matrix)
  }
  if (params.malwareClass) {
    details = details.filter((d) => d.malware.some((m) => m.class === params.malwareClass))
  }
  if (params.q?.trim()) {
    const q = params.q.trim()
    details = details.filter((d) => matchesQuery(d, q))
  }

  let items = details.map(toSummary)
  items = applyReviewFilter(items, params.review)

  const order = params.sortOrder === 'asc' ? 1 : -1
  items.sort((a, b) => a.updatedAt.localeCompare(b.updatedAt) * order)

  return delay({ items, total: items.length, needsReviewCount })
}

function normalizeKnowledgeDetail(row: KnowledgeDetail): KnowledgeDetail {
  return {
    ...row,
    malware: normalizeMalwareRefs(row.malware),
    linkedIntel: Array.isArray(row.linkedIntel) ? row.linkedIntel : [],
    techniques: Array.isArray(row.techniques) ? row.techniques : [],
    evidenceFigures: Array.isArray(row.evidenceFigures) ? row.evidenceFigures : [],
  }
}

export async function getKnowledge(id: number): Promise<KnowledgeDetail | null> {
  if (USE_API) {
    try {
      const row = await apiFetch<KnowledgeDetail>(`/api/v1/knowledge/${id}`)
      return row ? normalizeKnowledgeDetail(row) : null
    } catch {
      return null
    }
  }
  return delay(getKnowledgeById(id) ?? null)
}

export async function updateKnowledge(
  id: number,
  payload: KnowledgeUpdatePayload,
): Promise<KnowledgeDetail | null> {
  if (USE_API) {
    try {
      return await apiFetch<KnowledgeDetail>(`/api/v1/knowledge/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(payload),
      })
    } catch {
      return null
    }
  }
  return delay(updateKnowledgeInStore(id, payload))
}

export async function publishKnowledge(id: number): Promise<KnowledgeDetail | null> {
  if (USE_API) {
    try {
      const detail = await apiFetch<KnowledgeDetail>(`/api/v1/knowledge/${id}/publish`, {
        method: 'POST',
      })
      invalidateAnalyticsCache()
      await hydrateAnalytics(true)
      return detail
    } catch {
      return null
    }
  }
  return delay(updateKnowledgeInStore(id, { status: 'published' }))
}

export async function retryLLMFailed(limit = 200): Promise<{ requeued: number; remaining: number } | null> {
  if (USE_API) {
    try {
      return await apiFetch<{ requeued: number; remaining: number }>(
        `/api/v1/intel/llm-retry?limit=${limit}`,
        { method: 'POST' },
      )
    } catch {
      return null
    }
  }
  return { requeued: 0, remaining: 0 }
}

export async function mergeKnowledge(payload: KnowledgeMergePayload): Promise<boolean> {
  if (USE_API) {
    try {
      await apiFetch<{ ok: boolean }>(`/api/v1/knowledge/merge`, {
        method: 'POST',
        body: JSON.stringify(payload),
      })
      return true
    } catch {
      return false
    }
  }
  return delay(mergeKnowledgeInStore(payload.sourceId, payload.targetId))
}
