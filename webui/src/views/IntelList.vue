<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { includesQuery, usePagedFilter } from '@/composables/usePagedFilter'
import { useRouter } from 'vue-router'
import { listIntel } from '@/api/intel'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import QualityScore from '@/components/QualityScore.vue'
import WorkbenchHelp from '@/components/WorkbenchHelp.vue'

const router = useRouter()
const loading = ref(false)
const items = ref<Awaited<ReturnType<typeof listIntel>>['items']>([])

const filterSource = ref('')
const filterStatus = ref('')

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

const searchQ = ref('')

const listPager = usePagedFilter(() => items.value, {
  pageSize: 10,
  matchQuery: (row, q) =>
    includesQuery(q, row.title, row.sourceName, String(row.id), statusLabel(row.ingestStatus)),
})

const sourceOptions = [
  { label: '360', value: '360' },
  { label: 'CERT', value: 'CERT' },
]
const statusOptions = [
  { label: '已发布', value: 'published' },
  { label: '待审', value: 'pending_review' },
]

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

function formatDisclosed(at: string | null) {
  return at ?? '未知'
}

function statusType(s: string) {
  if (s === 'published') return 'success'
  if (s === 'pending_review') return 'warning'
  return 'info'
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
        title="情报工作台"
        description="采集与审核原始威胁报告；结构化内容在攻击知识库维护并挂载关联。"
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
          <el-form-item label="状态">
            <el-select v-model="filterStatus" clearable placeholder="全部" style="width: 120px">
              <el-option v-for="opt in statusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="search">查询</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" v-loading="loading" class="table-card">
        <el-table v-if="listPager.total.value > 0" :data="listPager.pageData.value" stripe>
          <el-table-column prop="id" label="ID" width="72" />
          <el-table-column label="情报标题" min-width="300">
            <template #default="{ row }">
              <el-link type="primary" class="title-link" @click="goDetail(row.id)">{{ row.title }}</el-link>
            </template>
          </el-table-column>
          <el-table-column label="披露时间" width="100">
            <template #default="{ row }">{{ formatDisclosed(row.disclosedAt) }}</template>
          </el-table-column>
          <el-table-column prop="sourceName" label="来源" width="88" />
          <el-table-column label="状态" width="96">
            <template #default="{ row }">
              <el-tag :type="statusType(row.ingestStatus)" size="small" effect="light" round>
                {{ statusLabel(row.ingestStatus) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="关联攻击知识" width="180">
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
              <span v-else class="muted">未关联</span>
            </template>
          </el-table-column>
          <el-table-column label="IOC" width="64" align="center">
            <template #default="{ row }">{{ row.iocCount }}</template>
          </el-table-column>
          <el-table-column label="质量" width="120">
            <template #default="{ row }">
              <QualityScore :score="row.qualityScore" />
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="无匹配情报" :image-size="80" />
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
  margin-right: 8px;
}
.muted {
  color: #94a3b8;
  font-size: 13px;
}
.list-pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
