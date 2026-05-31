import { buildThreatActorList } from '@/analytics/actors'
import { entityKeyEncode, infraEntityKey, malwareEntityKey } from '@/analytics/entity-profiles'
import { KNOWLEDGE_DETAILS } from '@/mock/knowledge'
import { INTEL_DETAILS } from '@/mock/intel'
import { malwareClassLabel, type MalwareClass } from '@/constants/malware'

export interface CountItem {
  name: string
  count: number
  extra?: string
}

export interface DimensionNavItem {
  key: string
  label: string
  icon: string
  count: number
}

export interface LinkedKnowledgeRef {
  knowledgeId: number
  knowledgeTitle: string
}

export interface CveRow {
  cveId: string
  slug: string
  description?: string
  inWild: boolean
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
}

export interface MalwareRow {
  name: string
  class: MalwareClass
  slug: string
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
}

export interface ActorRow {
  name: string
  slug: string
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
}

export interface TechniqueRow {
  techniqueId: string
  name: string
  tactic: string
  matrix: string
  knowledge: LinkedKnowledgeRef[]
}

export interface IocRow {
  type: string
  value: string
  source: 'intel' | 'infrastructure'
  knowledge: LinkedKnowledgeRef[]
}

export interface ToolRow {
  name: string
  slug: string
  category?: string
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
}

export interface InfraRow {
  type: string
  value: string
  slug: string
  role: string
  targetIndustries: string[]
  knowledgeCount: number
  knowledge: LinkedKnowledgeRef[]
}

function ref(kid: number, title: string): LinkedKnowledgeRef {
  return { knowledgeId: kid, knowledgeTitle: title }
}

function industriesFromKnowledge(refs: LinkedKnowledgeRef[]): string[] {
  const set = new Set<string>()
  for (const r of refs) {
    const k = KNOWLEDGE_DETAILS.find((x) => x.id === r.knowledgeId)
    if (k) {
      for (const ind of k.targetIndustries) set.add(ind)
    }
  }
  return [...set].sort((a, b) => a.localeCompare(b, 'zh-CN'))
}

function finalizeEntityRow<T extends { knowledge: LinkedKnowledgeRef[] }>(
  row: T,
  slug: string,
): T & { slug: string; targetIndustries: string[]; knowledgeCount: number } {
  return {
    ...row,
    slug,
    targetIndustries: industriesFromKnowledge(row.knowledge),
    knowledgeCount: row.knowledge.length,
  }
}

export function buildDimensionNav(): DimensionNavItem[] {
  const s = buildAnalyticsSnapshot()
  return [
    { key: 'overview', label: 'Dashboard', icon: '📊', count: s.knowledgeCount },
    { key: 'cve', label: '漏洞利用', icon: '🐛', count: s.cveCount },
    { key: 'malware', label: '恶意软件', icon: '🦠', count: s.malwareCount },
    { key: 'actors', label: '攻击组织', icon: '🕵️', count: s.actorCount },
    { key: 'mitre', label: 'MITRE 分布', icon: '🎯', count: s.techniqueCount },
    { key: 'tools', label: '攻击工具', icon: '🔧', count: s.toolCount },
    { key: 'infra', label: '基础设施', icon: '🏗️', count: s.infraCount },
  ]
}

export interface DashboardKnowledgeRow {
  id: number
  title: string
  attackMatrix: string
  qualityScore: number
  intelCount: number
  hasRansomware: boolean
  updatedAt: string
}

export interface DashboardOperationalStats {
  knowledgeCount: number
  intelSourceCount: number
  intelTotalCount: number
  intelSucceededCount: number
  intelProcessingCount: number
  intelFailedCount: number
  intelLinkedCount: number
  sourceLabels: string[]
}

export interface DashboardPayload {
  snapshot: ReturnType<typeof buildAnalyticsSnapshot>
  operations: DashboardOperationalStats
  knowledgeRows: DashboardKnowledgeRow[]
  ransomwareKnowledgeCount: number
  icsKnowledgeCount: number
  inWildCveCount: number
}

function buildOperationalStats(): DashboardOperationalStats {
  const sources = new Set(INTEL_DETAILS.map((i) => i.sourceName))
  return {
    knowledgeCount: KNOWLEDGE_DETAILS.length,
    intelSourceCount: sources.size,
    intelTotalCount: INTEL_DETAILS.length,
    intelSucceededCount: INTEL_DETAILS.filter((i) => i.ingestStatus === 'succeeded').length,
    intelProcessingCount: INTEL_DETAILS.filter((i) =>
      ['queued', 'parsing', 'extracting'].includes(i.ingestStatus),
    ).length,
    intelFailedCount: INTEL_DETAILS.filter((i) => i.ingestStatus === 'failed').length,
    intelLinkedCount: INTEL_DETAILS.filter((i) => i.knowledgeIds.length > 0).length,
    sourceLabels: [...sources].sort(),
  }
}

export function buildDashboardPayload(): DashboardPayload {
  const snapshot = buildAnalyticsSnapshot()
  const operations = buildOperationalStats()
  const knowledgeRows: DashboardKnowledgeRow[] = KNOWLEDGE_DETAILS.map((k) => ({
    id: k.id,
    title: k.title,
    attackMatrix: k.attackMatrix,
    qualityScore: k.qualityScore,
    intelCount: k.intelCount,
    hasRansomware: k.malware.some((m) => m.class === 'ransomware'),
    updatedAt: k.updatedAt,
  })).sort((a, b) => b.updatedAt.localeCompare(a.updatedAt))

  return {
    snapshot,
    operations,
    knowledgeRows,
    ransomwareKnowledgeCount: knowledgeRows.filter((k) => k.hasRansomware).length,
    icsKnowledgeCount: knowledgeRows.filter((k) => k.attackMatrix === 'ICS').length,
    inWildCveCount: buildCveRows().filter((c) => c.inWild).length,
  }
}

export function buildAnalyticsSnapshot() {
  const cves = buildCveRows()
  const malware = buildMalwareRows()
  const actors = buildActorRows()
  const techniques = buildTechniqueRows()
  const tools = buildToolRows()
  const infra = buildInfraRows()
  const iocs = buildIocRows()

  return {
    knowledgeCount: KNOWLEDGE_DETAILS.length,
    cveCount: cves.length,
    malwareCount: malware.length,
    actorCount: actors.length,
    techniqueCount: techniques.length,
    toolCount: tools.length,
    infraCount: infra.length,
    iocCount: iocs.length,
    industryDist: countBy(
      KNOWLEDGE_DETAILS.flatMap((k) => k.targetIndustries.map((i) => ({ name: i }))),
    ),
    matrixDist: countBy(KNOWLEDGE_DETAILS.map((k) => ({ name: k.attackMatrix }))),
    malwareClassDist: countBy(
      malware.map((m) => ({ name: malwareClassLabel(m.class) })),
    ),
    tacticDist: countBy(techniques.map((t) => ({ name: t.tactic }))),
  }
}

function countBy(items: { name: string }[]): CountItem[] {
  const map = new Map<string, number>()
  for (const item of items) {
    map.set(item.name, (map.get(item.name) ?? 0) + 1)
  }
  return [...map.entries()]
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
}

type CveRowDraft = Pick<CveRow, 'cveId' | 'description' | 'inWild' | 'knowledge'>

const rowCache: {
  cve: CveRow[] | null
  malware: MalwareRow[] | null
  tools: ToolRow[] | null
  infra: InfraRow[] | null
  techniques: TechniqueRow[] | null
  ioc: IocRow[] | null
} = {
  cve: null,
  malware: null,
  tools: null,
  infra: null,
  techniques: null,
  ioc: null,
}

export function invalidateAggregateCache(): void {
  rowCache.cve = null
  rowCache.malware = null
  rowCache.tools = null
  rowCache.infra = null
  rowCache.techniques = null
  rowCache.ioc = null
}

export function buildCveRows(): CveRow[] {
  if (rowCache.cve) return rowCache.cve
  const map = new Map<string, CveRowDraft>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const v of k.vulnerabilities) {
      const existing = map.get(v.cveId)
      if (existing) {
        existing.knowledge.push(ref(k.id, k.title))
      } else {
        map.set(v.cveId, {
          cveId: v.cveId,
          description: v.description,
          inWild: !!v.inWild,
          knowledge: [ref(k.id, k.title)],
        })
      }
    }
  }
  rowCache.cve = [...map.values()]
    .map((row) => finalizeEntityRow(row, entityKeyEncode(row.cveId)))
    .sort((a, b) => a.cveId.localeCompare(b.cveId))
  return rowCache.cve
}

type MalwareRowDraft = Pick<MalwareRow, 'name' | 'class' | 'knowledge'>

export function buildMalwareRows(): MalwareRow[] {
  if (rowCache.malware) return rowCache.malware
  const map = new Map<string, MalwareRowDraft>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const m of k.malware) {
      const key = `${m.name}|${m.class}`
      const existing = map.get(key)
      if (existing) {
        existing.knowledge.push(ref(k.id, k.title))
      } else {
        map.set(key, {
          name: m.name,
          class: m.class,
          knowledge: [ref(k.id, k.title)],
        })
      }
    }
  }
  rowCache.malware = [...map.values()]
    .map((row) => finalizeEntityRow(row, malwareEntityKey(row.name, row.class)))
    .sort((a, b) => a.name.localeCompare(b.name))
  return rowCache.malware
}

export function buildActorRows(): ActorRow[] {
  return buildThreatActorList()
}

export function buildTechniqueRows(): TechniqueRow[] {
  if (rowCache.techniques) return rowCache.techniques
  const map = new Map<string, TechniqueRow>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const t of k.techniques) {
      let row = map.get(t.techniqueId)
      if (!row) {
        row = {
          techniqueId: t.techniqueId,
          name: t.name,
          tactic: t.tactic,
          matrix: k.attackMatrix,
          knowledge: [],
        }
        map.set(t.techniqueId, row)
      }
      row.knowledge.push(ref(k.id, k.title))
    }
  }
  rowCache.techniques = [...map.values()].sort((a, b) => a.techniqueId.localeCompare(b.techniqueId))
  return rowCache.techniques
}

type ToolRowDraft = Pick<ToolRow, 'name' | 'category' | 'knowledge'>

export function buildToolRows(): ToolRow[] {
  if (rowCache.tools) return rowCache.tools
  const map = new Map<string, ToolRowDraft>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const t of k.tools) {
      const existing = map.get(t.name)
      if (existing) {
        existing.knowledge.push(ref(k.id, k.title))
      } else {
        map.set(t.name, {
          name: t.name,
          category: t.category,
          knowledge: [ref(k.id, k.title)],
        })
      }
    }
  }
  rowCache.tools = [...map.values()]
    .map((row) => finalizeEntityRow(row, entityKeyEncode(row.name)))
    .sort((a, b) => a.name.localeCompare(b.name))
  return rowCache.tools
}

type InfraRowDraft = Pick<InfraRow, 'type' | 'value' | 'role' | 'knowledge'>

export function buildInfraRows(): InfraRow[] {
  if (rowCache.infra) return rowCache.infra
  const map = new Map<string, InfraRowDraft>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const i of k.infrastructure) {
      const key = `${i.type}|${i.value}`
      const existing = map.get(key)
      if (existing) {
        existing.knowledge.push(ref(k.id, k.title))
      } else {
        map.set(key, {
          type: i.type,
          value: i.value,
          role: i.role,
          knowledge: [ref(k.id, k.title)],
        })
      }
    }
  }
  rowCache.infra = [...map.values()]
    .map((row) => finalizeEntityRow(row, infraEntityKey(row.type, row.value)))
    .sort((a, b) => a.value.localeCompare(b.value))
  return rowCache.infra
}

export function buildIocRows(): IocRow[] {
  if (rowCache.ioc) return rowCache.ioc
  const map = new Map<string, IocRow>()

  for (const intel of INTEL_DETAILS) {
    for (const ioc of intel.iocs) {
      const key = `${ioc.type}|${ioc.value}`
      const knowledgeRefs = intel.knowledgeIds.map((kid) => {
        const k = KNOWLEDGE_DETAILS.find((x) => x.id === kid)
        return ref(kid, k?.title ?? `#${kid}`)
      })
      map.set(key, {
        type: ioc.type,
        value: ioc.value,
        source: 'intel',
        knowledge: knowledgeRefs,
      })
    }
  }

  for (const k of KNOWLEDGE_DETAILS) {
    for (const i of k.infrastructure) {
      const key = `${i.type}|${i.value}`
      if (map.has(key)) continue
      map.set(key, {
        type: i.type,
        value: i.value,
        source: 'infrastructure',
        knowledge: [ref(k.id, k.title)],
      })
    }
  }

  rowCache.ioc = [...map.values()].sort((a, b) => a.type.localeCompare(b.type))
  return rowCache.ioc
}
