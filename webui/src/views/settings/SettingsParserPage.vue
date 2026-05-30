<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import { MOCK_PARSER_PROFILES } from '@/mock/settings'

const searchQ = ref('')
const rows = ref([...MOCK_PARSER_PROFILES])

function applyFilter() {
  const q = searchQ.value.trim().toLowerCase()
  rows.value = q
    ? MOCK_PARSER_PROFILES.filter((p) =>
        [p.name, p.description].some((s) => s.toLowerCase().includes(q)),
      )
    : [...MOCK_PARSER_PROFILES]
}
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="解析模板"
        description="HTML / 微信等正文抽取规则，供采集与 parse 模块引用。"
      />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="applyFilter">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="模板 ID / 说明" style="width: 200px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="applyFilter">查询</el-button>
            <el-button disabled>新建模板（Demo）</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" class="table-card">
        <el-table :data="rows" stripe size="small">
          <el-table-column prop="name" label="模板 ID" width="200" />
          <el-table-column prop="description" label="说明" min-width="280" />
          <el-table-column prop="usedBySources" label="引用采集源" width="110" align="center" />
          <el-table-column label="操作" width="100" fixed="right">
            <template #default>
              <el-button link type="primary" disabled>编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.table-card {
  background: var(--ami-surface) !important;
}
</style>
