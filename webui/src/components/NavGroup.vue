<script setup lang="ts">
import { ref, watch, type Component } from 'vue'

const props = defineProps<{
  label: string
  icon: Component
  open: boolean
  active?: boolean
}>()

defineEmits<{
  toggle: []
}>()

/** Mount submenu once (idle); then only toggle visibility — avoids layout work on each expand. */
const submenuMounted = ref(props.open)

watch(
  () => props.open,
  (v) => {
    if (v) submenuMounted.value = true
  },
)

function mountSubmenu() {
  submenuMounted.value = true
}

defineExpose({ mountSubmenu })
</script>

<template>
  <div class="nav-group" :class="{ open, active }">
    <button type="button" class="nav-item nav-group-head" @click="$emit('toggle')">
      <el-icon class="nav-icon"><component :is="icon" /></el-icon>
      <span class="nav-label">{{ label }}</span>
      <span class="nav-chevron" :class="{ open }">›</span>
    </button>
    <div v-if="submenuMounted" v-show="open" class="nav-sub">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 11px 14px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: #cbd5e1;
  font-size: 14px;
  cursor: pointer;
  text-align: left;
}
.nav-item:hover {
  background: var(--ami-sidebar-hover);
  color: #f1f5f9;
}
.nav-group.active .nav-group-head {
  color: #f1f5f9;
}
.nav-group-head {
  margin-bottom: 2px;
}
.nav-chevron {
  margin-left: auto;
  font-size: 16px;
  opacity: 0.7;
  transform: rotate(0deg);
}
.nav-chevron.open {
  transform: rotate(90deg);
}
.nav-sub {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 4px 0 8px 8px;
  margin-left: 12px;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
  contain: layout style;
}
.nav-icon {
  font-size: 18px;
}
.nav-label {
  flex: 1;
}
</style>
