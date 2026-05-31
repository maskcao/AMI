export interface LinkedKnowledgeRef {
  knowledgeId: number
  knowledgeTitle: string
}

function knowledgeRef(kid: number, title: string): LinkedKnowledgeRef {
  return { knowledgeId: kid, knowledgeTitle: title }
}
import { getAnalyticsBundle } from '@/api/analytics'
import { getThreatActorMeta } from '@/mock/threat-actors'
import { KNOWLEDGE_DETAILS } from '@/mock/knowledge'
import type { ThreatActorListItem, ThreatActorProfile } from '@/types/threat-actor'
import type {
  AttackInfrastructure,
  AttackTool,
  ExploitedVulnerability,
  MalwareRef,
} from '@/types/knowledge'
import type { AttackTechnique } from '@/types/intel'

export function actorNameToSlug(name: string): string {
  return encodeURIComponent(name)
}

export function slugToActorName(slug: string): string {
  return decodeURIComponent(slug)
}

function dedupeTools(items: AttackTool[]): AttackTool[] {
  const map = new Map<string, AttackTool>()
  for (const t of items) {
    if (!map.has(t.name)) map.set(t.name, t)
  }
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name))
}

function dedupeInfra(items: AttackInfrastructure[]): AttackInfrastructure[] {
  const map = new Map<string, AttackInfrastructure>()
  for (const i of items) {
    map.set(`${i.type}:${i.value}`, i)
  }
  return [...map.values()].sort((a, b) => a.value.localeCompare(b.value))
}

function dedupeMalware(items: MalwareRef[]): MalwareRef[] {
  const map = new Map<string, MalwareRef>()
  for (const m of items) {
    map.set(m.name, m)
  }
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name))
}

function dedupeCves(items: ExploitedVulnerability[]): ExploitedVulnerability[] {
  const map = new Map<string, ExploitedVulnerability>()
  for (const v of items) {
    map.set(v.cveId, v)
  }
  return [...map.values()].sort((a, b) => a.cveId.localeCompare(b.cveId))
}

function dedupeTechniques(items: AttackTechnique[]): AttackTechnique[] {
  const map = new Map<string, AttackTechnique>()
  for (const t of items) {
    map.set(t.techniqueId, t)
  }
  return [...map.values()].sort((a, b) => a.techniqueId.localeCompare(b.techniqueId))
}

export function buildThreatActorList(): ThreatActorListItem[] {
  const map = new Map<string, ThreatActorListItem>()

  for (const k of KNOWLEDGE_DETAILS) {
    for (const actorName of k.threatActors) {
      let row = map.get(actorName)
      if (!row) {
        const meta = getThreatActorMeta(actorName)
        row = {
          name: actorName,
          slug: actorNameToSlug(actorName),
          targetIndustries: [],
          knowledgeCount: 0,
          knowledge: [],
          aliases: meta.aliases ?? [],
          summary: meta.summary ?? null,
        }
        map.set(actorName, row)
      }
      row.knowledge.push(knowledgeRef(k.id, k.title))
      for (const ind of k.targetIndustries) {
        if (!row.targetIndustries.includes(ind)) row.targetIndustries.push(ind)
      }
    }
  }

  const list = [...map.values()]
  for (const row of list) {
    row.knowledgeCount = row.knowledge.length
    row.targetIndustries.sort((a, b) => a.localeCompare(b, 'zh-CN'))
  }
  return list.sort((a, b) => a.name.localeCompare(b.name))
}

export function buildThreatActorProfile(actorName: string): ThreatActorProfile | null {
  const list = buildThreatActorList()
  const base = list.find((a) => a.name === actorName)
  if (!base) return null

  const linked = KNOWLEDGE_DETAILS.filter((k) => k.threatActors.includes(actorName))
  if (!linked.length) return null

  const tools: AttackTool[] = []
  const infrastructure: AttackInfrastructure[] = []
  const malware: MalwareRef[] = []
  const vulnerabilities: ExploitedVulnerability[] = []
  const techniques: AttackTechnique[] = []
  const matrices = new Set<string>()
  let intelReportCount = 0
  let firstSeen: string | null = null
  let lastSeen: string | null = null

  for (const k of linked) {
    tools.push(...k.tools)
    infrastructure.push(...k.infrastructure)
    malware.push(...k.malware)
    vulnerabilities.push(...k.vulnerabilities)
    techniques.push(...k.techniques)
    matrices.add(k.attackMatrix)
    intelReportCount += k.linkedIntel.length
    if (!firstSeen || k.updatedAt < firstSeen) firstSeen = k.updatedAt
    if (!lastSeen || k.updatedAt > lastSeen) lastSeen = k.updatedAt
  }

  const meta = getThreatActorMeta(actorName)

  return {
    ...base,
    aliases: meta.aliases ?? base.aliases,
    summary: meta.summary ?? base.summary,
    tools: dedupeTools(tools),
    infrastructure: dedupeInfra(infrastructure),
    malware: dedupeMalware(malware),
    vulnerabilities: dedupeCves(vulnerabilities),
    techniques: dedupeTechniques(techniques),
    intelReportCount,
    matrices: [...matrices],
    firstSeen,
    lastSeen,
  }
}

export function getThreatActorProfileBySlug(slug: string): ThreatActorProfile | null {
  const b = getAnalyticsBundle()
  if (b?.actorProfiles?.[slug]) {
    return b.actorProfiles[slug] as ThreatActorProfile
  }
  try {
    return buildThreatActorProfile(slugToActorName(slug))
  } catch {
    return null
  }
}
