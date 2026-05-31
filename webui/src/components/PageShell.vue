<script setup lang="ts">
/** @deprecated wide is default; prop kept for compatibility */
defineProps<{
  wide?: boolean
}>()
</script>

<template>
  <div class="page-shell">
    <slot />
  </div>
</template>

<style scoped>
.page-shell {
  width: 100%;
  max-width: none;
  margin: 0;
  padding: 24px 28px 40px;
  box-sizing: border-box;
}
</style>
