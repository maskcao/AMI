<script setup lang="ts">
defineProps<{
  actors: string[]
}>()
</script>

<template>
  <span v-if="!actors?.length" class="muted">—</span>
  <span v-else class="actor-tags">
    <el-tag
      v-for="a in actors"
      :key="a"
      size="small"
      effect="plain"
      round
      class="actor-tag"
    >
      {{ a }}
    </el-tag>
  </span>
</template>

<style scoped>
.actor-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.actor-tag {
  max-width: 100%;
}
.actor-tag :deep(.el-tag__content) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.muted {
  color: #94a3b8;
  font-size: 13px;
}
</style>
