<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { includesQuery, usePagedFilter } from '@/composables/usePagedFilter'
import { useRouter } from 'vue-router'
import { listIntel } from '@/api/intel'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import WorkbenchHelp from '@/components/WorkbenchHelp.vue'
import {
  INGEST_STATUS_LABELS,
  KNOWLEDGE_ACTION_LABELS,
  ingestStatusTagType,
} from '@/constants/intel-pipeline'
import { INTEL_MERGE_POLICY_SUMMARY } from '@/constants/merge-policy'
import type { IngestStatus, IntelSummary, KnowledgeLinkAction } from '@/types/intel'

const router = useRouter()
const loading = ref(false)
const items = ref<IntelSummary[]>([])

const filterSource = ref('')
const filterStatus = ref<IngestStatus | ''>('')
const searchQ = ref('')

const listPager = usePagedFilter(() => items.value, {
  pageSize: 10,
  matchQuery: (row, q) =>
    includesQuery(
      q,
      row.title,
      row.sourceName,
      row.collectSourceName,
      String(row.id),
      INGEST_STATUS_LABELS[row.ingestStatus],
    ),
})

const sourceOptions = [
  { label: '360', value: '360' },
  { label: 'CERT', value: 'CERT' },
]

const statusOptions = (
  ['queued', 'parsing', 'extracting', 'succeeded', 'failed'] as IngestStatus[]
).map((value) => ({ value, label: INGEST_STATUS_LABELS[value] }))

async function load() {
  loading.value = true
  try {
    const result = await listIntel({
      sourceName: filterSource.value || undefined,
      ingestStatus: filterStatus.value || undefined,
    })
    items.value = result.items
  } finally {
    loading.value = false
  }
}

function formatCollected(at: string) {
  return at ? at.slice(0, 16).replace('T', ' ') : '—'
}

function formatConfidence(v: number | null) {
  if (v == null) return '—'
  return `${Math.round(v * 100)}%`
}

function ingestStatusLabel(status: IngestStatus): string {
  return INGEST_STATUS_LABELS[status]
}

function knowledgeActionLabel(action: KnowledgeLinkAction): string {
  return KNOWLEDGE_ACTION_LABELS[action]
}

function goDetail(id: number) {
  router.push({ name: 'intel-detail', params: { id: String(id) } })
}

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}

async function search() {
  await load()
  listPager.applyFilters({ query: searchQ.value, facet: '', extras: {} })
}

watch(items, () => {
  listPager.applyFilters({ query: searchQ.value, facet: '', extras: {} })
})

onMounted(search)
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="采集队列"
        :description="`原始报告采集与处理：采集 → 解析 → AI 抽取 → 自动写入攻击知识库。${INTEL_MERGE_POLICY_SUMMARY}`"
      />
      <WorkbenchHelp mode="intel" />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="search">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="标题 / 来源 / ID" style="width: 200px" />
          </el-form-item>
          <el-form-item label="来源">
            <el-select v-model="filterSource" clearable placeholder="全部" style="width: 120px">
              <el-option v-for="opt in sourceOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="处理状态">
            <el-select v-model="filterStatus" clearable placeholder="全部" style="width: 130px">
              <el-option v-for="opt in statusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="search">查询</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" v-loading="loading" class="table-card">
        <el-table v-if="listPager.total.value > 0" :data="listPager.pageData.value" stripe size="small">
          <el-table-column prop="id" label="ID" width="72" />
          <el-table-column label="标题" min-width="260">
            <template #default="{ row }">
              <el-link type="primary" class="title-link" @click="goDetail(row.id)">{{ row.title }}</el-link>
            </template>
          </el-table-column>
          <el-table-column prop="collectSourceName" label="采集源" width="130" show-overflow-tooltip />
          <el-table-column label="采集时间" width="140">
            <template #default="{ row }">{{ formatCollected(row.collectedAt) }}</template>
          </el-table-column>
          <el-table-column label="处理状态" width="108">
            <template #default="{ row }">
              <el-tag :type="ingestStatusTagType(row.ingestStatus)" size="small" effect="light" round>
                {{ ingestStatusLabel(row.ingestStatus) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="知识写入" width="108">
            <template #default="{ row }">
              <span class="action-text">{{ knowledgeActionLabel(row.knowledgeAction) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="关联知识" width="100">
            <template #default="{ row }">
              <template v-if="row.knowledgeIds.length">
                <el-link
                  v-for="kid in row.knowledgeIds"
                  :key="kid"
                  type="primary"
                  class="kid-link"
                  @click="goKnowledge(kid)"
                >
                  #{{ kid }}
                </el-link>
              </template>
              <span v-else class="muted">—</span>
            </template>
          </el-table-column>
          <el-table-column label="抽取置信" width="88" align="center">
            <template #default="{ row }">{{ formatConfidence(row.extractConfidence) }}</template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="无匹配任务" :image-size="80" />
        <div v-if="listPager.total.value > 0" class="list-pagination">
          <el-pagination
            v-model:current-page="listPager.page"
            v-model:page-size="listPager.pageSize"
            :total="listPager.total.value"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            background
            small
            @current-change="listPager.onPageChange"
            @size-change="listPager.onSizeChange"
          />
        </div>
      </el-card>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.table-card {
  background: var(--ami-surface) !important;
}
.title-link {
  font-weight: 500;
}
.kid-link {
  margin-right: 6px;
}
.muted {
  color: #94a3b8;
  font-size: 13px;
}
.action-text {
  font-size: 12px;
  color: var(--ami-text-secondary);
}
.list-pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
