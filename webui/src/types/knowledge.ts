import type { AttackMatrix, AttackTechnique, IntelFigure, KnowledgeLinkAction } from '@/types/intel'
import type { MalwareClass } from '@/constants/malware'

export type KnowledgeStatus = 'draft' | 'published'

export interface MalwareRef {
  name: string
  /** API may send `classification`; normalized to `class` in client. */
  class?: MalwareClass
  classification?: string
}

export interface AttackTool {
  name: string
  category?: string
  description?: string
}

export interface AttackInfrastructure {
  type: 'domain' | 'ip' | 'url' | 'email'
  value: string
  role: 'c2' | 'hosting' | 'payload' | 'scan' | 'other'
  confidence?: 'low' | 'medium' | 'high'
  note?: string
}

export interface ExploitedVulnerability {
  cveId: string
  description?: string
  inWild?: boolean
}

export interface KnowledgeProcedure {
  techniqueId: string
  text: string
  evidenceIntelId?: number
}

export interface KnowledgePipelineSource {
  lastIntelId: number
  lastIntelTitle: string
  knowledgeAction: KnowledgeLinkAction
  extractConfidence: number | null
  pipelinedAt: string
}

export interface KnowledgeDimensionCounts {
  techniques: number
  tools: number
  infrastructure: number
  vulnerabilities: number
  malware: number
}

export interface KnowledgeSummary {
  id: number
  title: string
  summary: string | null
  status: KnowledgeStatus
  needsReview: boolean
  targetIndustries: string[]
  attackMatrix: AttackMatrix
  malware: MalwareRef[]
  threatActors: string[]
  qualityScore: number
  intelCount: number
  updatedAt: string
  dimensions: KnowledgeDimensionCounts
}

export interface LinkedIntelBrief {
  id: number
  title: string
  sourceName: string
  sourceUrl?: string
  collectSourceName?: string
  disclosedAt: string | null
  role: 'primary' | 'supporting'
  hasRaw?: boolean
  hasParsed?: boolean
}

export interface KnowledgeDetail extends KnowledgeSummary {
  qualityHints: string[]
  techniques: AttackTechnique[]
  tools: AttackTool[]
  infrastructure: AttackInfrastructure[]
  vulnerabilities: ExploitedVulnerability[]
  linkedIntel: LinkedIntelBrief[]
  procedures: KnowledgeProcedure[]
  pipeline: KnowledgePipelineSource | null
  /** Diagrams / screenshots aggregated from linked intel reports. */
  evidenceFigures: IntelFigure[]
}

export type KnowledgeReviewFilter = 'all' | 'needs_review' | 'draft' | 'published'

export interface KnowledgeListParams {
  industries?: string[]
  matrix?: AttackMatrix
  malwareClass?: MalwareClass
  q?: string
  review?: KnowledgeReviewFilter
  sortOrder?: 'asc' | 'desc'
}

export interface KnowledgeListResult {
  items: KnowledgeSummary[]
  total: number
  needsReviewCount: number
}

export interface KnowledgeUpdatePayload {
  title?: string
  summary?: string | null
  targetIndustries?: string[]
  status?: KnowledgeStatus
}

export interface KnowledgeMergePayload {
  sourceId: number
  targetId: number
}
