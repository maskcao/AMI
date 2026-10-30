<script setup lang="ts">
import { useRouter } from 'vue-router'
import IndustryTags from '@/components/IndustryTags.vue'
import ThreatActorTags from '@/components/ThreatActorTags.vue'
import MalwareClassTags from '@/components/MalwareClassTags.vue'
import QualityScore from '@/components/QualityScore.vue'
import { formatDisclosedAt } from '@/utils/disclosed-at'
import { KNOWLEDGE_STATUS_LABELS } from '@/utils/knowledge-review'
import type { KnowledgeStatus, KnowledgeSummary } from '@/types/knowledge'

export interface KnowledgeTableSort {
  prop: string
  order: 'ascending' | 'descending'
}

defineProps<{
  items: KnowledgeSummary[]
  defaultSort: KnowledgeTableSort
}>()

const emit = defineEmits<{
  sortChange: [payload: { prop: string; order: 'ascending' | 'descending' | null }]
}>()

const router = useRouter()

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}

function formatUpdatedAt(iso: string): string {
  return iso.replace('T', ' ').slice(0, 16)
}

function statusLabel(status: KnowledgeStatus): string {
  return KNOWLEDGE_STATUS_LABELS[status]
}

function displayTitle(title: string): string {
  const t = title?.trim()
  return t || '（无标题）'
}

function onSortChange(data: { prop: string; order: string | null }) {
  emit('sortChange', {
    prop: data.prop,
    order: data.order as 'ascending' | 'descending' | null,
  })
}
</script>

<template>
  <el-table
    class="knowledge-table"
    :data="items"
    stripe
    :default-sort="{ prop: defaultSort.prop, order: defaultSort.order }"
    @sort-change="onSortChange"
  >
    <el-table-column prop="id" label="ID" width="72" sortable="custom" />
    <el-table-column label="标题" min-width="280" class-name="ami-table-title-cell">
      <template #default="{ row }">
        <div class="title-cell">
          <el-link
            type="primary"
            class="ami-title-wrap knowledge-title-link"
            :class="{ 'is-untitled': !row.title?.trim() }"
            @click="goKnowledge(row.id)"
          >
            {{ displayTitle(row.title) }}
          </el-link>
          <el-tag v-if="row.needsReview" type="warning" size="small" effect="plain" class="review-tag">
            待复核
          </el-tag>
          <el-tag
            v-if="row.enrichVersion === 'rules-v1'"
            type="danger"
            size="small"
            effect="plain"
            class="review-tag"
            title="关联情报的 LLM 富集失败，当前为规则兜底结果"
          >
            LLM 失败
          </el-tag>
          <el-tag
            v-if="row.dupCandidates?.length"
            type="danger"
            size="small"
            effect="plain"
            class="review-tag"
            :title="`疑似重复：与 ${row.dupCandidates.length} 条知识高度相似`"
          >
            疑似重复
          </el-tag>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="行业" width="120">
      <template #default="{ row }">
        <IndustryTags :industries="row.targetIndustries" />
      </template>
    </el-table-column>
    <el-table-column label="攻击组织" min-width="160" show-overflow-tooltip>
      <template #default="{ row }">
        <ThreatActorTags :actors="row.threatActors ?? []" />
      </template>
    </el-table-column>
    <el-table-column label="恶意软件" min-width="180" class-name="malware-cell">
      <template #default="{ row }">
        <MalwareClassTags :items="row.malware" />
      </template>
    </el-table-column>
    <el-table-column
      prop="disclosedAt"
      label="披露时间"
      width="140"
      sortable="custom"
      :sort-orders="['descending', 'ascending']"
    >
      <template #default="{ row }">
        <span
          class="time-cell"
          :title="row.disclosedAt ? '关联情报中最早的披露时间' : '暂无关联情报披露时间'"
        >
          {{ formatDisclosedAt(row.disclosedAt) }}
        </span>
      </template>
    </el-table-column>
    <el-table-column
      prop="updatedAt"
      label="更新时间"
      width="156"
      sortable="custom"
      :sort-orders="['descending', 'ascending']"
    >
      <template #default="{ row }">
        <span class="time-cell">{{ formatUpdatedAt(row.updatedAt) }}</span>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="88">
      <template #default="{ row }">
        <el-tag :type="row.status === 'published' ? 'success' : 'info'" size="small">
          {{ statusLabel(row.status) }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="qualityScore" label="质量" width="100" sortable="custom">
      <template #default="{ row }">
        <QualityScore :score="row.qualityScore" />
      </template>
    </el-table-column>
  </el-table>
</template>

<style scoped>
.knowledge-table {
  width: 100%;
}
.knowledge-table :deep(.malware-cell .cell) {
  white-space: normal;
  word-break: break-word;
  line-height: 1.5;
  align-items: flex-start;
}
.title-cell {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 6px;
}
.knowledge-title-link {
  height: auto;
  line-height: 1.5;
}
.knowledge-title-link :deep(.el-link__inner) {
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}
.review-tag {
  flex-shrink: 0;
}
.title-cell .is-untitled {
  color: var(--el-text-color-secondary);
  font-style: italic;
}
.time-cell {
  font-size: 12px;
  color: #303133;
  white-space: nowrap;
}
</style>
