<script setup lang="ts">
defineProps<{
  wide?: boolean
}>()
</script>

<template>
  <div class="page-shell" :class="{ wide }">
    <slot />
  </div>
</template>

<style scoped>
.page-shell {
  max-width: 1200px;
  margin: 0 auto;
  padding: 28px 32px 40px;
}
.page-shell.wide {
  max-width: 1440px;
}
</style>
