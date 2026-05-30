import type {
  CollectSourceSummary,
  ModelProviderSummary,
  ParserProfileSummary,
} from '@/types/settings'

export const COLLECT_TYPE_LABELS: Record<string, string> = {
  scheduled_url: '定时 URL 监测',
  wechat_feed: '微信文章采集',
  rss: 'RSS / Atom',
  manual_upload: '手工上传',
  webhook: 'Webhook 推送',
  stix_bundle: 'STIX 文件包',
  email_imap: '邮箱 IMAP',
  api_pull: 'API 拉取',
}

export const MOCK_COLLECT_SOURCES: CollectSourceSummary[] = [
  {
    id: 'cs-1',
    name: '360 Netlab 博客',
    type: 'scheduled_url',
    typeLabel: COLLECT_TYPE_LABELS.scheduled_url,
    enabled: true,
    scheduleCron: '0 */6 * * *',
    scheduleLabel: '每 6 小时',
    sourceLabel: '360',
    lastRunAt: '2026-05-30T08:00:00Z',
    lastStatus: 'success',
  },
  {
    id: 'cs-2',
    name: 'CERT 通告 RSS',
    type: 'rss',
    typeLabel: COLLECT_TYPE_LABELS.rss,
    enabled: true,
    scheduleCron: '0 8 * * *',
    scheduleLabel: '每天 08:00',
    sourceLabel: 'CERT',
    lastRunAt: '2026-05-30T08:05:00Z',
    lastStatus: 'success',
  },
  {
    id: 'cs-3',
    name: '示例公众号历史页',
    type: 'wechat_feed',
    typeLabel: COLLECT_TYPE_LABELS.wechat_feed,
    enabled: false,
    scheduleCron: '0 9 * * *',
    scheduleLabel: '每天 09:00',
    sourceLabel: 'WeChat',
    lastRunAt: '2026-05-29T09:00:00Z',
    lastStatus: 'partial',
  },
]

export const MOCK_MODEL_PROVIDERS: ModelProviderSummary[] = [
  {
    id: 'mp-1',
    name: '内部 LLM 网关',
    kind: 'llm',
    modelId: 'qwen2.5-72b-instruct',
    baseUrl: 'https://llm.internal.example/v1',
    enabled: true,
    defaultFor: 'parse, enrich',
  },
]

export const MOCK_PARSER_PROFILES: ParserProfileSummary[] = [
  {
    id: 'pp-1',
    name: 'generic_html_v1',
    description: '通用安全博客 HTML 正文',
    usedBySources: 2,
  },
  {
    id: 'pp-2',
    name: 'wechat_article_html_v1',
    description: '微信公众号文章排版',
    usedBySources: 1,
  },
]
