import { INTEL_DETAILS, getIntelById } from '@/mock/intel'
import type { IntelDetail, IntelListParams, IntelListResult, IntelSummary } from '@/types/intel'

const MOCK_DELAY_MS = 180

function delay<T>(value: T): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), MOCK_DELAY_MS))
}

function toSummary(d: IntelDetail): IntelSummary {
  return {
    id: d.id,
    title: d.title,
    disclosedAt: d.disclosedAt,
    sourceName: d.sourceName,
    ingestStatus: d.ingestStatus,
    knowledgeIds: d.knowledgeIds,
    iocCount: d.iocCount,
    qualityScore: d.qualityScore,
  }
}

export async function listIntel(params: IntelListParams = {}): Promise<IntelListResult> {
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
  return delay(getIntelById(id) ?? null)
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
