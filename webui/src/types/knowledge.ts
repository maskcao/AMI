import type { AttackMatrix, AttackTechnique } from '@/types/intel'
import type { MalwareClass } from '@/constants/malware'

export interface MalwareRef {
  name: string
  class: MalwareClass
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
  disclosedAt: string | null
  role: 'primary' | 'supporting'
}

export interface KnowledgeDetail extends KnowledgeSummary {
  summary: string | null
  qualityHints: string[]
  techniques: AttackTechnique[]
  tools: AttackTool[]
  infrastructure: AttackInfrastructure[]
  vulnerabilities: ExploitedVulnerability[]
  linkedIntel: LinkedIntelBrief[]
}

export interface KnowledgeListParams {
  industries?: string[]
  matrix?: AttackMatrix
  malwareClass?: MalwareClass
  q?: string
  sortOrder?: 'asc' | 'desc'
}

export interface KnowledgeListResult {
  items: KnowledgeSummary[]
  total: number
}
