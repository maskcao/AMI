/** Earliest non-empty disclosed_at among linked intel (ISO strings compare lexicographically). */
export function earliestDisclosedAt(
  items: ReadonlyArray<{ disclosedAt: string | null | undefined }>,
): string | null {
  let min: string | null = null
  for (const item of items) {
    const at = item.disclosedAt?.trim()
    if (!at) continue
    if (!min || at < min) min = at
  }
  return min
}

export function formatDisclosedAt(at: string | null | undefined): string {
  if (!at) return '—'
  return at.slice(0, 16).replace('T', ' ')
}
