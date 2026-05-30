<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { usePagedFilter } from '@/composables/usePagedFilter'
import { listKnowledge } from '@/api/knowledge'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import KnowledgeCards from '@/components/KnowledgeCards.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import WorkbenchHelp from '@/components/WorkbenchHelp.vue'
import { ALL_INDUSTRIES } from '@/constants/industry'
import { MALWARE_CLASSES } from '@/constants/malware'
import type { AttackMatrix } from '@/types/intel'
import type { KnowledgeSummary } from '@/types/knowledge'
import type { MalwareClass } from '@/constants/malware'

const loading = ref(false)
const items = ref<KnowledgeSummary[]>([])
const filterIndustries = ref<string[]>([])
const filterMatrix = ref<AttackMatrix | ''>('')
const filterMalwareClass = ref<MalwareClass | ''>('')
const searchQ = ref('')

async function load() {
  loading.value = true
  try {
    const result = await listKnowledge({
      industries: filterIndustries.value.length ? filterIndustries.value : undefined,
      matrix: filterMatrix.value || undefined,
      malwareClass: filterMalwareClass.value || undefined,
      q: searchQ.value || undefined,
    })
    items.value = result.items
  } finally {
    loading.value = false
  }
}

const listPager = usePagedFilter<KnowledgeSummary>(() => items.value, { pageSize: 10 })

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
        description="维护攻击手法、工具、基础设施、漏洞、恶意软件（含勒索等分类）及关联情报。"
      />
      <WorkbenchHelp mode="knowledge" />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="load">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="标题 / APT / 家族" style="width: 200px" />
          </el-form-item>
          <el-form-item label="恶意软件分类">
            <el-select v-model="filterMalwareClass" clearable placeholder="全部" style="width: 140px">
              <el-option v-for="c in MALWARE_CLASSES" :key="c.value" :label="c.label" :value="c.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="ATT&CK">
            <el-select v-model="filterMatrix" clearable placeholder="全部" style="width: 130px">
              <el-option label="Enterprise" value="Enterprise" />
              <el-option label="ICS" value="ICS" />
            </el-select>
          </el-form-item>
          <el-form-item label="攻击行业">
            <el-select
              v-model="filterIndustries"
              multiple
              collapse-tags
              clearable
              placeholder="全部"
              style="width: 200px"
            >
              <el-option v-for="ind in ALL_INDUSTRIES" :key="ind" :label="ind" :value="ind" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="load">查询</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" v-loading="loading" class="table-card">
        <KnowledgeCards v-if="listPager.total.value > 0" :items="listPager.pageData.value" />
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
.table-card {
  background: var(--ami-surface) !important;
}
.list-pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
