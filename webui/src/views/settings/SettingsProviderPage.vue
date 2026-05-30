<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import { MOCK_MODEL_PROVIDERS } from '@/mock/settings'

const searchQ = ref('')
const rows = ref([...MOCK_MODEL_PROVIDERS])

function applyFilter() {
  const q = searchQ.value.trim().toLowerCase()
  rows.value = q
    ? MOCK_MODEL_PROVIDERS.filter((p) =>
        [p.name, p.modelId, p.baseUrl].some((s) => s.toLowerCase().includes(q)),
      )
    : [...MOCK_MODEL_PROVIDERS]
}
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="模型服务"
        description="LLM / Embedding 网关配置，供 parse、enrich 模块调用；密钥仅存引用。"
      />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="applyFilter">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="名称 / 模型 ID" style="width: 200px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="applyFilter">查询</el-button>
            <el-button disabled>新建服务（Demo）</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" class="table-card">
        <el-table :data="rows" stripe size="small">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column label="类型" width="100">
            <template #default="{ row }">{{ row.kind === 'llm' ? 'LLM' : 'Embedding' }}</template>
          </el-table-column>
          <el-table-column prop="modelId" label="模型 ID" min-width="180" show-overflow-tooltip />
          <el-table-column prop="baseUrl" label="Base URL" min-width="200" show-overflow-tooltip />
          <el-table-column prop="defaultFor" label="默认用途" width="120" />
          <el-table-column label="启用" width="72" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
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
