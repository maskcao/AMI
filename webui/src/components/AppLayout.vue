<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'
</script>

<template>
  <el-container class="app-layout ami-app-layout">
    <AppSidebar />
    <el-main class="main">
      <slot />
    </el-main>
  </el-container>
</template>

<style scoped>
.app-layout {
  height: 100%;
  min-height: 100vh;
  overflow: hidden;
}

.app-layout :deep(.el-aside) {
  height: 100%;
  overflow: hidden;
  flex-shrink: 0;
}

.main {
  flex: 1;
  min-width: 0;
  height: 100%;
  padding: 0;
  background: var(--ami-bg);
  overflow-x: hidden;
  overflow-y: auto;
  contain: layout style;
}
</style>
