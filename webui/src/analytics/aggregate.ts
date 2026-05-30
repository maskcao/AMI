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
  description?: string
  inWild: boolean
  knowledge: LinkedKnowledgeRef[]
}

export interface MalwareRow {
  name: string
  class: MalwareClass
  knowledge: LinkedKnowledgeRef[]
}

export interface ActorRow {
  name: string
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
  category?: string
  knowledge: LinkedKnowledgeRef[]
}

export interface InfraRow {
  type: string
  value: string
  role: string
  knowledge: LinkedKnowledgeRef[]
}

function ref(kid: number, title: string): LinkedKnowledgeRef {
  return { knowledgeId: kid, knowledgeTitle: title }
}

export function buildDimensionNav(): DimensionNavItem[] {
  const s = buildAnalyticsSnapshot()
  return [
    { key: 'overview', label: 'Dashboard', icon: '📊', count: s.knowledgeCount },
    { key: 'cve', label: 'CVE 列表', icon: '🐛', count: s.cveCount },
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
  intelPublishedCount: number
  intelPendingCount: number
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
    intelPublishedCount: INTEL_DETAILS.filter((i) => i.ingestStatus === 'published').length,
    intelPendingCount: INTEL_DETAILS.filter((i) => i.ingestStatus === 'pending_review').length,
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

export function buildCveRows(): CveRow[] {
  const map = new Map<string, CveRow>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const v of k.vulnerabilities) {
      let row = map.get(v.cveId)
      if (!row) {
        row = {
          cveId: v.cveId,
          description: v.description,
          inWild: !!v.inWild,
          knowledge: [],
        }
        map.set(v.cveId, row)
      }
      row.knowledge.push(ref(k.id, k.title))
    }
  }
  return [...map.values()].sort((a, b) => a.cveId.localeCompare(b.cveId))
}

export function buildMalwareRows(): MalwareRow[] {
  const map = new Map<string, MalwareRow>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const m of k.malware) {
      const key = `${m.name}|${m.class}`
      let row = map.get(key)
      if (!row) {
        row = { name: m.name, class: m.class, knowledge: [] }
        map.set(key, row)
      }
      row.knowledge.push(ref(k.id, k.title))
    }
  }
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name))
}

export function buildActorRows(): ActorRow[] {
  const map = new Map<string, ActorRow>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const a of k.threatActors) {
      let row = map.get(a)
      if (!row) {
        row = { name: a, knowledge: [] }
        map.set(a, row)
      }
      row.knowledge.push(ref(k.id, k.title))
    }
  }
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name))
}

export function buildTechniqueRows(): TechniqueRow[] {
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
  return [...map.values()].sort((a, b) => a.techniqueId.localeCompare(b.techniqueId))
}

export function buildToolRows(): ToolRow[] {
  const map = new Map<string, ToolRow>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const t of k.tools) {
      let row = map.get(t.name)
      if (!row) {
        row = { name: t.name, category: t.category, knowledge: [] }
        map.set(t.name, row)
      }
      row.knowledge.push(ref(k.id, k.title))
    }
  }
  return [...map.values()].sort((a, b) => a.name.localeCompare(b.name))
}

export function buildInfraRows(): InfraRow[] {
  const map = new Map<string, InfraRow>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const i of k.infrastructure) {
      const key = `${i.type}|${i.value}`
      let row = map.get(key)
      if (!row) {
        row = { type: i.type, value: i.value, role: i.role, knowledge: [] }
        map.set(key, row)
      }
      row.knowledge.push(ref(k.id, k.title))
    }
  }
  return [...map.values()].sort((a, b) => a.value.localeCompare(b.value))
}

export function buildIocRows(): IocRow[] {
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

  return [...map.values()].sort((a, b) => a.type.localeCompare(b.type))
}
