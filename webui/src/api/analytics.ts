import { apiFetch, USE_API } from '@/api/http'
import type { DashboardPayload } from '@/analytics/aggregate'
import { emptyDashboardPayload, normalizeDashboardPayload } from '@/analytics/normalize'
import type { ActorRow, CveRow, InfraRow, IocRow, MalwareRow, TechniqueRow, ToolRow } from '@/analytics/aggregate'
import type { MalwareClass } from '@/constants/malware'
import type { DimensionNavItem } from '@/analytics/aggregate'
import type { ThreatActorProfile } from '@/types/threat-actor'

export interface AnalyticsBundle {
  dashboard: DashboardPayload
  nav: DimensionNavItem[]
  cve: CveRow[]
  malware: MalwareRow[]
  actors: ActorRow[]
  actorProfiles: Record<string, ThreatActorProfile>
  techniques: TechniqueRow[]
  tools: ToolRow[]
  infra: InfraRow[]
  ioc: IocRow[]
}

let cached: AnalyticsBundle | null = null
let hydrateError: string | null = null

function normalizeMalwareRows(rows: MalwareRow[]): MalwareRow[] {
  return rows.map((row) => {
    const raw = row as MalwareRow & { classification?: string }
    const cls = (row.class || raw.classification || '').trim() as MalwareClass
    return { ...row, class: cls }
  })
}

export function getAnalyticsHydrateError(): string | null {
  return hydrateError
}

export async function hydrateAnalytics(force = false): Promise<void> {
  if (!USE_API) return
  if (cached && !force) return
  hydrateError = null
  try {
    const raw = await apiFetch<AnalyticsBundle>('/api/v1/analytics/bundle')
    cached = {
      ...raw,
      dashboard: normalizeDashboardPayload(raw.dashboard ?? {}),
      cve: raw.cve ?? [],
      malware: normalizeMalwareRows(raw.malware ?? []),
      actors: raw.actors ?? [],
      actorProfiles: raw.actorProfiles ?? {},
      techniques: raw.techniques ?? [],
      tools: raw.tools ?? [],
      infra: raw.infra ?? [],
      ioc: raw.ioc ?? [],
      nav: raw.nav ?? [],
    }
  } catch (e) {
    cached = null
    hydrateError = e instanceof Error ? e.message : 'analytics bundle failed'
    throw e
  }
}

export async function loadDashboardPayload(): Promise<DashboardPayload> {
  if (!USE_API) {
    throw new Error('loadDashboardPayload requires VITE_USE_API=true')
  }
  await hydrateAnalytics()
  if (!cached?.dashboard) {
    return emptyDashboardPayload()
  }
  return cached.dashboard
}

export function getAnalyticsBundle(): AnalyticsBundle | null {
  return cached
}

export function invalidateAnalyticsCache(): void {
  cached = null
  hydrateError = null
}

export async function fetchActorProfileBySlug(slug: string): Promise<ThreatActorProfile | null> {
  if (!USE_API) return null
  await hydrateAnalytics()
  const b = getAnalyticsBundle()
  if (b?.actorProfiles?.[slug]) {
    return b.actorProfiles[slug]
  }
  try {
    return await apiFetch<ThreatActorProfile>(`/api/v1/analytics/actors/${encodeURIComponent(slug)}`)
  } catch {
    return null
  }
}
