<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getIntel, getIntelMarkdown, getIntelOriginal, retryIntel } from '@/api/intel'
import { techniqueLabel } from '@/constants/mitre-zh'
import { USE_API } from '@/api/http'
import { getKnowledge } from '@/api/knowledge'
import AppLayout from '@/components/AppLayout.vue'
import PageShell from '@/components/PageShell.vue'
import KnowledgeEvidenceGallery from '@/components/KnowledgeEvidenceGallery.vue'
import OriginalPanel from '@/components/OriginalPanel.vue'
import {
  INGEST_STATUS_LABELS,
  KNOWLEDGE_ACTION_LABELS,
  ingestStatusTagType,
} from '@/constants/intel-pipeline'
import { INTEL_MERGE_POLICY_SUMMARY } from '@/constants/merge-policy'
import type { IntelDetail } from '@/types/intel'
import type { KnowledgeSummary } from '@/types/knowledge'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const activeTab = ref('pipeline')
const detail = ref<IntelDetail | null>(null)
const linkedKnowledge = ref<KnowledgeSummary[]>([])
const snapshotHtml = ref('')
const markdownBody = ref('')
const htmlLoaded = ref(false)
const mdLoaded = ref(false)
const artifactLoading = ref(false)
const retrying = ref(false)

const intelId = computed(() => Number(route.params.id))

async function loadLinkedKnowledge(ids: number[]) {
  const list: KnowledgeSummary[] = []
  for (const id of ids) {
    const k = await getKnowledge(id)
    if (k) list.push(k)
  }
  linkedKnowledge.value = list
}

async function load() {
  loading.value = true
  try {
    const data = await getIntel(intelId.value)
    if (!data) {
      detail.value = null
      linkedKnowledge.value = []
      return
    }
    detail.value = data
    snapshotHtml.value = USE_API ? '' : data.originalHtml
    markdownBody.value = ''
    htmlLoaded.value = false
    mdLoaded.value = false
    await loadLinkedKnowledge(data.knowledgeIds)
    const qTab = normalizeArtifactTab(typeof route.query.tab === 'string' ? route.query.tab : '')
    if (qTab === 'html' || qTab === 'markdown' || qTab === 'extract') {
      activeTab.value = qTab
    } else if (data.hasRaw || data.hasParsed) {
      activeTab.value = data.hasRaw ? 'html' : 'markdown'
    } else if (data.ingestStatus === 'succeeded' && data.aiExtract) {
      activeTab.value = 'extract'
    } else if (data.ingestStatus === 'failed') {
      activeTab.value = 'pipeline'
    }
  } finally {
    loading.value = false
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }
  void router.push({ name: 'intel-list' })
}

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}

function formatDisclosed(at: string | null) {
  return at ?? '未知'
}

const canRetryPipeline = computed(() => {
  const s = detail.value?.ingestStatus
  if (!s) return false
  return s !== 'parsing' && s !== 'extracting'
})

const pipelineStep = computed(() => {
  const s = detail.value?.ingestStatus
  if (!s) return 0
  if (s === 'queued') return 1
  if (s === 'parsing') return 2
  if (s === 'extracting') return 3
  if (s === 'succeeded') return 4
  if (s === 'failed') return detail.value?.aiExtract ? 3 : 2
  return 0
})

/** Map legacy ?tab=local|original to html snapshot tab. */
function normalizeArtifactTab(tab: string): string {
  if (tab === 'local' || tab === 'original') return 'html'
  return tab
}

function resetArtifactContent() {
  snapshotHtml.value = ''
  markdownBody.value = ''
  htmlLoaded.value = false
  mdLoaded.value = false
}

async function loadHtmlSnapshotIfNeeded() {
  if (activeTab.value !== 'html' || !detail.value || htmlLoaded.value) return
  artifactLoading.value = true
  try {
    const d = detail.value
    if (USE_API && d.hasRaw) {
      snapshotHtml.value = (await getIntelOriginal(d.id)) ?? ''
    } else if (d.originalHtml?.trim()) {
      snapshotHtml.value = d.originalHtml
    }
    htmlLoaded.value = true
  } finally {
    artifactLoading.value = false
  }
}

async function loadMarkdownIfNeeded() {
  if (activeTab.value !== 'markdown' || !detail.value || mdLoaded.value) return
  artifactLoading.value = true
  try {
    const d = detail.value
    if (USE_API && d.hasParsed) {
      markdownBody.value = (await getIntelMarkdown(d.id)) ?? ''
    }
    mdLoaded.value = true
  } finally {
    artifactLoading.value = false
  }
}

watch(activeTab, (tab) => {
  if (tab === 'html') void loadHtmlSnapshotIfNeeded()
  if (tab === 'markdown') void loadMarkdownIfNeeded()
})

async function confirmRetry() {
  if (!detail.value) return
  try {
    await ElMessageBox.confirm(
      '将重新执行「解析 → AI 抽取 → 写入知识」流水线，不会重复采集原文。是否继续？',
      '重试流水线',
      { type: 'warning', confirmButtonText: '重试', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  retrying.value = true
  try {
    const updated = await retryIntel(detail.value.id)
    if (!updated) {
      ElMessage.error('重试失败')
      return
    }
    detail.value = {
      ...updated,
      figures: updated.figures ?? [],
      iocs: updated.iocs ?? [],
    }
    snapshotHtml.value = USE_API ? '' : updated.originalHtml
    resetArtifactContent()
    await loadLinkedKnowledge(updated.knowledgeIds)
    activeTab.value = 'pipeline'
    ElMessage.success('已重新入队，请稍候刷新状态')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '重试失败')
  } finally {
    retrying.value = false
  }
}

watch(intelId, load)
onMounted(load)
</script>

<template>
  <AppLayout>
    <PageShell>
      <div class="detail-nav">
        <el-button type="primary" link @click="goBack">← 返回采集队列</el-button>
      </div>
      <div v-loading="loading" class="detail-body">
      <template v-if="detail">
        <div class="detail-top">
          <div class="detail-top-row">
            <h1 class="detail-title ami-title-wrap">{{ detail.title }}</h1>
            <div class="detail-top-actions">
              <el-button
                v-if="canRetryPipeline"
                type="warning"
                plain
                :loading="retrying"
                @click="confirmRetry"
              >
                重试流水线
              </el-button>
            </div>
          </div>
          <el-tag :type="ingestStatusTagType(detail.ingestStatus)" size="small" effect="light">
            {{ INGEST_STATUS_LABELS[detail.ingestStatus] }}
          </el-tag>
          <el-tag size="small" type="info" effect="plain">
            {{ KNOWLEDGE_ACTION_LABELS[detail.knowledgeAction] }}
          </el-tag>
        </div>

        <el-card shadow="never" class="meta-card">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="采集源">{{ detail.collectSourceName }}</el-descriptions-item>
            <el-descriptions-item label="采集时间">{{ detail.collectedAt }}</el-descriptions-item>
            <el-descriptions-item label="披露时间">
              {{ formatDisclosed(detail.disclosedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="原始 URL" :span="2">
              <el-link
                v-if="detail.sourceUrl"
                :href="detail.sourceUrl"
                target="_blank"
                type="primary"
                class="url-link"
              >
                {{ detail.sourceUrl }}
              </el-link>
              <span v-else>—</span>
            </el-descriptions-item>
            <el-descriptions-item label="来源站点">{{ detail.sourceName || '—' }}</el-descriptions-item>
            <el-descriptions-item label="本地物证">
              <el-tag v-if="detail.hasParsed" size="small" type="success">已解析正文</el-tag>
              <el-tag v-else-if="detail.hasRaw" size="small" type="warning">仅 HTML 快照</el-tag>
              <el-tag v-else size="small" type="info">无</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="解析版本">{{ detail.parseVersion || '—' }}</el-descriptions-item>
            <el-descriptions-item label="抽取模型">{{ detail.enrichVersion || '—' }}</el-descriptions-item>
            <el-descriptions-item v-if="detail.extractConfidence != null" label="抽取置信">
              {{ Math.round(detail.extractConfidence * 100) }}%
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-alert
          v-if="detail.pipelineError"
          type="error"
          :closable="false"
          show-icon
          class="error-alert"
          :title="detail.pipelineError"
        />

        <el-card shadow="never">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="流水线" name="pipeline">
              <p class="tab-hint">
                自动流水线：采集 → 正文解析 → AI NLU 抽取 → 写入攻击知识。{{ INTEL_MERGE_POLICY_SUMMARY }}
              </p>
              <el-steps :active="pipelineStep" align-center class="pipeline-steps">
                <el-step title="采集入队" description="CollectSource" />
                <el-step title="解析" description="ParserProfile" />
                <el-step title="AI 抽取" description="LLM Enrich" />
                <el-step title="知识入库" description="AttackKnowledge" />
              </el-steps>
            </el-tab-pane>

            <el-tab-pane label="AI 抽取" name="extract">
              <template v-if="detail.aiExtract">
                <el-alert
                  v-if="detail.aiExtract.qualityHints?.length"
                  :type="detail.aiExtract.qualityOk ? 'success' : 'warning'"
                  :closable="false"
                  show-icon
                  class="quality-alert"
                  :title="detail.aiExtract.qualityOk ? '质量检查通过' : '质量检查待完善'"
                >
                  <ul class="quality-hints">
                    <li v-for="(h, i) in detail.aiExtract.qualityHints" :key="i">{{ h }}</li>
                  </ul>
                </el-alert>
                <p class="extract-summary">{{ detail.aiExtract.summary }}</p>
                <el-descriptions :column="1" border size="small" class="extract-block">
                  <el-descriptions-item label="候选 ATT&CK">
                    {{
                      detail.aiExtract.suggestedTechniques.length
                        ? detail.aiExtract.suggestedTechniques
                            .map((t) => `${t} ${techniqueLabel(t)}`)
                            .join('；')
                        : '—'
                    }}
                  </el-descriptions-item>
                  <el-descriptions-item label="候选恶意软件">
                    {{ detail.aiExtract.suggestedMalware.join('、') || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="恶意软件分类">
                    {{ detail.aiExtract.malwareClasses?.join('、') || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="攻击工具">
                    {{ detail.aiExtract.tools?.join('、') || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="攻击者基础设施">
                    {{ detail.aiExtract.infrastructure?.join('、') || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="候选 CVE">
                    {{ detail.aiExtract.suggestedCves.join('、') || '—' }}
                  </el-descriptions-item>
                </el-descriptions>
              </template>
              <el-empty v-else description="尚未完成 AI 抽取" :image-size="64" />
            </el-tab-pane>

            <el-tab-pane label="攻击知识（自动）" name="knowledge">
              <p class="tab-hint">
                同标题报告更新本条知识，不同标题则另建知识；语义重复可在攻击知识库人工合并。
              </p>
              <div v-if="linkedKnowledge.length" class="knowledge-cards">
                <el-card
                  v-for="k in linkedKnowledge"
                  :key="k.id"
                  shadow="hover"
                  class="k-card"
                  @click="goKnowledge(k.id)"
                >
                  <div class="k-card-title">
                    <el-link type="primary" @click.stop="goKnowledge(k.id)">
                      #{{ k.id }} {{ k.title }}
                    </el-link>
                  </div>
                  <div class="k-card-meta">
                    <el-tag size="small" :type="k.attackMatrix === 'ICS' ? 'warning' : 'primary'">
                      {{ k.attackMatrix }} ATT&CK
                    </el-tag>
                  </div>
                </el-card>
              </div>
              <el-empty v-else description="未完成知识写入（处理中或失败）" />
            </el-tab-pane>

            <el-tab-pane name="figures">
              <template #label>
                报告图示
                <el-badge
                  v-if="detail.figures.length"
                  :value="detail.figures.length"
                  class="tab-badge"
                />
              </template>
              <KnowledgeEvidenceGallery :figures="detail.figures" />
            </el-tab-pane>

            <el-tab-pane label="IOC" name="ioc">
              <p class="tab-hint">AI/规则从原文抽取的指标，挂载在情报证据层。</p>
              <el-table v-if="detail.iocs.length" :data="detail.iocs" size="small" stripe>
                <el-table-column prop="type" label="类型" width="100" />
                <el-table-column prop="value" label="值" min-width="280" />
                <el-table-column prop="confidence" label="置信度" width="100" />
              </el-table>
              <el-empty v-else description="暂无 IOC" :image-size="64" />
            </el-tab-pane>

            <el-tab-pane label="HTML 快照" name="html">
              <p class="tab-hint">
                采集落盘的 HTML（<code>html-v2</code> 含已下载的 css/js/图片）；链接失效时仍可离线查看。
              </p>
              <div v-loading="artifactLoading">
                <OriginalPanel
                  v-if="snapshotHtml"
                  :html="snapshotHtml"
                  :collected-at="detail.collectedAt"
                  :parse-version="detail.parseVersion"
                />
                <el-empty
                  v-else-if="!artifactLoading"
                  description="暂无 HTML 快照；请重试流水线"
                  :image-size="64"
                />
              </div>
            </el-tab-pane>

            <el-tab-pane label="Markdown" name="markdown">
              <p class="tab-hint">
                由 HTML 转换的正文（<code>body.md</code>），亦为 AI 抽取阶段送入 LLM 的输入。
              </p>
              <div v-loading="artifactLoading">
                <pre v-if="markdownBody" class="artifact-md">{{ markdownBody }}</pre>
                <el-empty
                  v-else-if="!artifactLoading"
                  description="暂无 Markdown；请先完成解析或重试流水线"
                  :image-size="64"
                />
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </template>

      <el-result v-else-if="!loading" icon="warning" title="未找到该任务">
        <template #extra>
          <el-button type="primary" @click="goBack">返回列表</el-button>
        </template>
      </el-result>
      </div>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.detail-nav {
  margin-bottom: 12px;
}
.detail-body {
  min-height: 120px;
}
.quality-alert {
  margin-bottom: 12px;
}
.quality-hints {
  margin: 4px 0 0;
  padding-left: 18px;
  font-size: 13px;
}
.detail-top {
  margin-bottom: 16px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.detail-top-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}
.detail-top-actions {
  display: flex;
  gap: 8px;
}
.detail-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  flex: 1;
  min-width: 0;
}
.url-link {
  word-break: break-all;
}
.artifact-md {
  margin: 0;
  padding: 12px;
  max-height: 72vh;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  background: #f8fafc;
  border: 1px solid #ebeef5;
  border-radius: 8px;
}
.meta-card {
  margin-bottom: 16px;
}
.error-alert {
  margin-bottom: 16px;
}
.tab-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #909399;
}
.tab-badge {
  margin-left: 6px;
  vertical-align: middle;
}
.pipeline-steps {
  margin-top: 16px;
}
.extract-summary {
  margin: 0 0 12px;
  font-size: 14px;
  line-height: 1.6;
}
.extract-block {
  margin-top: 8px;
}
.knowledge-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.k-card {
  cursor: pointer;
}
.k-card-title {
  font-weight: 500;
  margin-bottom: 8px;
}
.k-card-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}
</style>
