import type { CountItem, DashboardOperationalStats, DashboardPayload } from '@/analytics/aggregate'

function countItems(v: unknown): CountItem[] {
  if (!Array.isArray(v)) return []
  return v as CountItem[]
}

export function emptyDashboardPayload(): DashboardPayload {
  const snapshot = {
    knowledgeCount: 0,
    cveCount: 0,
    malwareCount: 0,
    actorCount: 0,
    techniqueCount: 0,
    toolCount: 0,
    infraCount: 0,
    iocCount: 0,
    industryDist: [] as CountItem[],
    matrixDist: [] as CountItem[],
    malwareClassDist: [] as CountItem[],
    tacticDist: [] as CountItem[],
  }
  const operations: DashboardOperationalStats = {
    knowledgeCount: 0,
    intelSourceCount: 0,
    intelTotalCount: 0,
    intelSucceededCount: 0,
    intelProcessingCount: 0,
    intelFailedCount: 0,
    intelLinkedCount: 0,
    sourceLabels: [],
  }
  return {
    snapshot,
    operations,
    knowledgeRows: [],
    ransomwareKnowledgeCount: 0,
    icsKnowledgeCount: 0,
    inWildCveCount: 0,
  }
}

export function normalizeDashboardPayload(raw: Partial<DashboardPayload>): DashboardPayload {
  const base = emptyDashboardPayload()
  const snap = (raw.snapshot ?? {}) as Record<string, unknown>
  const ops = (raw.operations ?? {}) as Partial<DashboardOperationalStats>
  return {
    snapshot: {
      ...base.snapshot,
      knowledgeCount: Number(snap.knowledgeCount ?? 0),
      cveCount: Number(snap.cveCount ?? 0),
      malwareCount: Number(snap.malwareCount ?? 0),
      actorCount: Number(snap.actorCount ?? 0),
      techniqueCount: Number(snap.techniqueCount ?? 0),
      toolCount: Number(snap.toolCount ?? 0),
      infraCount: Number(snap.infraCount ?? 0),
      iocCount: Number(snap.iocCount ?? 0),
      industryDist: countItems(snap.industryDist),
      matrixDist: countItems(snap.matrixDist),
      malwareClassDist: countItems(snap.malwareClassDist),
      tacticDist: countItems(snap.tacticDist),
    },
    operations: {
      knowledgeCount: Number(ops.knowledgeCount ?? 0),
      intelSourceCount: Number(ops.intelSourceCount ?? 0),
      intelTotalCount: Number(ops.intelTotalCount ?? 0),
      intelSucceededCount: Number(ops.intelSucceededCount ?? 0),
      intelProcessingCount: Number(ops.intelProcessingCount ?? 0),
      intelFailedCount: Number(ops.intelFailedCount ?? 0),
      intelLinkedCount: Number(ops.intelLinkedCount ?? 0),
      sourceLabels: Array.isArray(ops.sourceLabels) ? ops.sourceLabels : [],
    },
    knowledgeRows: Array.isArray(raw.knowledgeRows) ? raw.knowledgeRows : [],
    ransomwareKnowledgeCount: Number(raw.ransomwareKnowledgeCount ?? 0),
    icsKnowledgeCount: Number(raw.icsKnowledgeCount ?? 0),
    inWildCveCount: Number(raw.inWildCveCount ?? 0),
  }
}
