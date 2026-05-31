import type { KnowledgeQualityInput } from '@/utils/quality'
import { computeKnowledgeQuality } from '@/utils/quality'
import type { KnowledgeStatus } from '@/types/knowledge'

const REVIEW_QUALITY_THRESHOLD = 70

export function computeNeedsReview(input: {
  status: KnowledgeStatus
  qualityScore?: number
  intelCount: number
  qualityInput?: KnowledgeQualityInput
}): boolean {
  const score =
    input.qualityScore ??
    (input.qualityInput ? computeKnowledgeQuality(input.qualityInput) : 0)
  if (input.status === 'draft') return true
  if (score < REVIEW_QUALITY_THRESHOLD) return true
  if (input.intelCount === 0) return true
  return false
}

export const KNOWLEDGE_STATUS_LABELS: Record<KnowledgeStatus, string> = {
  draft: '草稿',
  published: '已发布',
}
