<script setup lang="ts">
import { InfoFilled } from '@element-plus/icons-vue'

defineProps<{
  mode: 'knowledge' | 'intel'
}>()
</script>

<template>
  <div v-if="mode === 'knowledge'" class="help help--info">
    <el-icon class="help-icon"><InfoFilled /></el-icon>
    <div>
      <strong>攻击知识库</strong>：沉淀手法、工具、基础设施、漏洞、恶意软件等结构化认知，可挂载多份情报作为依据。
      <span class="sep">|</span>
      可选人工合并/纠错；情报侧由流水线自动写入，不在此审核单则报告。
    </div>
  </div>
  <div v-else class="help help--info">
    <el-icon class="help-icon"><InfoFilled /></el-icon>
    <div>
      <strong>采集处理队列</strong>：采集 → 解析 → AI 抽取 → 写攻击知识（<strong>仅同标题报告合并</strong>，否则新建）。
      本页用于观测状态、失败重试与溯源；TTP、实体维度请在 <strong>攻击知识库</strong> 维护。
    </div>
  </div>
</template>

<style scoped>
.help {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 14px 16px;
  border-radius: var(--ami-radius);
  margin-bottom: 20px;
  font-size: 13px;
  line-height: 1.6;
}
.help--info {
  background: var(--ami-accent-soft);
  border: 1px solid #bfdbfe;
  color: #1e40af;
}
.help-icon {
  font-size: 18px;
  margin-top: 2px;
  flex-shrink: 0;
}
.sep {
  margin: 0 8px;
  opacity: 0.5;
}
</style>
