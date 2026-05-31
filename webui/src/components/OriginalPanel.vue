<script setup lang="ts">
defineProps<{
  html: string
  collectedAt: string
  parseVersion: string
}>()
</script>

<template>
  <div class="original-panel">
    <el-descriptions :column="2" border size="small" class="meta">
      <el-descriptions-item label="采集时间">{{ collectedAt }}</el-descriptions-item>
      <el-descriptions-item label="解析版本">{{ parseVersion }}</el-descriptions-item>
    </el-descriptions>
    <div class="original-body" v-html="html" />
  </div>
</template>

<style scoped>
.original-panel {
  padding-top: 8px;
}
.meta {
  margin-bottom: 16px;
}
.original-body {
  padding: 16px;
  background: #fafafa;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  line-height: 1.6;
  font-size: 14px;
}
.original-body :deep(img) {
  display: block;
  max-width: 100%;
  height: auto;
  margin: 12px 0;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  cursor: zoom-in;
}
.original-body :deep(figure.report-figure) {
  margin: 16px 0;
}
</style>
