<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { IntelFigure } from '@/types/intel'
import { INTEL_FIGURE_KIND_LABELS } from '@/utils/intel-figures'

const props = defineProps<{
  figures: IntelFigure[]
}>()

const router = useRouter()

const previewList = computed(() => props.figures.map((f) => f.src))

function kindLabel(kind: IntelFigure['kind']): string {
  return INTEL_FIGURE_KIND_LABELS[kind]
}

function goIntel(id: number) {
  router.push({ name: 'intel-detail', params: { id: String(id) } })
}

</script>

<template>
  <div v-if="figures.length" class="gallery">
    <p class="gallery-hint">
      以下图示来自支撑情报原文（攻击流程、ATT&amp;CK 映射、报告截图等），便于对照结构化 TTP 字段。
    </p>
    <div class="gallery-grid">
      <figure v-for="item in figures" :key="item.id" class="figure-card">
        <div class="figure-media">
          <el-image
            :src="item.src"
            :alt="item.title"
            fit="contain"
            class="figure-img"
            loading="lazy"
            :preview-src-list="previewList"
            :initial-index="figures.findIndex((f) => f.id === item.id)"
            preview-teleported
          />
        </div>
        <figcaption class="figure-caption">
          <div class="caption-title">{{ item.title }}</div>
          <div class="caption-meta">
            <el-tag size="small" effect="plain">{{ kindLabel(item.kind) }}</el-tag>
            <el-link type="primary" @click="goIntel(item.intelId)">情报 #{{ item.intelId }}</el-link>
            <span v-if="item.techniqueIds?.length" class="tech-ids">
              {{ item.techniqueIds.join(' · ') }}
            </span>
          </div>
          <p v-if="item.caption && item.caption !== item.title" class="caption-text">
            {{ item.caption }}
          </p>
        </figcaption>
      </figure>
    </div>
  </div>
  <el-empty v-else description="暂无报告图示（支撑情报未包含可解析图片）" :image-size="72" />
</template>

<style scoped>
.gallery-hint {
  margin: 0 0 16px;
  font-size: 13px;
  color: #909399;
  line-height: 1.6;
}
.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}
.figure-card {
  margin: 0;
  border: 1px solid var(--ami-border, #e2e8f0);
  border-radius: 10px;
  overflow: hidden;
  background: #fff;
}
.figure-media {
  cursor: zoom-in;
  background: #f8fafc;
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
}
.figure-img {
  width: 100%;
  max-height: 280px;
}
.figure-caption {
  padding: 12px 14px;
  border-top: 1px solid var(--ami-border, #e2e8f0);
}
.caption-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}
.caption-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.tech-ids {
  color: #606266;
  font-family: ui-monospace, monospace;
}
.caption-text {
  margin: 8px 0 0;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}
</style>
