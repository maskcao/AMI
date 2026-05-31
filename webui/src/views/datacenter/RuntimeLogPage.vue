<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { clearRuntimeLogs, listRuntimeLogs, type RuntimeLogEntry, type RuntimeLogLevel } from '@/api/datacenter'
import { USE_API } from '@/api/http'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'

const loading = ref(false)
const items = ref<RuntimeLogEntry[]>([])
const total = ref(0)
const sources = ref<string[]>([])
const filterLevel = ref<RuntimeLogLevel | ''>('')
const filterSource = ref('')
const searchQ = ref('')
const autoRefresh = ref(true)
let timer: ReturnType<typeof setInterval> | null = null

const levelOptions = [
  { label: '全部', value: '' },
  { label: 'info', value: 'info' },
  { label: 'warn', value: 'warn' },
  { label: 'error', value: 'error' },
]

async function load() {
  if (!USE_API) {
    items.value = []
    total.value = 0
    return
  }
  loading.value = true
  try {
    const res = await listRuntimeLogs({
      level: filterLevel.value || undefined,
      source: filterSource.value || undefined,
      q: searchQ.value.trim() || undefined,
      limit: 300,
    })
    items.value = res.items
    total.value = res.total
    sources.value = res.sources
  } catch (e) {
    const msg = e instanceof Error ? e.message : '加载日志失败'
    if (/not found/i.test(msg)) {
      ElMessage.error('运行日志接口未就绪，请重新编译并重启 API（cd service && make run）')
    } else {
      ElMessage.error(msg)
    }
  } finally {
    loading.value = false
  }
}

async function confirmClear() {
  try {
    await ElMessageBox.confirm('确定清空全部运行日志？', '清空确认', { type: 'warning' })
    await clearRuntimeLogs()
    ElMessage.success('已清空')
    await load()
  } catch (e) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e instanceof Error ? e.message : '清空失败')
  }
}

function levelTagType(level: string) {
  if (level === 'error') return 'danger'
  if (level === 'warn') return 'warning'
  return 'info'
}

function formatTime(iso: string) {
  return iso.replace('T', ' ').slice(0, 19)
}

function setupTimer() {
  if (timer) clearInterval(timer)
  timer = null
  if (autoRefresh.value) {
    timer = setInterval(() => {
      void load()
    }, 8000)
  }
}

onMounted(() => {
  void load()
  setupTimer()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="运行日志"
        :description="`采集、流水线、调度与 API 触发的运行记录（共 ${total} 条，展示最近 300 条）。`"
      />

      <el-alert
        v-if="!USE_API"
        type="warning"
        :closable="false"
        show-icon
        title="当前为 Mock 模式，运行日志需连接 API 服务（VITE_USE_API=true）。"
        class="api-hint"
      />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="load">
          <el-form-item label="级别">
            <el-select v-model="filterLevel" clearable placeholder="全部" style="width: 110px">
              <el-option v-for="opt in levelOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="来源">
            <el-select v-model="filterSource" clearable placeholder="全部" style="width: 140px">
              <el-option v-for="s in sources" :key="s" :label="s" :value="s" />
            </el-select>
          </el-form-item>
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="消息 / 详情" style="width: 200px" />
          </el-form-item>
          <el-form-item label="自动刷新">
            <el-switch v-model="autoRefresh" @change="setupTimer" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="load">查询</el-button>
            <el-button :disabled="!USE_API" @click="confirmClear">清空</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" v-loading="loading" class="table-card">
        <el-table :data="items" stripe size="small" class="log-table" empty-text="暂无日志">
          <el-table-column label="时间" width="168">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="级别" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="levelTagType(row.level)" size="small" effect="light">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="source" label="来源" width="100" />
          <el-table-column prop="message" label="消息" min-width="240" show-overflow-tooltip />
          <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
        </el-table>
      </el-card>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.api-hint {
  margin-bottom: 12px;
}
.table-card {
  background: var(--ami-surface) !important;
}
.log-table {
  width: 100%;
}
</style>
