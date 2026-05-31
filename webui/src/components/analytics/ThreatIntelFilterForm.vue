<script setup lang="ts">
import { computed } from 'vue'
import FilterCard from '@/components/FilterCard.vue'
import { MALWARE_CLASSES, type MalwareClass } from '@/constants/malware'

export interface ThreatIntelFilterDraft {
  query: string
  facet: string
  inWild: '' | 'yes' | 'no'
  matrix: '' | 'Enterprise' | 'ICS'
  malwareClass: MalwareClass | ''
  infraRole: string
  toolCategory: string
  tactic: string
}

const props = defineProps<{
  dimension: string
  draft: ThreatIntelFilterDraft
  tacticOptions: string[]
  toolCategoryOptions: string[]
  infraRoleOptions: { key: string; label: string }[]
}>()

const emit = defineEmits<{
  search: []
}>()

const keywordPlaceholder = computed(() => {
  const map: Record<string, string> = {
    cve: 'CVE / 漏洞描述 / 关联知识',
    malware: '家族名 / 分类 / 关联知识',
    actors: '组织名 / 关联知识',
    mitre: '技术 ID / 名称 / 战术',
    tools: '工具名 / 类别 / 关联知识',
    infra: '类型 / 值 / 用途',
  }
  return map[props.dimension] ?? '关键词'
})

function submit() {
  emit('search')
}
</script>

<template>
  <FilterCard>
    <el-form :inline="true" @submit.prevent="submit">
      <el-form-item label="关键词">
        <el-input v-model="draft.query" clearable :placeholder="keywordPlaceholder" style="width: 220px" />
      </el-form-item>

      <template v-if="dimension === 'cve'">
        <el-form-item label="在野">
          <el-select v-model="draft.inWild" clearable placeholder="全部" style="width: 120px">
            <el-option label="是" value="yes" />
            <el-option label="否" value="no" />
          </el-select>
        </el-form-item>
      </template>

      <template v-else-if="dimension === 'malware'">
        <el-form-item label="恶意软件分类">
          <el-select v-model="draft.malwareClass" clearable placeholder="全部" style="width: 140px">
            <el-option v-for="c in MALWARE_CLASSES" :key="c.value" :label="c.label" :value="c.value" />
          </el-select>
        </el-form-item>
      </template>

      <template v-else-if="dimension === 'mitre'">
        <el-form-item label="战术">
          <el-select v-model="draft.tactic" clearable placeholder="全部" style="width: 160px">
            <el-option v-for="t in tacticOptions" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="ATT&CK">
          <el-select v-model="draft.matrix" clearable placeholder="全部" style="width: 130px">
            <el-option label="Enterprise" value="Enterprise" />
            <el-option label="ICS" value="ICS" />
          </el-select>
        </el-form-item>
      </template>

      <template v-else-if="dimension === 'tools'">
        <el-form-item label="类别">
          <el-select v-model="draft.toolCategory" clearable placeholder="全部" style="width: 140px">
            <el-option v-for="c in toolCategoryOptions" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
      </template>

      <template v-else-if="dimension === 'infra'">
        <el-form-item label="用途">
          <el-select v-model="draft.infraRole" clearable placeholder="全部" style="width: 120px">
            <el-option v-for="r in infraRoleOptions" :key="r.key" :label="r.label" :value="r.key" />
          </el-select>
        </el-form-item>
      </template>

      <el-form-item>
        <el-button type="primary" @click="submit">查询</el-button>
      </el-form-item>
    </el-form>
  </FilterCard>
</template>
