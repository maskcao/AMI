<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getIntel } from '@/api/intel'
import { getKnowledge } from '@/api/knowledge'
import AppLayout from '@/components/AppLayout.vue'
import PageShell from '@/components/PageShell.vue'
import OriginalPanel from '@/components/OriginalPanel.vue'
import QualityScore from '@/components/QualityScore.vue'
import type { IntelDetail } from '@/types/intel'
import type { KnowledgeSummary } from '@/types/knowledge'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const activeTab = ref('original')
const detail = ref<IntelDetail | null>(null)
const linkedKnowledge = ref<KnowledgeSummary[]>([])

const intelId = computed(() => Number(route.params.id))

async function loadLinkedKnowledge(ids: number[]) {
  const list: KnowledgeSummary[] = []
  for (const id of ids) {
    const k = await getKnowledge(id)
    if (k) {
      list.push({
        id: k.id,
        title: k.title,
        targetIndustries: k.targetIndustries,
        attackMatrix: k.attackMatrix,
        malware: k.malware,
        threatActors: k.threatActors,
        qualityScore: k.qualityScore,
        intelCount: k.intelCount,
        updatedAt: k.updatedAt,
        dimensions: k.dimensions,
      })
    }
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

function statusLabel(s: string) {
  const map: Record<string, string> = {
    published: '已发布',
    pending_review: '待审',
    draft: '草稿',
    parsing: '解析中',
    archived: '归档',
  }
  return map[s] ?? s
}

watch(intelId, load)
onMounted(load)
</script>

<template>
  <AppLayout>
    <PageShell v-loading="loading">
      <template v-if="detail">
        <div class="detail-top">
          <el-button text type="primary" @click="goBack">← 返回情报列表</el-button>
          <h1 class="detail-title">{{ detail.title }}</h1>
          <el-tag size="small">{{ statusLabel(detail.ingestStatus) }}</el-tag>
        </div>

        <el-card shadow="never" class="meta-card">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="披露时间">
              {{ formatDisclosed(detail.disclosedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="来源">
              <el-link :href="detail.sourceUrl" target="_blank" type="primary">
                {{ detail.sourceName }}
              </el-link>
            </el-descriptions-item>
            <el-descriptions-item label="原始链接" :span="2">
              <el-link :href="detail.sourceUrl" target="_blank" type="primary">
                {{ detail.sourceUrl }}
              </el-link>
            </el-descriptions-item>
            <el-descriptions-item label="情报质量">
              <QualityScore :score="detail.qualityScore" />
              <ul v-if="detail.qualityHints.length" class="quality-hints">
                <li v-for="hint in detail.qualityHints" :key="hint">{{ hint }}</li>
              </ul>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="关联攻击知识" name="knowledge">
              <p class="tab-hint">
                TTP、行业、知识概述在攻击知识中维护；本则情报作为支撑证据挂载其下。
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
                    <el-tag
                      size="small"
                      :type="k.attackMatrix === 'ICS' ? 'warning' : 'primary'"
                    >
                      {{ k.attackMatrix }} ATT&CK
                    </el-tag>
                    <span v-if="k.targetIndustries.length">
                      行业：{{ k.targetIndustries.join('、') }}
                    </span>
                  </div>
                  <div class="k-card-dim">
                    手法 {{ k.dimensions.techniques }} · 工具 {{ k.dimensions.tools }} · 设施
                    {{ k.dimensions.infrastructure }} · 漏洞 {{ k.dimensions.vulnerabilities }}
                  </div>
                </el-card>
              </div>
              <el-empty v-else description="尚未关联攻击知识，待分析师提炼挂载" />
            </el-tab-pane>

            <el-tab-pane label="IOC（附属）" name="ioc">
              <p class="tab-hint">指标从情报原文抽取，用于佐证；非平台检索主入口。</p>
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

      <el-result v-else-if="!loading" icon="warning" title="未找到该情报">
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
.tab-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #909399;
}
.quality-hints {
  margin: 8px 0 0;
  padding-left: 18px;
  font-size: 12px;
  color: #e6a23c;
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
  color: #606266;
}
.k-card-dim {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
