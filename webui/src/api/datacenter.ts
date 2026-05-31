import { apiFetch, USE_API } from '@/api/http'

export type RuntimeLogLevel = 'info' | 'warn' | 'error'

export interface RuntimeLogEntry {
  id: number
  createdAt: string
  level: RuntimeLogLevel
  source: string
  message: string
  detail?: string
}

export interface RuntimeLogListResult {
  items: RuntimeLogEntry[]
  total: number
  sources: string[]
}

export async function listRuntimeLogs(params?: {
  level?: string
  source?: string
  q?: string
  limit?: number
}): Promise<RuntimeLogListResult> {
  if (!USE_API) {
    return { items: [], total: 0, sources: [] }
  }
  const q = new URLSearchParams()
  if (params?.level) q.set('level', params.level)
  if (params?.source) q.set('source', params.source)
  if (params?.q) q.set('q', params.q)
  if (params?.limit) q.set('limit', String(params.limit))
  const qs = q.toString()
  return apiFetch<RuntimeLogListResult>(`/api/v1/datacenter/runtime-logs${qs ? `?${qs}` : ''}`)
}

export async function clearRuntimeLogs(): Promise<void> {
  if (!USE_API) return
  await apiFetch<{ status: string }>('/api/v1/datacenter/runtime-logs/clear', { method: 'POST' })
}
