import type { LinkedKnowledgeRef } from '@/analytics/actors'
import type {
  AttackInfrastructure,
  AttackTool,
  ExploitedVulnerability,
  MalwareRef,
} from '@/types/knowledge'
import type { AttackTechnique } from '@/types/intel'

export interface ThreatActorListItem {
  name: string
  slug: string
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
  aliases: string[]
  summary: string | null
}

export interface ThreatActorProfile extends ThreatActorListItem {
  tools: AttackTool[]
  infrastructure: AttackInfrastructure[]
  malware: MalwareRef[]
  vulnerabilities: ExploitedVulnerability[]
  techniques: AttackTechnique[]
  intelReportCount: number
  matrices: string[]
  firstSeen: string | null
  lastSeen: string | null
}
