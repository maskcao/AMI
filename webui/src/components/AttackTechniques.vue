<script setup lang="ts">
import { computed } from 'vue'
import type { AttackTechnique, IntelFigure } from '@/types/intel'
import { figuresForTechnique } from '@/utils/intel-figures'

const props = defineProps<{
  techniques: AttackTechnique[]
  matrix: string
  evidenceFigures?: IntelFigure[]
}>()

const figuresByTechnique = computed(() => {
  const map = new Map<string, IntelFigure[]>()
  if (!props.evidenceFigures?.length) return map
  for (const t of props.techniques) {
    const linked = figuresForTechnique(props.evidenceFigures, t.techniqueId)
    if (linked.length) map.set(t.techniqueId, linked)
  }
  return map
})
</script>

<template>
  <div v-if="!techniques.length" class="empty-block">
    <el-empty description="暂无攻击手法标注" :image-size="64" />
  </div>
  <div v-else class="technique-list">
    <article v-for="row in techniques" :key="row.techniqueId" class="technique-block">
      <header class="technique-head">
        <el-tag size="small" type="info">{{ row.tactic }}</el-tag>
        <span class="technique-id">{{ row.techniqueId }}</span>
        <span class="technique-name">{{ row.name }}</span>
      </header>
      <p v-if="row.description" class="technique-desc">{{ row.description }}</p>
      <div v-if="figuresByTechnique.get(row.techniqueId)?.length" class="technique-figures">
        <div
          v-for="fig in figuresByTechnique.get(row.techniqueId)"
          :key="fig.id"
          class="technique-figure-wrap"
        >
          <el-image
            :src="fig.src"
            :alt="fig.title"
            fit="contain"
            class="technique-figure-img"
            :preview-src-list="figuresByTechnique.get(row.techniqueId)!.map((f) => f.src)"
            preview-teleported
          />
          <span class="technique-figure-label">{{ fig.title }}</span>
        </div>
      </div>
    </article>
  </div>
  <p v-if="techniques.length" class="matrix-hint">当前矩阵：{{ matrix }} ATT&CK</p>
</template>

<style scoped>
.empty-block {
  padding: 12px 0;
}
.technique-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.technique-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.technique-id {
  font-family: ui-monospace, monospace;
  font-size: 13px;
  font-weight: 600;
  color: #3b82f6;
}
.technique-name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}
.technique-desc {
  margin: 0 0 10px;
  font-size: 13px;
  line-height: 1.6;
  color: #606266;
}
.technique-figures {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}
.technique-figure-wrap {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
  background: #f8fafc;
}
.technique-figure-img {
  width: 100%;
  min-height: 160px;
  max-height: 240px;
  cursor: zoom-in;
}
.technique-figure-label {
  display: block;
  padding: 8px 10px;
  font-size: 12px;
  color: #606266;
  border-top: 1px solid #ebeef5;
  background: #fff;
}
.matrix-hint {
  margin-top: 12px;
  font-size: 12px;
  color: #909399;
}
</style>
