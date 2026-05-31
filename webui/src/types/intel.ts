export type AttackMatrix = 'Enterprise' | 'ICS'

export interface AttackTechnique {
  tactic: string
  techniqueId: string
  name: string
  description?: string
}

export type IntelFigureKind = 'attack-flow' | 'technique-map' | 'screenshot' | 'other'

/** Image / diagram extracted from intel report (evidence layer). */
export interface IntelFigure {
  id: string
  intelId: number
  title: string
  caption?: string
  kind: IntelFigureKind
  src: string
  techniqueIds?: string[]
}

export interface IntelIoc {
  type: string
  value: string
  confidence?: string
  expiresAt?: string | null
}

/** Pipeline job status — no human review gate */
export type IngestStatus =
  | 'queued'
  | 'parsing'
  | 'extracting'
  | 'succeeded'
  | 'failed'
  | 'archived'

export type KnowledgeLinkAction = 'created' | 'updated' | 'skipped' | 'pending'

/** Intel report (证据层) — structured TTP/entities live on AttackKnowledge (AI-written) */
export interface IntelSummary {
  id: number
  title: string
  sourceUrl: string
  disclosedAt: string | null
  sourceName: string
  collectSourceName: string
  ingestStatus: IngestStatus
  collectedAt: string
  knowledgeIds: number[]
  knowledgeAction: KnowledgeLinkAction
  extractConfidence: number | null
  iocCount: number
  pipelineError?: string
}

export interface IntelAiExtractPreview {
  summary: string
  suggestedTechniques: string[]
  suggestedMalware: string[]
  suggestedCves: string[]
  tools?: string[]
  infrastructure?: string[]
  malwareClasses?: string[]
  qualityHints?: string[]
  qualityOk?: boolean
}

export interface IntelDetail extends IntelSummary {
  sourceUrl: string
  hasRaw?: boolean
  hasParsed?: boolean
  iocs: IntelIoc[]
  originalHtml: string
  /** Parser-extracted figures (plus inline &lt;img&gt; in originalHtml). */
  figures: IntelFigure[]
  parseVersion: string
  enrichVersion: string
  aiExtract: IntelAiExtractPreview | null
}

export interface IntelListParams {
  sourceName?: string
  knowledgeId?: number
  ingestStatus?: IngestStatus
  sortOrder?: 'asc' | 'desc'
}

export interface IntelListResult {
  items: IntelSummary[]
  total: number
}
