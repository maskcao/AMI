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
import { computeNeedsReview } from '@/utils/knowledge-review'

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

export async function getKnowledge(id: number): Promise<KnowledgeDetail | null> {
  return delay(getKnowledgeById(id) ?? null)
}

export async function updateKnowledge(
  id: number,
  payload: KnowledgeUpdatePayload,
): Promise<KnowledgeDetail | null> {
  return delay(updateKnowledgeInStore(id, payload))
}

export async function publishKnowledge(id: number): Promise<KnowledgeDetail | null> {
  return delay(updateKnowledgeInStore(id, { status: 'published' }))
}

export async function mergeKnowledge(payload: KnowledgeMergePayload): Promise<boolean> {
  return delay(mergeKnowledgeInStore(payload.sourceId, payload.targetId))
}
