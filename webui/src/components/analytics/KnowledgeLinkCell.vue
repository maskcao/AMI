<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { LinkedKnowledgeRef } from '@/analytics/aggregate'

const props = withDefaults(
  defineProps<{
    items: LinkedKnowledgeRef[]
    /** Max links shown inline before collapsing into +N */
    maxInline?: number
  }>(),
  {
    maxInline: 2,
  },
)

const router = useRouter()

const inlineItems = computed(() => props.items.slice(0, props.maxInline))
const overflowItems = computed(() => props.items.slice(props.maxInline))
const overflowCount = computed(() => overflowItems.value.length)

function go(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}
</script>

<template>
  <span v-if="!items.length" class="muted">-</span>
  <span v-else class="knowledge-links">
    <el-tooltip
      v-for="k in inlineItems"
      :key="k.knowledgeId"
      :content="k.knowledgeTitle"
      placement="top"
      :show-after="300"
    >
      <el-link type="primary" class="link-item" @click="go(k.knowledgeId)">
        #{{ k.knowledgeId }}
      </el-link>
    </el-tooltip>

    <el-popover
      v-if="overflowCount > 0"
      placement="top-start"
      :width="320"
      trigger="click"
      popper-class="knowledge-link-popover"
    >
      <template #reference>
        <el-button type="primary" link class="more-btn">+{{ overflowCount }}</el-button>
      </template>
      <div class="popover-head">关联攻击知识（{{ items.length }}）</div>
      <ul class="popover-list">
        <li v-for="k in items" :key="k.knowledgeId">
          <el-link type="primary" @click="go(k.knowledgeId)">
            #{{ k.knowledgeId }}
          </el-link>
          <span class="popover-title" :title="k.knowledgeTitle">{{ k.knowledgeTitle }}</span>
        </li>
      </ul>
    </el-popover>
  </span>
</template>

<style scoped>
.knowledge-links {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 8px;
  max-width: 100%;
}
.link-item {
  font-size: 12px;
}
.more-btn {
  font-size: 12px;
  padding: 0;
  height: auto;
  font-weight: 600;
}
.muted {
  color: #c0c4cc;
}
.popover-head {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--ami-text);
}
.popover-list {
  margin: 0;
  padding: 0;
  list-style: none;
  max-height: 240px;
  overflow-y: auto;
}
.popover-list li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 6px 0;
  border-bottom: 1px solid var(--ami-border);
}
.popover-list li:last-child {
  border-bottom: none;
}
.popover-title {
  flex: 1;
  font-size: 12px;
  color: var(--ami-text-secondary);
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>

<style>
.knowledge-link-popover {
  padding: 12px 14px !important;
}
</style>
