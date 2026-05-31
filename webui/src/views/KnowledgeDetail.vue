<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getKnowledge, listKnowledge, mergeKnowledge, publishKnowledge, updateKnowledge } from '@/api/knowledge'
import AppLayout from '@/components/AppLayout.vue'
import KnowledgeDimensionPanels from '@/components/KnowledgeDimensionPanels.vue'
import KnowledgeEvidenceGallery from '@/components/KnowledgeEvidenceGallery.vue'
import KnowledgePipelineCard from '@/components/KnowledgePipelineCard.vue'
import KnowledgeRelationOverview from '@/components/KnowledgeRelationOverview.vue'
import LinkedIntelList from '@/components/LinkedIntelList.vue'
import PageShell from '@/components/PageShell.vue'
import QualityScore from '@/components/QualityScore.vue'
import { ALL_INDUSTRIES, resolveAttackMatrix } from '@/constants/industry'
import type { KnowledgeDetail, KnowledgeSummary } from '@/types/knowledge'
import { KNOWLEDGE_STATUS_LABELS } from '@/utils/knowledge-review'
import {
  buildKnowledgeQualityHints,
  computeKnowledgeQuality,
  knowledgeQualityInput,
} from '@/utils/quality'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('dimensions')
const detail = ref<KnowledgeDetail | null>(null)
const editIndustries = ref<string[]>([])
const editSummary = ref('')
const mergeDialogVisible = ref(false)
const mergeTargetId = ref<number | null>(null)
const mergeCandidates = ref<KnowledgeSummary[]>([])

const knowledgeId = computed(() => Number(route.params.id))
const attackMatrix = computed(() => resolveAttackMatrix(editIndustries.value))

const liveQualityScore = computed(() => {
  if (!detail.value) return 0
  const base = knowledgeQualityInput({
    ...detail.value,
    targetIndustries: editIndustries.value,
    summary: editSummary.value || null,
  })
  return computeKnowledgeQuality(base)
})

const liveQualityHints = computed(() => {
  if (!detail.value) return []
  return buildKnowledgeQualityHints(
    knowledgeQualityInput({
      ...detail.value,
      targetIndustries: editIndustries.value,
      summary: editSummary.value || null,
    }),
  )
})

const isDirty = computed(() => {
  if (!detail.value) return false
  const industriesChanged =
    editIndustries.value.length !== detail.value.targetIndustries.length ||
    editIndustries.value.some((i, idx) => i !== detail.value!.targetIndustries[idx])
  const summaryChanged = (editSummary.value || '') !== (detail.value.summary || '')
  return industriesChanged || summaryChanged
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
    editSummary.value = data.summary ?? ''
    activeTab.value = data.evidenceFigures.length ? 'figures' : 'dimensions'
  } finally {
    loading.value = false
  }
}

async function saveEdits() {
  if (!detail.value || !isDirty.value) return
  saving.value = true
  try {
    const updated = await updateKnowledge(detail.value.id, {
      summary: editSummary.value.trim() || null,
      targetIndustries: [...editIndustries.value],
    })
    if (updated) {
      detail.value = updated
      editIndustries.value = [...updated.targetIndustries]
      editSummary.value = updated.summary ?? ''
      ElMessage.success('已保存修改')
    }
  } finally {
    saving.value = false
  }
}

async function publish() {
  if (!detail.value || detail.value.status === 'published') return
  await ElMessageBox.confirm('确认发布该攻击知识？发布后仍可在质量不足时标记为待复核。', '发布知识', {
    type: 'info',
  })
  saving.value = true
  try {
    const updated = await publishKnowledge(detail.value.id)
    if (updated) {
      detail.value = updated
      ElMessage.success('已发布')
    }
  } finally {
    saving.value = false
  }
}

async function openMergeDialog() {
  const result = await listKnowledge({ sortOrder: 'desc' })
  mergeCandidates.value = result.items.filter((k) => k.id !== knowledgeId.value)
  mergeTargetId.value = mergeCandidates.value[0]?.id ?? null
  mergeDialogVisible.value = true
}

async function confirmMerge() {
  if (!detail.value || mergeTargetId.value == null) return
  await ElMessageBox.confirm(
    `将把当前知识 #${detail.value.id} 的支撑情报合并到 #${mergeTargetId.value}，并删除当前条目（Demo）。`,
    '合并知识',
    { type: 'warning' },
  )
  saving.value = true
  try {
    const ok = await mergeKnowledge({
      sourceId: detail.value.id,
      targetId: mergeTargetId.value,
    })
    if (ok) {
      ElMessage.success('合并完成')
      mergeDialogVisible.value = false
      router.push({ name: 'knowledge-detail', params: { id: String(mergeTargetId.value) } })
    } else {
      ElMessage.error('合并失败')
    }
  } finally {
    saving.value = false
  }
}

function scrollToSummary() {
  document.getElementById('knowledge-summary-edit')?.scrollIntoView({ behavior: 'smooth' })
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
          <div class="detail-top-row">
            <el-button text type="primary" @click="goBack">← 返回知识库</el-button>
            <div class="detail-actions">
              <el-button
                v-if="detail.status === 'draft'"
                type="success"
                :loading="saving"
                @click="publish"
              >
                发布
              </el-button>
              <el-button :loading="saving" @click="openMergeDialog">合并到其他知识</el-button>
              <el-button type="primary" :disabled="!isDirty" :loading="saving" @click="saveEdits">
                保存修改
              </el-button>
            </div>
          </div>
          <h1 class="detail-title ami-title-wrap">{{ detail.title }}</h1>
          <p class="detail-meta">
            知识 ID {{ detail.id }}
            ·
            <el-tag :type="detail.status === 'published' ? 'success' : 'info'" size="small">
              {{ KNOWLEDGE_STATUS_LABELS[detail.status] }}
            </el-tag>
            <el-tag v-if="detail.needsReview" type="warning" size="small" class="review-tag">
              待复核
            </el-tag>
            · 更新 {{ detail.updatedAt.replace('T', ' ').slice(0, 16) }}
          </p>
        </div>

        <KnowledgePipelineCard :pipeline="detail.pipeline" />

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
                <li v-for="hint in liveQualityHints" :key="hint">
                  <el-link type="warning" @click="scrollToSummary">{{ hint }}</el-link>
                </li>
              </ul>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card id="knowledge-summary-edit" shadow="never" class="summary-card">
          <div class="summary-header">
            <h3>知识概述</h3>
            <span v-if="isDirty" class="dirty-hint">有未保存修改</span>
          </div>
          <el-input
            v-model="editSummary"
            type="textarea"
            :rows="4"
            placeholder="填写攻击知识概述，便于检索与复核"
          />
        </el-card>

        <el-card shadow="never">
          <el-tabs v-model="activeTab">
            <el-tab-pane name="figures">
              <template #label>
                攻击图示
                <el-badge
                  v-if="detail.evidenceFigures.length"
                  :value="detail.evidenceFigures.length"
                  class="tab-badge"
                />
              </template>
              <KnowledgeEvidenceGallery :figures="detail.evidenceFigures" />
            </el-tab-pane>
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

    <el-dialog v-model="mergeDialogVisible" title="合并知识（Demo）" width="480px">
      <p class="merge-hint">将当前知识的支撑情报并入目标条目后，删除当前知识。仅相同标题情报在流水线中自动合并。</p>
      <el-form label-width="100px">
        <el-form-item label="目标知识">
          <el-select v-model="mergeTargetId" placeholder="选择目标" style="width: 100%">
            <el-option
              v-for="k in mergeCandidates"
              :key="k.id"
              :label="`#${k.id} ${k.title}`"
              :value="k.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mergeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="!mergeTargetId" @click="confirmMerge">
          确认合并
        </el-button>
      </template>
    </el-dialog>
  </AppLayout>
</template>

<style scoped>
.detail-top {
  margin-bottom: 16px;
}
.detail-top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}
.detail-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
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
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}
.review-tag {
  margin-left: 0;
}
.meta-card,
.summary-card {
  margin-bottom: 16px;
}
.summary-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.summary-card h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}
.dirty-hint {
  font-size: 12px;
  color: #e6a23c;
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
.tab-badge {
  margin-left: 6px;
  vertical-align: middle;
}
.merge-hint {
  margin: 0 0 16px;
  font-size: 13px;
  color: #606266;
  line-height: 1.6;
}
.actor-tag {
  margin-right: 6px;
}
.muted {
  color: #c0c4cc;
}
</style>
