<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { usePagedFilter } from '@/composables/usePagedFilter'
import { listKnowledge } from '@/api/knowledge'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import KnowledgeCards from '@/components/KnowledgeCards.vue'
import type { KnowledgeTableSort } from '@/components/KnowledgeCards.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import WorkbenchHelp from '@/components/WorkbenchHelp.vue'
import { ALL_INDUSTRIES } from '@/constants/industry'
import { MALWARE_CLASSES } from '@/constants/malware'
import type { AttackMatrix } from '@/types/intel'
import type { KnowledgeReviewFilter, KnowledgeSummary } from '@/types/knowledge'
import type { MalwareClass } from '@/constants/malware'

const loading = ref(false)
const items = ref<KnowledgeSummary[]>([])
const needsReviewCount = ref(0)
const filterIndustries = ref<string[]>([])
const filterMatrix = ref<AttackMatrix | ''>('')
const filterMalwareClass = ref<MalwareClass | ''>('')
const filterReview = ref<KnowledgeReviewFilter>('all')
const searchQ = ref('')

const tableSort = ref<KnowledgeTableSort>({
  prop: 'updatedAt',
  order: 'descending',
})

const reviewOptions: { label: string; value: KnowledgeReviewFilter }[] = [
  { label: '全部', value: 'all' },
  { label: '待复核', value: 'needs_review' },
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' },
]

const sortedItems = computed(() => {
  const rows = [...items.value]
  const { prop, order } = tableSort.value
  const dir = order === 'ascending' ? 1 : -1

  return rows.sort((a, b) => {
    switch (prop) {
      case 'id':
        return (a.id - b.id) * dir
      case 'intelCount':
        return (a.intelCount - b.intelCount) * dir
      case 'qualityScore':
        return (a.qualityScore - b.qualityScore) * dir
      case 'updatedAt':
      default:
        return a.updatedAt.localeCompare(b.updatedAt) * dir
    }
  })
})

async function load() {
  loading.value = true
  try {
    const result = await listKnowledge({
      industries: filterIndustries.value.length ? filterIndustries.value : undefined,
      matrix: filterMatrix.value || undefined,
      malwareClass: filterMalwareClass.value || undefined,
      q: searchQ.value || undefined,
      review: filterReview.value === 'all' ? undefined : filterReview.value,
    })
    items.value = result.items
    needsReviewCount.value = result.needsReviewCount
  } finally {
    loading.value = false
  }
}

function onTableSortChange(payload: { prop: string; order: 'ascending' | 'descending' | null }) {
  if (!payload.prop || !payload.order) {
    tableSort.value = { prop: 'updatedAt', order: 'descending' }
  } else {
    tableSort.value = { prop: payload.prop, order: payload.order }
  }
  listPager.page.value = 1
}

const listPager = usePagedFilter<KnowledgeSummary>(() => sortedItems.value, { pageSize: 10 })

watch(items, () => {
  listPager.page.value = 1
})

onMounted(load)
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="攻击知识库"
        description="维护攻击手法、工具、基础设施、漏洞、恶意软件及关联情报；支持待复核筛选与流水线溯源。"
      >
        <template #actions>
          <el-tag v-if="needsReviewCount > 0" type="warning" effect="plain">
            {{ needsReviewCount }} 条待复核
          </el-tag>
        </template>
      </PageHeader>
      <WorkbenchHelp mode="knowledge" />

      <FilterCard>
        <el-form class="knowledge-filter" label-width="96px" @submit.prevent="load">
          <div class="filter-row">
            <el-form-item label="关键词" class="field-keyword">
              <el-input
                v-model="searchQ"
                clearable
                placeholder="标题 / 概述 / CVE / Technique ID"
              />
            </el-form-item>
            <el-form-item label="复核范围">
              <el-select v-model="filterReview" placeholder="全部">
                <el-option
                  v-for="opt in reviewOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="恶意软件分类">
              <el-select v-model="filterMalwareClass" clearable placeholder="全部">
                <el-option v-for="c in MALWARE_CLASSES" :key="c.value" :label="c.label" :value="c.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="ATT&CK">
              <el-select v-model="filterMatrix" clearable placeholder="全部">
                <el-option label="Enterprise" value="Enterprise" />
                <el-option label="ICS" value="ICS" />
              </el-select>
            </el-form-item>
          </div>
          <div class="filter-row filter-row--actions">
            <el-form-item label="攻击行业" class="field-industry">
              <el-select
                v-model="filterIndustries"
                multiple
                collapse-tags
                collapse-tags-tooltip
                clearable
                placeholder="全部"
              >
                <el-option v-for="ind in ALL_INDUSTRIES" :key="ind" :label="ind" :value="ind" />
              </el-select>
            </el-form-item>
            <el-form-item label=" " class="field-actions">
              <div class="action-buttons">
                <el-button type="primary" @click="load">查询</el-button>
              </div>
            </el-form-item>
          </div>
        </el-form>
      </FilterCard>

      <el-card shadow="never" v-loading="loading" class="table-card">
        <KnowledgeCards
          v-if="listPager.total.value > 0"
          :items="listPager.pageData.value"
          :default-sort="tableSort"
          @sort-change="onTableSortChange"
        />
        <el-empty v-else description="无匹配知识条目" :image-size="80" />
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
.knowledge-filter :deep(.el-form-item) {
  margin-bottom: 0;
}

.filter-row {
  display: grid;
  grid-template-columns: minmax(220px, 2fr) repeat(3, minmax(120px, 1fr));
  gap: 12px 16px;
  align-items: start;
}

.filter-row--actions {
  margin-top: 12px;
  grid-template-columns: minmax(280px, 2fr) 1fr;
}

.field-keyword :deep(.el-input),
.field-industry :deep(.el-select) {
  width: 100%;
}

.filter-row :deep(.el-select) {
  width: 100%;
}

.field-actions :deep(.el-form-item__label) {
  visibility: hidden;
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

@media (max-width: 1100px) {
  .filter-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .filter-row--actions {
    grid-template-columns: 1fr;
  }
  .action-buttons {
    justify-content: flex-start;
  }
}

.table-card {
  width: 100%;
  background: var(--ami-surface) !important;
}
.table-card :deep(.el-card__body) {
  overflow-x: auto;
}
.list-pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
