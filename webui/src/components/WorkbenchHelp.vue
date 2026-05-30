<script setup lang="ts">
import { InfoFilled, WarningFilled } from '@element-plus/icons-vue'

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
      <strong>情报工作台</strong>：管理原始报告与 IOC，属于证据层。
    </div>
  </div>
  <div v-else class="help help--warn">
    <el-icon class="help-icon"><WarningFilled /></el-icon>
    <div>
      此处管理单则报告的采集与审核；TTP、工具、漏洞等请在 <strong>攻击知识库</strong> 维护并关联本情报。
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
.help--warn {
  background: #fffbeb;
  border: 1px solid #fde68a;
  color: #92400e;
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
