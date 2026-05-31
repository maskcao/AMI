<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { KnowledgePipelineSource } from '@/types/knowledge'
import type { KnowledgeLinkAction } from '@/types/intel'

defineProps<{
  pipeline: KnowledgePipelineSource | null
}>()

const router = useRouter()

const actionLabels: Record<KnowledgeLinkAction, string> = {
  created: '新建知识',
  updated: '更新知识',
  skipped: '未写入',
  pending: '待处理',
}

function goIntel(id: number) {
  router.push({ name: 'intel-detail', params: { id: String(id) } })
}

function goIntelList() {
  router.push({ name: 'intel-list' })
}

function formatConfidence(v: number | null): string {
  if (v == null) return '-'
  return `${Math.round(v * 100)}%`
}
</script>

<template>
  <el-card v-if="pipeline" shadow="never" class="pipeline-card">
    <template #header>
      <div class="pipeline-header">
        <span>流水线溯源</span>
        <el-button link type="primary" size="small" @click="goIntelList">采集队列</el-button>
      </div>
    </template>
    <el-descriptions :column="2" border size="small">
      <el-descriptions-item label="最近写入">
        <el-tag size="small" type="success">
          {{ actionLabels[pipeline.knowledgeAction] ?? pipeline.knowledgeAction }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="抽取置信度">
        {{ formatConfidence(pipeline.extractConfidence) }}
      </el-descriptions-item>
      <el-descriptions-item label="来源情报" :span="2">
        <el-link type="primary" @click="goIntel(pipeline.lastIntelId)">
          #{{ pipeline.lastIntelId }} · {{ pipeline.lastIntelTitle }}
        </el-link>
      </el-descriptions-item>
      <el-descriptions-item label="流水线时间" :span="2">
        {{ pipeline.pipelinedAt }}
      </el-descriptions-item>
    </el-descriptions>
  </el-card>
  <el-alert
    v-else
    type="info"
    :closable="false"
    show-icon
    title="暂无流水线写入记录"
    description="该知识条目可能为手工创建，或尚未完成 AI 抽取写入。"
  />
</template>

<style scoped>
.pipeline-card {
  margin-bottom: 16px;
}
.pipeline-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}
</style>
