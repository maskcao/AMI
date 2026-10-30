export const API_BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? ''
export const USE_API = import.meta.env.VITE_USE_API === 'true' || import.meta.env.VITE_USE_API === '1'

/** Origin for browser-visible API URLs (VITE_API_BASE, or current page origin for Vite /api proxy). */
export function apiPublicOrigin(): string {
  const configured = API_BASE.replace(/\/$/, '')
  if (configured) return configured
  if (typeof window !== 'undefined') return window.location.origin
  return ''
}

export function apiUrl(path: string): string {
  const base = apiPublicOrigin()
  return `${base}${path}`
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(apiUrl(path), {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...init?.headers,
    },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error((err as { error?: string }).error ?? res.statusText)
  }
  if (res.status === 204) {
    return undefined as T
  }
  return res.json() as Promise<T>
}
