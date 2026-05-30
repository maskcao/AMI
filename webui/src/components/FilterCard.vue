<template>
  <el-card shadow="never" class="filter-card">
    <slot />
  </el-card>
</template>

<style scoped>
.filter-card {
  margin-bottom: 16px;
  background: var(--ami-surface) !important;
}
.filter-card :deep(.el-form-item) {
  margin-bottom: 0;
}
</style>
