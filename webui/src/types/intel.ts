export type AttackMatrix = 'Enterprise' | 'ICS'

export interface AttackTechnique {
  tactic: string
  techniqueId: string
  name: string
  description?: string
}

export interface IntelIoc {
  type: string
  value: string
  confidence?: string
  expiresAt?: string | null
}

/** Intel report (证据层) — TTP/summary live on AttackKnowledge */
export interface IntelSummary {
  id: number
  title: string
  disclosedAt: string | null
  sourceName: string
  ingestStatus: 'draft' | 'parsing' | 'pending_review' | 'published' | 'archived'
  knowledgeIds: number[]
  iocCount: number
  qualityScore: number
}

export interface IntelDetail extends IntelSummary {
  sourceUrl: string
  qualityHints: string[]
  iocs: IntelIoc[]
  originalHtml: string
  collectedAt: string
  parseVersion: string
}

export interface IntelListParams {
  sourceName?: string
  knowledgeId?: number
  ingestStatus?: string
  sortOrder?: 'asc' | 'desc'
}

export interface IntelListResult {
  items: IntelSummary[]
  total: number
}
