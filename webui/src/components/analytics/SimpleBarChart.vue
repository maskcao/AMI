<script setup lang="ts">
import type { CountItem } from '@/analytics/aggregate'

const props = defineProps<{
  items?: CountItem[]
  maxItems?: number
}>()

const display = () => (props.items ?? []).slice(0, props.maxItems ?? 10)
const maxCount = () => Math.max(...display().map((i) => i.count), 1)
</script>

<template>
  <div class="bar-chart">
    <div v-for="item in display()" :key="item.name" class="bar-row">
      <span class="bar-label" :title="item.name">{{ item.name }}</span>
      <div class="bar-track">
        <div
          class="bar-fill"
          :style="{ width: `${(item.count / maxCount()) * 100}%` }"
        />
      </div>
      <span class="bar-value">{{ item.count }}</span>
    </div>
    <el-empty v-if="!display().length" description="暂无数据" :image-size="48" />
  </div>
</template>

<style scoped>
.bar-chart {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 4px 0;
}
.bar-row {
  display: grid;
  grid-template-columns: minmax(80px, 140px) 1fr 40px;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}
.bar-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--ami-text-secondary);
  font-weight: 500;
}
.bar-track {
  height: 8px;
  background: #f1f5f9;
  border-radius: 4px;
  overflow: hidden;
}
.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #818cf8);
  border-radius: 4px;
  min-width: 6px;
  transition: width 0.35s ease;
}
.bar-value {
  text-align: right;
  font-weight: 700;
  font-size: 13px;
  color: var(--ami-text);
  font-variant-numeric: tabular-nums;
}
</style>
