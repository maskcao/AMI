<script setup lang="ts">
import type { AttackTechnique } from '@/types/intel'

defineProps<{
  techniques: AttackTechnique[]
  matrix: string
}>()
</script>

<template>
  <div v-if="!techniques.length" class="empty-block">
    <el-empty description="暂无攻击手法标注" :image-size="64" />
  </div>
  <el-table v-else :data="techniques" stripe size="small" class="technique-table">
    <el-table-column label="战术" prop="tactic" width="160" />
    <el-table-column label="技术 ID" prop="techniqueId" width="100" />
    <el-table-column label="名称" prop="name" min-width="180" />
    <el-table-column label="说明" prop="description" min-width="240" show-overflow-tooltip />
  </el-table>
  <p v-if="techniques.length" class="matrix-hint">当前矩阵：{{ matrix }} ATT&CK</p>
</template>

<style scoped>
.empty-block {
  padding: 12px 0;
}
.matrix-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
