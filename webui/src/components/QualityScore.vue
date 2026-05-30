<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  score: number
}>()

const color = computed(() => {
  if (props.score >= 70) return '#67c23a'
  if (props.score >= 40) return '#e6a23c'
  return '#f56c6c'
})
</script>

<template>
  <div class="quality-score">
    <el-progress
      :percentage="score"
      :color="color"
      :stroke-width="6"
      :show-text="false"
      style="width: 48px"
    />
    <span class="quality-score__text">{{ score }}%</span>
  </div>
</template>

<style scoped>
.quality-score {
  display: flex;
  align-items: center;
  gap: 8px;
}
.quality-score__text {
  font-size: 13px;
  color: #606266;
  min-width: 36px;
}
</style>
