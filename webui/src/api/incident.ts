import { apiFetch, USE_API } from '@/api/http'

export type IncidentExt = '.md' | '.markdown' | '.docx'

export interface IncidentReport {
  id: number
  title: string
  filename: string
  ext: IncidentExt | string
  mime: string
  sizeBytes: number
  sha256: string
  note: string
  reportTime: string
  uploadedAt: string
}

export interface IncidentReportListResult {
  items: IncidentReport[]
  total: number
}

export async function listIncidentReports(): Promise<IncidentReportListResult> {
  if (!USE_API) return { items: [], total: 0 }
  return apiFetch<IncidentReportListResult>('/api/v1/incident-reports')
}

export async function uploadIncidentReport(file: File, reportTime: string): Promise<IncidentReport> {
  const form = new FormData()
  form.append('file', file)
  if (reportTime) form.append('reportTime', reportTime)
  const res = await fetch('/api/v1/incident-reports', { method: 'POST', body: form })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error((err as { error?: string }).error ?? res.statusText)
  }
  return res.json() as Promise<IncidentReport>
}

export async function updateIncidentReportTime(id: number, reportTime: string): Promise<IncidentReport> {
  return apiFetch<IncidentReport>(`/api/v1/incident-reports/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ reportTime }),
  })
}

export function incidentPreviewUrl(id: number): string {
  return `/api/v1/incident-reports/${id}/preview`
}

export function incidentDownloadUrl(id: number): string {
  return `/api/v1/incident-reports/${id}/download`
}

export async function fetchIncidentMarkdown(id: number): Promise<string> {
  const res = await fetch(incidentPreviewUrl(id))
  if (!res.ok) throw new Error('预览内容获取失败')
  return res.text()
}

export async function fetchIncidentDocxHtml(id: number): Promise<string> {
  const res = await fetch(incidentPreviewUrl(id))
  if (!res.ok) throw new Error('预览内容获取失败')
  const buf = await res.arrayBuffer()
  const mammoth = await import('mammoth')
  const result = await (mammoth as unknown as {
    convertToHtml: (input: { arrayBuffer: ArrayBuffer }) => Promise<{ value: string }>
  }).convertToHtml({ arrayBuffer: buf })
  return result.value
}

export async function deleteIncidentReport(id: number): Promise<void> {
  await apiFetch<{ status: string }>(`/api/v1/incident-reports/${id}`, { method: 'DELETE' })
}
