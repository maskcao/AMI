<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getKnowledge } from '@/api/knowledge'
import AppLayout from '@/components/AppLayout.vue'
import PageShell from '@/components/PageShell.vue'
import KnowledgeDimensionPanels from '@/components/KnowledgeDimensionPanels.vue'
import KnowledgeRelationOverview from '@/components/KnowledgeRelationOverview.vue'
import LinkedIntelList from '@/components/LinkedIntelList.vue'
import QualityScore from '@/components/QualityScore.vue'
import { ALL_INDUSTRIES, resolveAttackMatrix } from '@/constants/industry'
import type { KnowledgeDetail } from '@/types/knowledge'
import {
  buildKnowledgeQualityHints,
  computeKnowledgeQuality,
  knowledgeQualityInput,
} from '@/utils/quality'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const activeTab = ref('dimensions')
const detail = ref<KnowledgeDetail | null>(null)
const editIndustries = ref<string[]>([])

const knowledgeId = computed(() => Number(route.params.id))
const attackMatrix = computed(() => resolveAttackMatrix(editIndustries.value))

const liveQualityScore = computed(() => {
  if (!detail.value) return 0
  const base = knowledgeQualityInput({
    ...detail.value,
    targetIndustries: editIndustries.value,
  })
  return computeKnowledgeQuality(base)
})

const liveQualityHints = computed(() => {
  if (!detail.value) return []
  return buildKnowledgeQualityHints(
    knowledgeQualityInput({
      ...detail.value,
      targetIndustries: editIndustries.value,
    }),
  )
})

async function load() {
  loading.value = true
  try {
    const data = await getKnowledge(knowledgeId.value)
    if (!data) {
      detail.value = null
      return
    }
    detail.value = data
    editIndustries.value = [...data.targetIndustries]
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push({ name: 'knowledge-list' })
}

watch(knowledgeId, load)
onMounted(load)
</script>

<template>
  <AppLayout>
    <PageShell v-loading="loading">
      <template v-if="detail">
        <div class="detail-top">
          <el-button text type="primary" @click="goBack">← 返回知识库</el-button>
          <h1 class="detail-title">{{ detail.title }}</h1>
          <p class="detail-meta">知识 ID {{ detail.id }} · 更新 {{ detail.updatedAt }}</p>
        </div>

        <el-card shadow="never" class="meta-card">
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="攻击行业">
              <el-select
                v-model="editIndustries"
                multiple
                collapse-tags
                placeholder="选择行业"
                style="width: 100%; max-width: 360px"
              >
                <el-option v-for="ind in ALL_INDUSTRIES" :key="ind" :label="ind" :value="ind" />
              </el-select>
            </el-descriptions-item>
            <el-descriptions-item label="ATT&CK 矩阵">
              <el-tag :type="attackMatrix === 'ICS' ? 'warning' : 'primary'">
                {{ attackMatrix }} ATT&CK
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="威胁主体">
              <template v-if="detail.threatActors.length">
                <el-tag v-for="a in detail.threatActors" :key="a" size="small" class="actor-tag">
                  {{ a }}
                </el-tag>
              </template>
              <span v-else class="muted">-</span>
            </el-descriptions-item>
            <el-descriptions-item label="知识维度" :span="2">
              手法 {{ detail.dimensions.techniques }} · 工具 {{ detail.dimensions.tools }} ·
              基础设施 {{ detail.dimensions.infrastructure }} · 漏洞
              {{ detail.dimensions.vulnerabilities }} · 恶意软件 {{ detail.dimensions.malware }}
            </el-descriptions-item>
            <el-descriptions-item label="质量评分">
              <QualityScore :score="liveQualityScore" />
              <ul v-if="liveQualityHints.length" class="quality-hints">
                <li v-for="hint in liveQualityHints" :key="hint">{{ hint }}</li>
              </ul>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never" class="summary-card">
          <h3>知识概述</h3>
          <p v-if="detail.summary" class="summary-text">{{ detail.summary }}</p>
          <el-empty v-else description="暂无概述" :image-size="64" />
        </el-card>

        <el-card shadow="never">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="攻击维度" name="dimensions">
              <KnowledgeDimensionPanels :detail="detail" :attack-matrix="attackMatrix" />
            </el-tab-pane>
            <el-tab-pane label="关联视图" name="relation">
              <KnowledgeRelationOverview :detail="detail" />
            </el-tab-pane>
            <el-tab-pane label="支撑情报" name="intel">
              <p class="tab-hint">以下情报为本条攻击知识的依据与溯源（证据层）。</p>
              <LinkedIntelList :items="detail.linkedIntel" />
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </template>

      <el-result v-else-if="!loading" icon="warning" title="未找到该攻击知识">
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
}
.detail-title {
  margin: 8px 0 0;
  font-size: 20px;
  font-weight: 600;
}
.detail-meta {
  margin: 4px 0 0;
  font-size: 12px;
  color: #909399;
}
.meta-card,
.summary-card {
  margin-bottom: 16px;
}
.summary-card h3 {
  margin: 0 0 8px;
  font-size: 15px;
  font-weight: 600;
}
.summary-text {
  margin: 0;
  line-height: 1.7;
  color: #606266;
}
.quality-hints {
  margin: 8px 0 0;
  padding-left: 18px;
  font-size: 12px;
  color: #e6a23c;
}
.tab-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #909399;
}
.actor-tag {
  margin-right: 6px;
}
.muted {
  color: #c0c4cc;
}
</style>
