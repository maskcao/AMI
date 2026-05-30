<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import { COLLECT_TYPE_LABELS, MOCK_COLLECT_SOURCES } from '@/mock/settings'
import type { CollectSourceType } from '@/types/settings'

const searchQ = ref('')
const filterType = ref<CollectSourceType | ''>('')
const filterEnabled = ref<'' | 'yes' | 'no'>('')

const typeOptions = Object.entries(COLLECT_TYPE_LABELS).map(([value, label]) => ({ value, label }))

const rows = ref([...MOCK_COLLECT_SOURCES])

function applyFilter() {
  let list = [...MOCK_COLLECT_SOURCES]
  const q = searchQ.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (r) =>
        r.name.toLowerCase().includes(q) ||
        r.sourceLabel.toLowerCase().includes(q) ||
        r.typeLabel.includes(q),
    )
  }
  if (filterType.value) list = list.filter((r) => r.type === filterType.value)
  if (filterEnabled.value === 'yes') list = list.filter((r) => r.enabled)
  if (filterEnabled.value === 'no') list = list.filter((r) => !r.enabled)
  rows.value = list
}

function statusTagType(status: string) {
  if (status === 'success') return 'success'
  if (status === 'partial') return 'warning'
  if (status === 'failed') return 'danger'
  return 'info'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    success: '成功',
    partial: '部分成功',
    failed: '失败',
    skipped: '跳过',
    never: '未运行',
  }
  return map[status] ?? status
}
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="采集配置"
        description="采集源与调度，供 service ingest 模块使用；产物进入情报工作台待审。"
      />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="applyFilter">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="名称 / 来源标签" style="width: 200px" />
          </el-form-item>
          <el-form-item label="采集类型">
            <el-select v-model="filterType" clearable placeholder="全部" style="width: 160px">
              <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="启用">
            <el-select v-model="filterEnabled" clearable placeholder="全部" style="width: 100px">
              <el-option label="是" value="yes" />
              <el-option label="否" value="no" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="applyFilter">查询</el-button>
            <el-button disabled>新建采集源（Demo）</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" class="table-card">
        <el-table :data="rows" stripe size="small">
          <el-table-column prop="name" label="名称" min-width="180" />
          <el-table-column prop="typeLabel" label="类型" width="140" />
          <el-table-column prop="sourceLabel" label="来源标签" width="100" />
          <el-table-column prop="scheduleLabel" label="调度" width="110" />
          <el-table-column label="启用" width="72" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small" effect="light">
                {{ row.enabled ? '是' : '否' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="上次运行" width="100">
            <template #default="{ row }">
              {{ row.lastRunAt ? row.lastRunAt.slice(0, 10) : '—' }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.lastStatus)" size="small" effect="light">
                {{ statusLabel(row.lastStatus) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" fixed="right">
            <template #default>
              <el-button link type="primary" disabled>编辑</el-button>
              <el-button link type="primary" disabled>立即运行</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-alert
        class="hint"
        type="info"
        :closable="false"
        show-icon
        title="设计说明见 docs/system-config.md：含定时 URL、微信两阶段采集，以及 RSS、上传、Webhook 等扩展场景。"
      />
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.table-card {
  background: var(--ami-surface) !important;
}
.hint {
  margin-top: 16px;
  border-radius: var(--ami-radius);
}
</style>
