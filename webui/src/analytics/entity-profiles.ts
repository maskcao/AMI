import type { LinkedKnowledgeRef } from '@/analytics/actors'
import { getAnalyticsBundle, hydrateAnalytics } from '@/api/analytics'
import { getKnowledge } from '@/api/knowledge'
import { USE_API } from '@/api/http'
import { KNOWLEDGE_DETAILS } from '@/mock/knowledge'
import type { EntityDimensionKind, EntityDimensionProfile } from '@/types/entity-profile'
import type { KnowledgeDetail } from '@/types/knowledge'
import type { MalwareClass } from '@/constants/malware'
import type {
  AttackInfrastructure,
  AttackTool,
  ExploitedVulnerability,
  MalwareRef,
} from '@/types/knowledge'
import type { AttackTechnique } from '@/types/intel'

export function entityKeyEncode(value: string): string {
  return encodeURIComponent(value)
}

export function entityKeyDecode(key: string): string {
  return decodeURIComponent(key)
}

export function malwareEntityKey(name: string, cls: MalwareClass): string {
  return entityKeyEncode(`${name}|${cls}`)
}

export function infraEntityKey(type: string, value: string): string {
  return entityKeyEncode(`${type}|${value}`)
}

function dedupeTools(items: AttackTool[]): AttackTool[] {
  const map = new Map<string, AttackTool>()
  for (const t of items) map.set(t.name, t)
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name))
}

function dedupeInfra(items: AttackInfrastructure[]): AttackInfrastructure[] {
  const map = new Map<string, AttackInfrastructure>()
  for (const i of items) map.set(`${i.type}:${i.value}`, i)
  return [...map.values()].sort((a, b) => a.value.localeCompare(b.value))
}

function dedupeMalware(items: MalwareRef[]): MalwareRef[] {
  const map = new Map<string, MalwareRef>()
  for (const m of items) map.set(m.name, m)
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name))
}

function dedupeCves(items: ExploitedVulnerability[]): ExploitedVulnerability[] {
  const map = new Map<string, ExploitedVulnerability>()
  for (const v of items) map.set(v.cveId, v)
  return [...map.values()].sort((a, b) => a.cveId.localeCompare(b.cveId))
}

function dedupeTechniques(items: AttackTechnique[]): AttackTechnique[] {
  const map = new Map<string, AttackTechnique>()
  for (const t of items) map.set(t.techniqueId, t)
  return [...map.values()].sort((a, b) => a.techniqueId.localeCompare(b.techniqueId))
}

function dedupeStrings(items: string[]): string[] {
  return [...new Set(items)].sort((a, b) => a.localeCompare(b, 'zh-CN'))
}

function collectIndustries(knowledge: KnowledgeDetail[]): string[] {
  const set = new Set<string>()
  for (const k of knowledge) {
    for (const ind of k.targetIndustries) set.add(ind)
  }
  return [...set].sort((a, b) => a.localeCompare(b, 'zh-CN'))
}

export function buildProfileFromKnowledge(
  dimension: EntityDimensionKind,
  title: string,
  subtitle: string,
  slug: string,
  linked: KnowledgeDetail[],
  extra: Partial<EntityDimensionProfile> = {},
): EntityDimensionProfile | null {
  if (!linked.length) return null

  const knowledge: LinkedKnowledgeRef[] = linked.map((k) => ({
    knowledgeId: k.id,
    knowledgeTitle: k.title,
  }))

  const tools: AttackTool[] = []
  const infrastructure: AttackInfrastructure[] = []
  const malware: MalwareRef[] = []
  const vulnerabilities: ExploitedVulnerability[] = []
  const techniques: AttackTechnique[] = []
  const threatActors: string[] = []
  const matrices = new Set<string>()
  let intelReportCount = 0

  for (const k of linked) {
    tools.push(...k.tools)
    infrastructure.push(...k.infrastructure)
    malware.push(...k.malware)
    vulnerabilities.push(...k.vulnerabilities)
    techniques.push(...k.techniques)
    threatActors.push(...k.threatActors)
    matrices.add(k.attackMatrix)
    intelReportCount += k.linkedIntel.length
  }

  return {
    dimension,
    title,
    subtitle,
    slug,
    summary: linked[0]?.summary ?? null,
    targetIndustries: collectIndustries(linked),
    knowledgeCount: knowledge.length,
    knowledge,
    threatActors: dedupeStrings(threatActors),
    techniques: dedupeTechniques(techniques),
    tools: dedupeTools(tools),
    infrastructure: dedupeInfra(infrastructure),
    malware: dedupeMalware(malware),
    vulnerabilities: dedupeCves(vulnerabilities),
    intelReportCount,
    matrices: [...matrices],
    ...extra,
  }
}

export function buildCveProfile(cveId: string): EntityDimensionProfile | null {
  const linked = KNOWLEDGE_DETAILS.filter((k) =>
    k.vulnerabilities.some((v) => v.cveId === cveId),
  )
  const vuln = linked.flatMap((k) => k.vulnerabilities).find((v) => v.cveId === cveId)
  return buildProfileFromKnowledge(
    'cve',
    cveId,
    '漏洞利用',
    entityKeyEncode(cveId),
    linked,
    { cveId, inWild: !!vuln?.inWild, summary: vuln?.description ?? linked[0]?.summary ?? null },
  )
}

export function buildMalwareProfile(name: string, cls: MalwareClass): EntityDimensionProfile | null {
  const linked = KNOWLEDGE_DETAILS.filter((k) =>
    k.malware.some((m) => m.name === name && m.class === cls),
  )
  return buildProfileFromKnowledge(
    'malware',
    name,
    '恶意软件',
    malwareEntityKey(name, cls),
    linked,
    { malwareClass: cls },
  )
}

export function buildToolProfile(name: string): EntityDimensionProfile | null {
  const linked = KNOWLEDGE_DETAILS.filter((k) => k.tools.some((t) => t.name === name))
  const tool = linked.flatMap((k) => k.tools).find((t) => t.name === name)
  return buildProfileFromKnowledge('tool', name, '攻击工具', entityKeyEncode(name), linked, {
    toolCategory: tool?.category,
  })
}

export function buildInfraProfile(type: string, value: string): EntityDimensionProfile | null {
  const linked = KNOWLEDGE_DETAILS.filter((k) =>
    k.infrastructure.some((i) => i.type === type && i.value === value),
  )
  const infra = linked.flatMap((k) => k.infrastructure).find((i) => i.type === type && i.value === value)
  return buildProfileFromKnowledge(
    'infra',
    value,
    `${type.toUpperCase()} · 基础设施`,
    infraEntityKey(type, value),
    linked,
    { infraType: type, infraRole: infra?.role },
  )
}

export function getEntityProfile(
  dimension: EntityDimensionKind,
  entityKey: string,
): EntityDimensionProfile | null {
  const decoded = entityKeyDecode(entityKey)
  switch (dimension) {
    case 'cve':
      return buildCveProfile(decoded)
    case 'malware': {
      const [name, cls] = decoded.split('|')
      if (!name || !cls) return null
      return buildMalwareProfile(name, cls as MalwareClass)
    }
    case 'tool':
      return buildToolProfile(decoded)
    case 'infra': {
      const [type, ...rest] = decoded.split('|')
      const value = rest.join('|')
      if (!type || !value) return null
      return buildInfraProfile(type, value)
    }
    default:
      return null
  }
}

export async function getEntityProfileAsync(
  dimension: EntityDimensionKind,
  entityKey: string,
): Promise<EntityDimensionProfile | null> {
  if (!USE_API) return getEntityProfile(dimension, entityKey)

  await hydrateAnalytics()
  const bundle = getAnalyticsBundle()
  if (!bundle) return null

  const decoded = entityKeyDecode(entityKey)
  let row: { knowledge: LinkedKnowledgeRef[] } | undefined
  let title = decoded
  let subtitle = ''
  let extra: Partial<EntityDimensionProfile> = {}

  switch (dimension) {
    case 'cve': {
      const r = bundle.cve.find((x) => x.slug === entityKey || x.cveId === decoded)
      if (!r) return null
      row = r
      title = r.cveId
      subtitle = '漏洞利用'
      extra = { cveId: r.cveId, inWild: r.inWild, ...(r.description ? { summary: r.description } : {}) }
      break
    }
    case 'malware': {
      const [name, cls] = decoded.split('|')
      const r = bundle.malware.find((x) => x.slug === entityKey || (x.name === name && x.class === cls))
      if (!r) return null
      row = r
      title = r.name
      subtitle = '恶意软件'
      extra = { malwareClass: r.class }
      break
    }
    case 'tool': {
      const r = bundle.tools.find((x) => x.slug === entityKey || x.name === decoded)
      if (!r) return null
      row = r
      title = r.name
      subtitle = '攻击工具'
      extra = { toolCategory: r.category }
      break
    }
    case 'infra': {
      const [type, ...rest] = decoded.split('|')
      const value = rest.join('|')
      const r = bundle.infra.find((x) => x.slug === entityKey || (x.type === type && x.value === value))
      if (!r) return null
      row = r
      title = r.value
      subtitle = `${r.type.toUpperCase()} · 基础设施`
      extra = { infraType: r.type, infraRole: r.role }
      break
    }
    default:
      return null
  }

  const linked = (
    await Promise.all(row.knowledge.map((kref) => getKnowledge(kref.knowledgeId)))
  ).filter((k): k is KnowledgeDetail => !!k)

  return buildProfileFromKnowledge(dimension, title, subtitle, entityKey, linked, extra)
}

export const DIMENSION_LIST_PATH: Record<EntityDimensionKind, string> = {
  cve: '/analytics/cve',
  malware: '/analytics/malware',
  tool: '/analytics/tools',
  infra: '/analytics/infra',
}
