<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <header class="page-header">
    <div class="page-header__text">
      <h1>{{ title }}</h1>
      <p v-if="description" class="desc">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}
.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--ami-text);
  letter-spacing: -0.02em;
}
.desc {
  margin: 8px 0 0;
  font-size: 14px;
  color: var(--ami-text-secondary);
  line-height: 1.5;
  max-width: 720px;
}
</style>
