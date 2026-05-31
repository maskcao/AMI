<script setup lang="ts">
import { useRouter } from 'vue-router'
import IndustryTags from '@/components/IndustryTags.vue'
import MalwareClassTags from '@/components/MalwareClassTags.vue'
import QualityScore from '@/components/QualityScore.vue'
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

function dimSummary(d: KnowledgeSummary['dimensions']) {
  return `手法 ${d.techniques} · 工具 ${d.tools} · 设施 ${d.infrastructure} · 漏洞 ${d.vulnerabilities}`
}

function formatUpdatedAt(iso: string): string {
  return iso.replace('T', ' ').slice(0, 16)
}

function statusLabel(status: KnowledgeStatus): string {
  return KNOWLEDGE_STATUS_LABELS[status]
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
    <el-table-column label="攻击知识" min-width="280" show-overflow-tooltip>
      <template #default="{ row }">
        <div class="title-cell">
          <el-link type="primary" @click="goKnowledge(row.id)">{{ row.title }}</el-link>
          <el-tag v-if="row.needsReview" type="warning" size="small" effect="plain" class="review-tag">
            待复核
          </el-tag>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="88">
      <template #default="{ row }">
        <el-tag :type="row.status === 'published' ? 'success' : 'info'" size="small">
          {{ statusLabel(row.status) }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="矩阵" width="100">
      <template #default="{ row }">
        <el-tag :type="row.attackMatrix === 'ICS' ? 'warning' : 'primary'" size="small">
          {{ row.attackMatrix }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="行业" width="120">
      <template #default="{ row }">
        <IndustryTags :industries="row.targetIndustries" />
      </template>
    </el-table-column>
    <el-table-column label="恶意软件" min-width="180" show-overflow-tooltip>
      <template #default="{ row }">
        <MalwareClassTags :items="row.malware" />
      </template>
    </el-table-column>
    <el-table-column label="知识维度" min-width="260" show-overflow-tooltip>
      <template #default="{ row }">
        <span class="dim-text">{{ dimSummary(row.dimensions) }}</span>
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
    <el-table-column prop="intelCount" label="情报" width="80" align="center" sortable="custom">
      <template #default="{ row }">
        <el-badge :value="row.intelCount" />
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
.title-cell {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}
.review-tag {
  flex-shrink: 0;
}
.dim-text {
  font-size: 12px;
  color: #606266;
}
.time-cell {
  font-size: 12px;
  color: #303133;
  white-space: nowrap;
}
</style>
