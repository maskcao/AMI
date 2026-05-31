<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getIntel } from '@/api/intel'
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
    await loadLinkedKnowledge(data.knowledgeIds)
    if (data.ingestStatus === 'succeeded' && data.aiExtract) {
      activeTab.value = 'extract'
    } else if (data.ingestStatus === 'failed') {
      activeTab.value = 'pipeline'
    }
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push({ name: 'intel-list' })
}

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}

function formatDisclosed(at: string | null) {
  return at ?? '未知'
}

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

watch(intelId, load)
onMounted(load)
</script>

<template>
  <AppLayout>
    <PageShell v-loading="loading">
      <template v-if="detail">
        <div class="detail-top">
          <el-button text type="primary" @click="goBack">← 返回采集队列</el-button>
          <h1 class="detail-title">{{ detail.title }}</h1>
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
            <el-descriptions-item label="来源">
              <el-link :href="detail.sourceUrl" target="_blank" type="primary">
                {{ detail.sourceName }}
              </el-link>
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
                <p class="extract-summary">{{ detail.aiExtract.summary }}</p>
                <el-descriptions :column="1" border size="small" class="extract-block">
                  <el-descriptions-item label="候选技术">
                    {{ detail.aiExtract.suggestedTechniques.join('、') || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="候选恶意软件">
                    {{ detail.aiExtract.suggestedMalware.join('、') || '—' }}
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

            <el-tab-pane label="原始信息" name="original">
              <OriginalPanel
                :html="detail.originalHtml"
                :collected-at="detail.collectedAt"
                :parse-version="detail.parseVersion"
              />
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </template>

      <el-result v-else-if="!loading" icon="warning" title="未找到该任务">
        <template #extra>
          <el-button type="primary" @click="goBack">返回列表</el-button>
        </template>
      </el-result>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.detail-top {
  margin-bottom: 16px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.detail-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  flex: 1 1 100%;
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
