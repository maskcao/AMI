import type { LinkedKnowledgeRef } from '@/analytics/actors'
import type { MalwareClass } from '@/constants/malware'
import type {
  AttackInfrastructure,
  AttackTool,
  ExploitedVulnerability,
  MalwareRef,
} from '@/types/knowledge'
import type { AttackTechnique } from '@/types/intel'

export type EntityDimensionKind = 'cve' | 'malware' | 'tool' | 'infra'

export interface EntityDimensionProfile {
  dimension: EntityDimensionKind
  title: string
  subtitle: string
  slug: string
  summary: string | null
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
  threatActors: string[]
  techniques: AttackTechnique[]
  tools: AttackTool[]
  infrastructure: AttackInfrastructure[]
  malware: MalwareRef[]
  vulnerabilities: ExploitedVulnerability[]
  intelReportCount: number
  matrices: string[]
  inWild?: boolean
  cveId?: string
  malwareClass?: MalwareClass
  toolCategory?: string
  infraType?: string
  infraRole?: string
}

export interface EntityListRowBase {
  slug: string
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
}
