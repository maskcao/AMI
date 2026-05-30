import { computed, ref } from 'vue'

export interface PagedFilterOptions<T> {
  pageSize?: number
  matchQuery?: (row: T, q: string) => boolean
  matchFacet?: (row: T, facet: string) => boolean
  matchExtras?: (row: T, extras: Record<string, string>) => boolean
}

export function usePagedFilter<T>(source: () => T[], options: PagedFilterOptions<T> = {}) {
  const query = ref('')
  const facet = ref('')
  const extras = ref<Record<string, string>>({})
  const page = ref(1)
  const pageSize = ref(options.pageSize ?? 10)

  const filtered = computed(() => {
    let rows = source()
    const f = facet.value
    if (f && options.matchFacet) {
      rows = rows.filter((r) => options.matchFacet!(r, f))
    }
    const ex = extras.value
    if (options.matchExtras && Object.keys(ex).length > 0) {
      rows = rows.filter((r) => options.matchExtras!(r, ex))
    }
    const q = query.value.trim().toLowerCase()
    if (q && options.matchQuery) {
      rows = rows.filter((r) => options.matchQuery!(r, q))
    }
    return rows
  })

  const total = computed(() => filtered.value.length)

  const pageData = computed(() => {
    const start = (page.value - 1) * pageSize.value
    return filtered.value.slice(start, start + pageSize.value)
  })

  function applyFilters(next: { query?: string; facet?: string; extras?: Record<string, string> }) {
    if (next.query !== undefined) query.value = next.query
    if (next.facet !== undefined) facet.value = next.facet
    if (next.extras !== undefined) extras.value = next.extras
    page.value = 1
  }

  function resetFilters() {
    query.value = ''
    facet.value = ''
    extras.value = {}
    page.value = 1
  }

  function onPageChange(p: number) {
    page.value = p
  }

  function onSizeChange(size: number) {
    pageSize.value = size
    page.value = 1
  }

  return {
    query,
    facet,
    extras,
    page,
    pageSize,
    filtered,
    total,
    pageData,
    applyFilters,
    resetFilters,
    onPageChange,
    onSizeChange,
  }
}

export function includesQuery(q: string, ...parts: (string | undefined | null)[]): boolean {
  return parts.some((p) => p?.toLowerCase().includes(q))
}
