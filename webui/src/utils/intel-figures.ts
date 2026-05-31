import type { IntelFigure, IntelFigureKind } from '@/types/intel'

export const INTEL_FIGURE_KIND_LABELS: Record<IntelFigureKind, string> = {
  'attack-flow': '攻击流程',
  'technique-map': '手法映射',
  screenshot: '截图',
  other: '其他',
}

/** Pull &lt;img&gt; tags from report HTML (demo parser). */
export function extractFiguresFromHtml(html: string, intelId: number): IntelFigure[] {
  if (!html?.trim()) return []
  const figures: IntelFigure[] = []
  const re = /<img\b[^>]*\bsrc=["']([^"']+)["'][^>]*>/gi
  let match: RegExpExecArray | null
  let index = 0
  while ((match = re.exec(html)) !== null) {
    const tag = match[0]
    const src = match[1]
    const altMatch = tag.match(/\balt=["']([^"']*)["']/i)
    const title = altMatch?.[1]?.trim() || `报告插图 ${index + 1}`
    figures.push({
      id: `html-${intelId}-${index}`,
      intelId,
      title,
      kind: 'screenshot',
      src,
      caption: title,
    })
    index += 1
  }
  return figures
}

export function mergeIntelFigures(
  explicit: IntelFigure[] | undefined,
  html: string,
  intelId: number,
): IntelFigure[] {
  const seen = new Set<string>()
  const out: IntelFigure[] = []
  for (const f of [...(explicit ?? []), ...extractFiguresFromHtml(html, intelId)]) {
    const key = `${f.intelId}:${f.src}`
    if (seen.has(key)) continue
    seen.add(key)
    out.push(f)
  }
  return out
}

export function collectFiguresForIntelIds(
  intelIds: number[],
  resolveIntel: (id: number) => { figures?: IntelFigure[] } | undefined,
): IntelFigure[] {
  const seen = new Set<string>()
  const out: IntelFigure[] = []
  for (const id of intelIds) {
    const intel = resolveIntel(id)
    if (!intel?.figures?.length) continue
    for (const f of intel.figures) {
      if (seen.has(f.id)) continue
      seen.add(f.id)
      out.push(f)
    }
  }
  return out
}

export function figuresForTechnique(
  figures: IntelFigure[],
  techniqueId: string,
): IntelFigure[] {
  return figures.filter((f) => f.techniqueIds?.includes(techniqueId))
}
