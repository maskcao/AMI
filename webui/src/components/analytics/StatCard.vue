<script setup lang="ts">
defineProps<{
  label: string
  value: number | string
  accent?: string
  hint?: string
}>()
</script>

<template>
  <div class="stat-card" :style="{ '--accent': accent ?? '#3b82f6' }">
    <span class="stat-label">{{ label }}</span>
    <span class="stat-value">{{ value }}</span>
    <span v-if="hint" class="stat-hint">{{ hint }}</span>
  </div>
</template>

<style scoped>
.stat-card {
  padding: 20px 22px;
  background: var(--ami-surface);
  border: 1px solid var(--ami-border);
  border-radius: var(--ami-radius);
  box-shadow: var(--ami-shadow);
  border-left: 4px solid var(--accent);
}
.stat-label {
  display: block;
  font-size: 13px;
  color: var(--ami-text-secondary);
  font-weight: 500;
}
.stat-value {
  display: block;
  margin-top: 8px;
  font-size: 32px;
  font-weight: 800;
  color: var(--ami-text);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.03em;
}
.stat-hint {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: var(--ami-text-muted);
  line-height: 1.4;
}
</style>
