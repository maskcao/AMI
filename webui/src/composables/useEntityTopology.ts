import { computed, type Ref } from 'vue'
import type { TopologyDimensionNode } from '@/components/analytics/EntityTopologyGraph.vue'
import type { EntityDimensionKind } from '@/types/entity-profile'

const PALETTE = {
  knowledge: '#3b82f6',
  industry: '#6366f1',
  actor: '#8b5cf6',
  technique: '#a855f7',
  tool: '#06b6d4',
  infra: '#f59e0b',
  malware: '#ef4444',
  cve: '#22c55e',
}

export type TopologyCenterKind = EntityDimensionKind | 'actor'

export interface TopologyProfileSlice {
  dimension?: TopologyCenterKind
  knowledgeCount: number
  targetIndustries: string[]
  threatActors: string[]
  techniques: unknown[]
  tools: unknown[]
  infrastructure: unknown[]
  malware: unknown[]
  vulnerabilities: unknown[]
}

export function useEntityTopologyNodes(profile: Ref<TopologyProfileSlice | null>) {
  return computed((): TopologyDimensionNode[] => {
    const p = profile.value
    if (!p) return []
    const nodes: TopologyDimensionNode[] = [
      { key: 'knowledge', label: '攻击知识', count: p.knowledgeCount, color: PALETTE.knowledge },
      {
        key: 'industry',
        label: '攻击行业',
        count: p.targetIndustries.length,
        color: PALETTE.industry,
      },
      { key: 'actors', label: '攻击组织', count: p.threatActors.length, color: PALETTE.actor },
      { key: 'techniques', label: '攻击手法', count: p.techniques.length, color: PALETTE.technique },
      { key: 'tools', label: '攻击工具', count: p.tools.length, color: PALETTE.tool },
      { key: 'infra', label: '基础设施', count: p.infrastructure.length, color: PALETTE.infra },
      { key: 'malware', label: '恶意软件', count: p.malware.length, color: PALETTE.malware },
      { key: 'cve', label: '漏洞利用', count: p.vulnerabilities.length, color: PALETTE.cve },
    ]
    return nodes.filter((n) => {
      if (p.dimension === 'actor' && n.key === 'actors') return false
      if (p.dimension === 'cve' && n.key === 'cve') return false
      if (p.dimension === 'malware' && n.key === 'malware') return false
      if (p.dimension === 'tool' && n.key === 'tools') return false
      if (p.dimension === 'infra' && n.key === 'infra') return false
      return true
    })
  })
}
