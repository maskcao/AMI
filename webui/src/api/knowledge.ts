import { KNOWLEDGE_DETAILS, getKnowledgeById } from '@/mock/knowledge'
import type {
  KnowledgeDetail,
  KnowledgeListParams,
  KnowledgeListResult,
  KnowledgeSummary,
} from '@/types/knowledge'

const MOCK_DELAY_MS = 180

function delay<T>(value: T): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), MOCK_DELAY_MS))
}

function toSummary(d: KnowledgeDetail): KnowledgeSummary {
  return {
    id: d.id,
    title: d.title,
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

export async function listKnowledge(
  params: KnowledgeListParams = {},
): Promise<KnowledgeListResult> {
  let items = KNOWLEDGE_DETAILS.map(toSummary)

  if (params.industries?.length) {
    items = items.filter((item) =>
      params.industries!.some((ind) => item.targetIndustries.includes(ind)),
    )
  }
  if (params.matrix) {
    items = items.filter((item) => item.attackMatrix === params.matrix)
  }
  if (params.malwareClass) {
    items = items.filter((item) => item.malware.some((m) => m.class === params.malwareClass))
  }
  if (params.q?.trim()) {
    const q = params.q.trim().toLowerCase()
    items = items.filter(
      (item) =>
        item.title.toLowerCase().includes(q) ||
        item.threatActors.some((a) => a.toLowerCase().includes(q)) ||
        item.malware.some((m) => m.name.toLowerCase().includes(q)),
    )
  }

  const order = params.sortOrder === 'asc' ? 1 : -1
  items.sort((a, b) => a.updatedAt.localeCompare(b.updatedAt) * order)

  return delay({ items, total: items.length })
}

export async function getKnowledge(id: number): Promise<KnowledgeDetail | null> {
  return delay(getKnowledgeById(id) ?? null)
}
