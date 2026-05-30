import type { IntelDetail } from '@/types/intel'
import { buildIntelQualityHints, computeIntelQuality } from '@/utils/quality'

interface RawIntel {
  id: number
  title: string
  disclosedAt: string | null
  sourceName: string
  sourceUrl: string
  knowledgeIds: number[]
  ingestStatus: IntelDetail['ingestStatus']
  iocs: IntelDetail['iocs']
  originalHtml: string
  collectedAt: string
  parseVersion: string
}

const RAW: RawIntel[] = [
  {
    id: 4203,
    title:
      'Ten families of malicious samples are spreading using the Log4j2 vulnerability Now',
    disclosedAt: null,
    sourceName: '360',
    sourceUrl: 'https://blog.netlab.360.com/log4j2-spread/',
    knowledgeIds: [1001],
    ingestStatus: 'published',
    iocs: [],
    originalHtml:
      '<p>Multiple malware families are exploiting CVE-2021-44228 (Log4Shell) in the wild.</p><p>Ten distinct families observed in sampled telemetry.</p>',
    collectedAt: '2026-05-28T08:00:00Z',
    parseVersion: 'demo-1',
  },
  {
    id: 4205,
    title: 'CERT advisory: active Log4Shell exploitation in critical sectors',
    disclosedAt: '2026-05-18',
    sourceName: 'CERT',
    sourceUrl: 'https://cert.example.invalid/advisory/log4shell-2026',
    knowledgeIds: [1001],
    ingestStatus: 'published',
    iocs: [{ type: 'domain', value: 'log4j-c2.example.invalid', confidence: 'low' }],
    originalHtml: '<p>National CERT confirms ongoing Log4Shell exploitation.</p>',
    collectedAt: '2026-05-19T08:00:00Z',
    parseVersion: 'demo-1',
  },
  {
    id: 4234,
    title:
      "Threat Alert: New update from Sysrv-hello, now infecting victims' webpages to push malicious exe to end users",
    disclosedAt: null,
    sourceName: '360',
    sourceUrl: 'https://blog.netlab.360.com/sysrv-hello-update/',
    knowledgeIds: [1002],
    ingestStatus: 'published',
    iocs: [
      {
        type: 'domain',
        value: 'malicious-cdn.example.invalid',
        confidence: 'medium',
      },
    ],
    originalHtml: '<p>Sysrv-hello campaign analysis and IOC list.</p>',
    collectedAt: '2026-05-29T10:00:00Z',
    parseVersion: 'demo-1',
  },
  {
    id: 4301,
    title: 'Mirai variant targets energy sector ICS exposed to internet',
    disclosedAt: '2026-05-20',
    sourceName: '360',
    sourceUrl: 'https://netlab.360.com/mirai-ics-energy/',
    knowledgeIds: [1003],
    ingestStatus: 'published',
    iocs: [
      { type: 'ip', value: '198.51.100.10', confidence: 'high' },
      {
        type: 'hash',
        value: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
        confidence: 'low',
      },
    ],
    originalHtml:
      '<h2>Energy sector targeting</h2><p>Mirai variant observed against ICS-exposed assets.</p>',
    collectedAt: '2026-05-25T12:00:00Z',
    parseVersion: 'demo-1',
  },
  {
    id: 4320,
    title: 'Ransomware advisory: LockBit campaigns against manufacturing sector',
    disclosedAt: '2026-05-22',
    sourceName: '360',
    sourceUrl: 'https://blog.netlab.360.com/lockbit-manufacturing/',
    knowledgeIds: [1004],
    ingestStatus: 'published',
    iocs: [
      { type: 'hash', value: 'a1b2c3d4e5f6789012345678901234567890abcd', confidence: 'medium' },
    ],
    originalHtml:
      '<p>LockBit 3.0 observed in manufacturing targets with double extortion.</p>',
    collectedAt: '2026-05-23T09:00:00Z',
    parseVersion: 'demo-1',
  },
  {
    id: 4310,
    title: 'Weekly threat feed digest — unlinked draft sample',
    disclosedAt: null,
    sourceName: '360',
    sourceUrl: 'https://blog.netlab.360.com/weekly-digest/',
    knowledgeIds: [],
    ingestStatus: 'pending_review',
    iocs: [],
    originalHtml: '<p>Digest awaiting analyst linking to attack knowledge.</p>',
    collectedAt: '2026-05-30T06:00:00Z',
    parseVersion: 'demo-1',
  },
]

function toDetail(raw: RawIntel): IntelDetail {
  const input = {
    disclosedAt: raw.disclosedAt,
    originalHtml: raw.originalHtml,
    iocCount: raw.iocs.length,
    knowledgeIds: raw.knowledgeIds,
  }
  return {
    id: raw.id,
    title: raw.title,
    disclosedAt: raw.disclosedAt,
    sourceName: raw.sourceName,
    sourceUrl: raw.sourceUrl,
    knowledgeIds: raw.knowledgeIds,
    ingestStatus: raw.ingestStatus,
    iocCount: raw.iocs.length,
    iocs: raw.iocs,
    originalHtml: raw.originalHtml,
    collectedAt: raw.collectedAt,
    parseVersion: raw.parseVersion,
    qualityScore: computeIntelQuality(input),
    qualityHints: buildIntelQualityHints(input),
  }
}

export const INTEL_DETAILS: IntelDetail[] = RAW.map(toDetail)

export function getIntelById(id: number): IntelDetail | undefined {
  return INTEL_DETAILS.find((item) => item.id === id)
}
