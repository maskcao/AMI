<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { LinkedIntelBrief } from '@/types/knowledge'

defineProps<{
  items: LinkedIntelBrief[]
}>()

const router = useRouter()

function formatDisclosed(at: string | null) {
  return at ?? '未知'
}

function goIntel(id: number, tab?: string) {
  void router.push({
    name: 'intel-detail',
    params: { id: String(id) },
    query: tab ? { tab } : {},
  })
}

function openSourceUrl(url?: string) {
  if (!url?.trim()) return
  window.open(url, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <el-table v-if="items.length" :data="items" size="small" stripe>
    <el-table-column label="角色" width="88">
      <template #default="{ row }">
        <el-tag :type="row.role === 'primary' ? 'primary' : 'info'" size="small">
          {{ row.role === 'primary' ? '主要' : '补充' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="情报标题" min-width="220">
      <template #default="{ row }">
        <el-link type="primary" @click="goIntel(row.id)">{{ row.title }}</el-link>
        <div v-if="row.collectSourceName" class="sub-line">采集源：{{ row.collectSourceName }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="sourceName" label="来源" width="88" />
    <el-table-column label="披露时间" width="100">
      <template #default="{ row }">
        {{ formatDisclosed(row.disclosedAt) }}
      </template>
    </el-table-column>
    <el-table-column label="调阅" min-width="280" fixed="right">
      <template #default="{ row }">
        <div class="actions">
          <el-button size="small" type="primary" link @click="goIntel(row.id)">
            采集队列
          </el-button>
          <el-button
            size="small"
            type="success"
            link
            :disabled="!row.hasRaw"
            @click="goIntel(row.id, 'html')"
          >
            HTML 快照
          </el-button>
          <el-button
            size="small"
            type="success"
            link
            :disabled="!row.hasParsed"
            @click="goIntel(row.id, 'markdown')"
          >
            Markdown
          </el-button>
          <el-button
            size="small"
            type="info"
            link
            :disabled="!row.sourceUrl"
            @click="openSourceUrl(row.sourceUrl)"
          >
            原始 URL
          </el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>
  <p v-if="items.length" class="hint">
    建议优先使用「HTML 快照 / Markdown」查看落盘正文；原始 URL 可能因链接失效无法打开。
  </p>
  <el-empty v-else description="暂无挂载情报" :image-size="64" />
</template>

<style scoped>
.sub-line {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 10px;
}
.hint {
  margin: 10px 0 0;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}
</style>
