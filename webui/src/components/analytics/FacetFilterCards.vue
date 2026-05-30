<script setup lang="ts">
export interface FacetItem {
  key: string
  label: string
  count: number
}

const props = defineProps<{
  title?: string
  items: FacetItem[]
  modelValue: string
  allCount?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [key: string]
}>()
</script>

<template>
  <div class="facet-cards">
    <p v-if="title" class="facet-title">{{ title }}</p>
    <button
      type="button"
      class="facet-card"
      :class="{ active: !modelValue }"
      @click="emit('update:modelValue', '')"
    >
      <span class="facet-label">全部</span>
      <span class="facet-count">{{ allCount ?? items.reduce((s, i) => s + i.count, 0) }}</span>
    </button>
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="facet-card"
      :class="{ active: modelValue === item.key }"
      @click="emit('update:modelValue', item.key)"
    >
      <span class="facet-label">{{ item.label }}</span>
      <span class="facet-count">{{ item.count }}</span>
    </button>
  </div>
</template>

<style scoped>
.facet-cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.facet-title {
  margin: 0 0 4px;
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
}
.facet-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--ami-border);
  border-radius: 8px;
  background: #fff;
  color: var(--ami-text);
  font-size: 13px;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.15s, background 0.15s;
}
.facet-card:hover {
  border-color: #93c5fd;
  background: #f8fafc;
}
.facet-card.active {
  border-color: #3b82f6;
  background: #eff6ff;
  box-shadow: inset 3px 0 0 #3b82f6;
}
.facet-label {
  flex: 1;
  line-height: 1.3;
}
.facet-count {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  background: #f1f5f9;
  color: #64748b;
  font-variant-numeric: tabular-nums;
}
.facet-card.active .facet-count {
  background: #dbeafe;
  color: #1d4ed8;
}
</style>
