<script setup lang="ts">
defineProps<{
  industries: string[]
}>()
</script>

<template>
  <span v-if="!industries.length" class="muted">-</span>
  <el-tag
    v-for="ind in industries"
    :key="ind"
    size="small"
    type="info"
    class="tag-item"
  >
    {{ ind }}
  </el-tag>
</template>

<style scoped>
.tag-item {
  margin-right: 4px;
}
.muted {
  color: #c0c4cc;
}
</style>
