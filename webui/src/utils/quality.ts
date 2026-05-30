import type { KnowledgeDetail } from '@/types/knowledge'
import type { AttackTechnique } from '@/types/intel'

export interface KnowledgeQualityInput {
  targetIndustries: string[]
  summary: string | null
  techniques: AttackTechnique[]
  tools: { name: string }[]
  infrastructure: { value: string }[]
  vulnerabilities: { cveId: string }[]
  malware: { name: string }[]
  intelCount: number
}

export function computeKnowledgeQuality(input: KnowledgeQualityInput): number {
  let score = 0
  if (input.targetIndustries.length > 0) score += 15
  if (input.summary && input.summary.trim().length > 10) score += 15
  if (input.techniques.length > 0) score += 20
  if (input.tools.length > 0) score += 10
  if (input.infrastructure.length > 0) score += 10
  if (input.vulnerabilities.length > 0) score += 10
  if (input.malware.length > 0) score += 10
  if (input.intelCount > 0) score += 5
  if (input.intelCount > 1) score += 5
  return Math.min(100, score)
}

export function buildKnowledgeQualityHints(input: KnowledgeQualityInput): string[] {
  const hints: string[] = []
  if (input.targetIndustries.length === 0) hints.push('未填写攻击行业')
  if (!input.summary?.trim()) hints.push('缺少知识概述')
  if (input.techniques.length === 0) hints.push('未标注攻击手法 (ATT&CK)')
  if (input.tools.length === 0) hints.push('未维护攻击工具')
  if (input.infrastructure.length === 0) hints.push('未维护攻击者基础设施')
  if (input.vulnerabilities.length === 0) hints.push('未关联利用漏洞')
  if (input.malware.length === 0) hints.push('未标注恶意软件/家族')
  if (input.intelCount === 0) hints.push('未挂载支撑情报')
  return hints
}

export function knowledgeQualityInput(d: KnowledgeDetail): KnowledgeQualityInput {
  return {
    targetIndustries: d.targetIndustries,
    summary: d.summary,
    techniques: d.techniques,
    tools: d.tools,
    infrastructure: d.infrastructure,
    vulnerabilities: d.vulnerabilities,
    malware: d.malware,
    intelCount: d.intelCount,
  }
}

export interface IntelQualityInput {
  disclosedAt: string | null
  originalHtml: string
  iocCount: number
  knowledgeIds: number[]
}

export function computeIntelQuality(input: IntelQualityInput): number {
  let score = 0
  if (input.disclosedAt) score += 25
  if (input.originalHtml?.trim().length > 20) score += 35
  if (input.iocCount > 0) score += 20
  if (input.knowledgeIds.length > 0) score += 20
  return Math.min(100, score)
}

export function buildIntelQualityHints(input: IntelQualityInput): string[] {
  const hints: string[] = []
  if (!input.disclosedAt) hints.push('未填写披露时间')
  if (!input.originalHtml?.trim()) hints.push('缺少原文内容')
  if (input.knowledgeIds.length === 0) hints.push('未关联攻击知识')
  return hints
}
