<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { LinkedIntelBrief } from '@/types/knowledge'

defineProps<{
  items: LinkedIntelBrief[]
}>()

const router = useRouter()

function goIntel(id: number) {
  router.push({ name: 'intel-detail', params: { id: String(id) } })
}

function formatDisclosed(at: string | null) {
  return at ?? '未知'
}
</script>

<template>
  <el-table v-if="items.length" :data="items" size="small" stripe>
    <el-table-column label="角色" width="100">
      <template #default="{ row }">
        <el-tag :type="row.role === 'primary' ? 'primary' : 'info'" size="small">
          {{ row.role === 'primary' ? '主要' : '补充' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="情报标题" min-width="280">
      <template #default="{ row }">
        <el-link type="primary" @click="goIntel(row.id)">{{ row.title }}</el-link>
      </template>
    </el-table-column>
    <el-table-column prop="sourceName" label="来源" width="90" />
    <el-table-column label="披露时间" width="110">
      <template #default="{ row }">
        {{ formatDisclosed(row.disclosedAt) }}
      </template>
    </el-table-column>
  </el-table>
  <el-empty v-else description="暂无挂载情报" :image-size="64" />
</template>
