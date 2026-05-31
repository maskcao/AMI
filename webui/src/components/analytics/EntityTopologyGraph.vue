<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { LinkedKnowledgeRef } from '@/analytics/actors'

export interface TopologyDimensionNode {
  key: string
  label: string
  count: number
  color: string
}

const props = defineProps<{
  centerLabel: string
  centerSubtitle: string
  dimensionNodes: TopologyDimensionNode[]
  knowledge: LinkedKnowledgeRef[]
  hint?: string
}>()

const router = useRouter()
const CX = 360
const CY = 190

function polar(cx: number, cy: number, r: number, index: number, total: number) {
  const angle = (index / total) * 2 * Math.PI - Math.PI / 2
  return { x: cx + r * Math.cos(angle), y: cy + r * Math.sin(angle) }
}

const dimLayout = computed(() => {
  const nodes = props.dimensionNodes.filter((d) => d.count > 0)
  const total = Math.max(nodes.length, 1)
  return nodes.map((dim, i) => ({
    ...dim,
    ...polar(CX, CY, 130, i, total),
  }))
})

const knowledgeLayout = computed(() => {
  const items = props.knowledge.slice(0, 6)
  const n = Math.max(items.length, 1)
  return items.map((k, i) => ({
    ...k,
    ...polar(CX, CY, 210, i, n),
    inner: polar(CX, CY, 130, i, n),
  }))
})

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}
</script>

<template>
  <div class="graph-wrap">
    <p class="graph-hint">
      {{
        hint ??
        `以${centerSubtitle}为中心，聚合关联攻击知识及各维度要素（Demo 示意，非图数据库）。`
      }}
    </p>
    <svg class="graph-svg" viewBox="0 0 720 380" xmlns="http://www.w3.org/2000/svg">
      <circle :cx="CX" :cy="CY" r="52" class="node-center" />
      <text :x="CX" :y="CY - 6" text-anchor="middle" class="label-center">{{ centerLabel }}</text>
      <text :x="CX" :y="CY + 12" text-anchor="middle" class="label-sub">{{ centerSubtitle }}</text>

      <g v-for="dim in dimLayout" :key="dim.key">
        <line :x1="CX" :y1="CY" :x2="dim.x" :y2="dim.y" class="edge-dim" />
        <circle :cx="dim.x" :cy="dim.y" r="28" :fill="dim.color" opacity="0.92" />
        <text :x="dim.x" :y="dim.y + 4" text-anchor="middle" class="label-dim">{{ dim.count }}</text>
        <text :x="dim.x" :y="dim.y + 44" text-anchor="middle" class="label-dim-name">{{ dim.label }}</text>
      </g>

      <g v-for="k in knowledgeLayout" :key="k.knowledgeId">
        <line :x1="k.inner.x" :y1="k.inner.y" :x2="k.x" :y2="k.y" class="edge-know" />
        <g class="know-node" @click="goKnowledge(k.knowledgeId)">
          <rect :x="k.x - 56" :y="k.y - 18" width="112" height="36" rx="6" class="node-knowledge" />
          <text :x="k.x" :y="k.y + 4" text-anchor="middle" class="label-know">#{{ k.knowledgeId }}</text>
        </g>
      </g>
    </svg>

    <ul class="legend">
      <li v-for="dim in dimensionNodes" :key="dim.key">
        <span class="dot" :style="{ background: dim.color }" />
        {{ dim.label }} {{ dim.count }}
      </li>
    </ul>
  </div>
</template>

<style scoped>
.graph-wrap {
  padding: 8px 0;
}
.graph-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #909399;
}
.graph-svg {
  width: 100%;
  max-height: 400px;
  display: block;
  background: linear-gradient(180deg, #f8fafc 0%, #fff 100%);
  border-radius: 12px;
  border: 1px solid var(--ami-border, #ebeef5);
}
.node-center {
  fill: #1e293b;
}
.label-center {
  fill: #fff;
  font-size: 11px;
  font-weight: 700;
}
.label-sub {
  fill: #cbd5e1;
  font-size: 9px;
}
.edge-dim {
  stroke: #cbd5e1;
  stroke-width: 1;
  stroke-dasharray: 4 3;
}
.edge-know {
  stroke: #93c5fd;
  stroke-width: 1.5;
}
.label-dim {
  fill: #fff;
  font-size: 12px;
  font-weight: 700;
}
.label-dim-name {
  fill: #64748b;
  font-size: 9px;
}
.know-node {
  cursor: pointer;
}
.node-knowledge {
  fill: #eff6ff;
  stroke: #3b82f6;
  stroke-width: 1.5;
}
.know-node:hover .node-knowledge {
  fill: #dbeafe;
}
.label-know {
  fill: #1d4ed8;
  font-size: 11px;
  font-weight: 600;
  pointer-events: none;
}
.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
  font-size: 12px;
  color: #606266;
}
.dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}
</style>
