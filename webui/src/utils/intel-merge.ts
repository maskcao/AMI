/**
 * Pipeline merge policy: attach/update AttackKnowledge only when report titles match.
 * See docs/design.md §6.5
 */

export function normalizeIntelTitle(title: string): string {
  return title.trim().replace(/\s+/g, ' ')
}

/** Exact match after trim + whitespace collapse (case-sensitive). */
export function intelTitlesMatch(a: string, b: string): boolean {
  return normalizeIntelTitle(a) === normalizeIntelTitle(b)
}

/**
 * Resolve target knowledge for an incoming report title against prior intel rows.
 * Returns knowledge id to update, or null → pipeline should create new knowledge.
 */
export function findKnowledgeIdForMergeByTitle(
  incomingTitle: string,
  priorIntel: { title: string; knowledgeIds: number[] }[],
): number | null {
  for (const row of priorIntel) {
    if (!intelTitlesMatch(row.title, incomingTitle)) continue
    const kid = row.knowledgeIds[0]
    if (kid != null) return kid
  }
  return null
}
