<script setup lang="ts">
import { computed } from 'vue'
import type { TechniqueRow } from '@/analytics/aggregate'
import { tacticLabel, techniqueLabel } from '@/constants/mitre-zh'

const props = defineProps<{
  techniques: TechniqueRow[]
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

interface TreeNode {
  id: string
  label: string
  count?: number
  children?: TreeNode[]
}

const treeData = computed<TreeNode[]>(() => {
  const byTactic = new Map<string, TechniqueRow[]>()
  for (const t of props.techniques) {
    const list = byTactic.get(t.tactic) ?? []
    list.push(t)
    byTactic.set(t.tactic, list)
  }

  return [
    {
      id: '',
      label: `全部战术（${props.techniques.length}）`,
      children: [...byTactic.entries()]
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([tactic, techs]) => ({
          id: `tactic:${tactic}`,
          label: `${tacticLabel(tactic)}（${techs.length}）`,
          children: techs
            .sort((a, b) => a.techniqueId.localeCompare(b.techniqueId))
            .map((t) => ({
              id: `tech:${t.techniqueId}`,
              label: `${t.techniqueId} · ${techniqueLabel(t.techniqueId, t.name)}`,
            })),
        })),
    },
  ]
})

const currentNodeKey = computed(() => props.modelValue || '')

function onNodeClick(data: TreeNode) {
  emit('update:modelValue', data.id)
}
</script>

<template>
  <div class="mitre-tree">
    <p class="tree-title">战术 / 技术</p>
    <el-tree
      :data="treeData"
      node-key="id"
      :current-node-key="currentNodeKey"
      highlight-current
      default-expand-all
      :expand-on-click-node="false"
      @node-click="onNodeClick"
    />
  </div>
</template>

<style scoped>
.mitre-tree {
  padding: 12px;
  border: 1px solid var(--ami-border);
  border-radius: var(--ami-radius);
  background: #fff;
}
.tree-title {
  margin: 0 0 10px;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
}
.mitre-tree :deep(.el-tree-node__content) {
  height: 32px;
  border-radius: 6px;
}
.mitre-tree :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background: #eff6ff;
  color: #1d4ed8;
}
</style>
