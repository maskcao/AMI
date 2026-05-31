import type { IngestStatus, KnowledgeLinkAction } from '@/types/intel'

export const INGEST_STATUS_LABELS: Record<IngestStatus, string> = {
  queued: '排队中',
  parsing: '解析中',
  extracting: 'AI 抽取中',
  succeeded: '已完成',
  failed: '失败',
  archived: '已归档',
}

export const KNOWLEDGE_ACTION_LABELS: Record<KnowledgeLinkAction, string> = {
  created: '已新建知识',
  updated: '已更新知识',
  skipped: '未写入知识',
  pending: '等待写入',
}

export function ingestStatusTagType(status: IngestStatus): '' | 'success' | 'warning' | 'danger' | 'info' {
  if (status === 'succeeded') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'extracting' || status === 'parsing') return 'warning'
  if (status === 'queued') return 'info'
  return 'info'
}
