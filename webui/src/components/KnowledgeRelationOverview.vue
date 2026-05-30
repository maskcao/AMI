<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { malwareClassLabel } from '@/constants/malware'
import type { KnowledgeDetail } from '@/types/knowledge'

const props = defineProps<{
  detail: KnowledgeDetail
}>()

const router = useRouter()

const nodes = computed(() => {
  const d = props.detail
  const list: { type: string; label: string; count: number; items: string[] }[] = [
    {
      type: 'technique',
      label: '攻击手法',
      count: d.techniques.length,
      items: d.techniques.map((t) => `${t.techniqueId} ${t.name}`),
    },
    {
      type: 'tool',
      label: '攻击工具',
      count: d.tools.length,
      items: d.tools.map((t) => t.name),
    },
    {
      type: 'infra',
      label: '基础设施',
      count: d.infrastructure.length,
      items: d.infrastructure.map((i) => `${i.type}: ${i.value}`),
    },
    {
      type: 'cve',
      label: '漏洞',
      count: d.vulnerabilities.length,
      items: d.vulnerabilities.map((v) => v.cveId),
    },
    {
      type: 'malware',
      label: '恶意软件',
      count: d.malware.length,
      items: d.malware.map((m) => `${m.name}（${malwareClassLabel(m.class)}）`),
    },
    {
      type: 'actor',
      label: '威胁主体',
      count: d.threatActors.length,
      items: d.threatActors,
    },
    {
      type: 'intel',
      label: '支撑情报',
      count: d.linkedIntel.length,
      items: d.linkedIntel.map((i) => `#${i.id} ${i.title.slice(0, 40)}…`),
    },
  ]
  return list
})

function goIntel(id: number) {
  router.push({ name: 'intel-detail', params: { id: String(id) } })
}
</script>

<template>
  <p class="hint">同一攻击知识下的多维要素关联视图（Demo 为列表聚合，非图数据库）。</p>
  <el-row :gutter="16">
    <el-col v-for="node in nodes" :key="node.type" :xs="24" :sm="12" :md="8">
      <el-card shadow="hover" class="rel-card">
        <template #header>
          <span>{{ node.label }}</span>
          <el-badge :value="node.count" class="badge" />
        </template>
        <ul v-if="node.items.length" class="item-list">
          <li v-for="(item, idx) in node.items" :key="idx">{{ item }}</li>
        </ul>
        <el-empty v-else :image-size="48" description="暂无" />
      </el-card>
    </el-col>
  </el-row>

  <h4 class="sub-title">情报 → 知识 挂载关系</h4>
  <el-timeline>
    <el-timeline-item
      v-for="intel in detail.linkedIntel"
      :key="intel.id"
      :type="intel.role === 'primary' ? 'primary' : 'info'"
    >
      <el-link type="primary" @click="goIntel(intel.id)">情报 #{{ intel.id }}</el-link>
      <el-tag size="small" class="role-tag">{{ intel.role === 'primary' ? '主要' : '补充' }}</el-tag>
      <div class="intel-title">{{ intel.title }}</div>
    </el-timeline-item>
  </el-timeline>
</template>

<style scoped>
.hint {
  margin: 0 0 16px;
  font-size: 13px;
  color: #909399;
}
.rel-card {
  margin-bottom: 16px;
  min-height: 140px;
}
.badge {
  float: right;
}
.item-list {
  margin: 0;
  padding-left: 18px;
  font-size: 13px;
  color: #606266;
  line-height: 1.7;
}
.sub-title {
  margin: 24px 0 12px;
  font-size: 14px;
  font-weight: 600;
}
.role-tag {
  margin-left: 8px;
}
.intel-title {
  margin-top: 4px;
  font-size: 13px;
  color: #909399;
}
</style>
