<script setup lang="ts">
import { useRouter } from 'vue-router'
import IndustryTags from '@/components/IndustryTags.vue'
import MalwareClassTags from '@/components/MalwareClassTags.vue'
import QualityScore from '@/components/QualityScore.vue'
import type { KnowledgeSummary } from '@/types/knowledge'

defineProps<{
  items: KnowledgeSummary[]
}>()

const router = useRouter()

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}

function dimSummary(d: KnowledgeSummary['dimensions']) {
  return `手法 ${d.techniques} · 工具 ${d.tools} · 设施 ${d.infrastructure} · 漏洞 ${d.vulnerabilities}`
}
</script>

<template>
  <el-table :data="items" stripe style="width: 100%">
    <el-table-column prop="id" label="ID" width="72" />
    <el-table-column label="攻击知识" min-width="260">
      <template #default="{ row }">
        <el-link type="primary" @click="goKnowledge(row.id)">{{ row.title }}</el-link>
      </template>
    </el-table-column>
    <el-table-column label="矩阵" width="110">
      <template #default="{ row }">
        <el-tag :type="row.attackMatrix === 'ICS' ? 'warning' : 'primary'" size="small">
          {{ row.attackMatrix }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="行业" width="130">
      <template #default="{ row }">
        <IndustryTags :industries="row.targetIndustries" />
      </template>
    </el-table-column>
    <el-table-column label="恶意软件" min-width="200">
      <template #default="{ row }">
        <MalwareClassTags :items="row.malware" />
      </template>
    </el-table-column>
    <el-table-column label="知识维度" min-width="220">
      <template #default="{ row }">
        <span class="dim-text">{{ dimSummary(row.dimensions) }}</span>
      </template>
    </el-table-column>
    <el-table-column label="情报" width="64" align="center">
      <template #default="{ row }">
        <el-badge :value="row.intelCount" />
      </template>
    </el-table-column>
    <el-table-column label="质量" width="120">
      <template #default="{ row }">
        <QualityScore :score="row.qualityScore" />
      </template>
    </el-table-column>
  </el-table>
</template>

<style scoped>
.dim-text {
  font-size: 12px;
  color: #606266;
}
</style>
