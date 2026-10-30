<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  buildActorRows,
  buildCveRows,
  buildInfraRows,
  buildMalwareRows,
  buildTechniqueRows,
  buildToolRows,
  type CveRow,
  type MalwareRow,
  type ActorRow,
  type TechniqueRow,
  type ToolRow,
  type InfraRow,
} from '@/analytics/aggregate'
import { THREAT_LIB_NAV_ITEMS } from '@/constants/threat-lib-nav'
import AppLayout from '@/components/AppLayout.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import EntityListPanel from '@/components/analytics/EntityListPanel.vue'
import MitreTacticTree from '@/components/analytics/MitreTacticTree.vue'
import ThreatIntelFilterForm, {
  type ThreatIntelFilterDraft,
} from '@/components/analytics/ThreatIntelFilterForm.vue'
import IndustryTags from '@/components/IndustryTags.vue'
import KnowledgeLinkCell from '@/components/analytics/KnowledgeLinkCell.vue'
import { actorNameToSlug } from '@/analytics/actors'
import type { EntityDimensionKind } from '@/types/entity-profile'
import { includesQuery, usePagedFilter } from '@/composables/usePagedFilter'
import { malwareClassLabel, malwareClassColor } from '@/constants/malware'

const route = useRoute()
const router = useRouter()

const dimension = computed(() => route.params.dimension as string)
const nav = THREAT_LIB_NAV_ITEMS
const activeNav = computed(() => nav.find((n) => n.key === dimension.value) ?? nav[0])

const infraRoleLabel: Record<string, string> = {
  c2: 'C2',
  hosting: '托管',
  payload: '载荷分发',
  scan: '扫描源',
  other: '其他',
}

function emptyDraft(): ThreatIntelFilterDraft {
  return {
    query: '',
    facet: '',
    inWild: '',
    matrix: '',
    malwareClass: '',
    infraRole: '',
    toolCategory: '',
    tactic: '',
  }
}

const draft = reactive<ThreatIntelFilterDraft>(emptyDraft())

const cveList = usePagedFilter<CveRow>(() => (dimension.value === 'cve' ? buildCveRows() : []), {
  matchQuery: (row, q) =>
    includesQuery(q, row.cveId, row.description) ||
    row.targetIndustries.some((ind) => includesQuery(q, ind)) ||
    row.knowledge.some((k) => includesQuery(q, k.knowledgeTitle, String(k.knowledgeId))),
  matchExtras: (row, ex) => {
    if (ex.inWild === 'yes') return row.inWild
    if (ex.inWild === 'no') return !row.inWild
    return true
  },
})

const malwareList = usePagedFilter<MalwareRow>(() => (dimension.value === 'malware' ? buildMalwareRows() : []), {
  matchQuery: (row, q) =>
    includesQuery(q, row.name, malwareClassLabel(row.class)) ||
    row.targetIndustries.some((ind) => includesQuery(q, ind)) ||
    row.knowledge.some((k) => includesQuery(q, k.knowledgeTitle, String(k.knowledgeId))),
  matchFacet: (row, facet) => !facet || row.class === facet,
})

const actorList = usePagedFilter<ActorRow>(() => (dimension.value === 'actors' ? buildActorRows() : []), {
  matchQuery: (row, q) =>
    includesQuery(q, row.name) ||
    row.targetIndustries.some((ind) => includesQuery(q, ind)) ||
    row.knowledge.some((k) => includesQuery(q, k.knowledgeTitle, String(k.knowledgeId))),
})

function goActorDetail(name: string) {
  router.push({ name: 'threat-actor-detail', params: { actorKey: actorNameToSlug(name) } })
}

function goEntityDetail(dimension: EntityDimensionKind, slug: string) {
  router.push({ name: 'entity-dimension-detail', params: { dimension, entityKey: slug } })
}

const mitreList = usePagedFilter<TechniqueRow>(
  () => (dimension.value === 'mitre' ? buildTechniqueRows() : []),
  {
  matchQuery: (row, q) =>
    includesQuery(q, row.techniqueId, row.name, row.tactic, row.matrix) ||
    row.knowledge.some((k) => includesQuery(q, k.knowledgeTitle, String(k.knowledgeId))),
  matchFacet: (row, facet) => {
    if (!facet) return true
    if (facet.startsWith('tactic:')) return row.tactic === facet.slice(7)
    if (facet.startsWith('tech:')) return row.techniqueId === facet.slice(5)
    return true
  },
  matchExtras: (row, ex) => {
    if (ex.matrix && row.matrix !== ex.matrix) return false
    if (ex.tactic && row.tactic !== ex.tactic) return false
    return true
  },
})

const toolList = usePagedFilter<ToolRow>(() => (dimension.value === 'tools' ? buildToolRows() : []), {
  matchQuery: (row, q) =>
    includesQuery(q, row.name, row.category) ||
    row.targetIndustries.some((ind) => includesQuery(q, ind)) ||
    row.knowledge.some((k) => includesQuery(q, k.knowledgeTitle, String(k.knowledgeId))),
  matchExtras: (row, ex) => !ex.category || row.category === ex.category,
})

const infraList = usePagedFilter<InfraRow>(() => (dimension.value === 'infra' ? buildInfraRows() : []), {
  matchQuery: (row, q) =>
    includesQuery(q, row.type, row.value, row.role, infraRoleLabel[row.role]) ||
    row.targetIndustries.some((ind) => includesQuery(q, ind)) ||
    row.knowledge.some((k) => includesQuery(q, k.knowledgeTitle, String(k.knowledgeId))),
  matchFacet: (row, facet) => !facet || row.role === facet,
})

const tacticOptions = computed(() => {
  if (dimension.value !== 'mitre') return []
  return [...new Set(buildTechniqueRows().map((t) => t.tactic))].sort((a, b) =>
    a.localeCompare(b),
  )
})

const toolCategoryOptions = computed(() => {
  if (dimension.value !== 'tools') return []
  return [...new Set(buildToolRows().map((t) => t.category).filter(Boolean) as string[])].sort(
    (a, b) => a.localeCompare(b),
  )
})

const infraRoleOptions = computed(() => {
  if (dimension.value !== 'infra') return []
  const keys = [...new Set(buildInfraRows().map((i) => i.role))]
  return keys.map((key) => ({ key, label: infraRoleLabel[key] ?? key }))
})

const mitreTechniqueRows = computed(() =>
  dimension.value === 'mitre' ? buildTechniqueRows() : [],
)

function applySearch() {
  const q = draft.query
  switch (dimension.value) {
    case 'cve':
      cveList.applyFilters({
        query: q,
        facet: '',
        extras: draft.inWild ? { inWild: draft.inWild } : {},
      })
      break
    case 'malware':
      malwareList.applyFilters({
        query: q,
        facet: draft.malwareClass,
        extras: {},
      })
      break
    case 'actors':
      actorList.applyFilters({ query: q, facet: '', extras: {} })
      break
    case 'mitre': {
      const extras: Record<string, string> = {}
      if (draft.matrix) extras.matrix = draft.matrix
      if (draft.tactic) extras.tactic = draft.tactic
      mitreList.applyFilters({
        query: q,
        facet: draft.tactic ? '' : draft.facet,
        extras,
      })
      break
    }
    case 'tools':
      toolList.applyFilters({
        query: q,
        facet: '',
        extras: draft.toolCategory ? { category: draft.toolCategory } : {},
      })
      break
    case 'infra':
      infraList.applyFilters({
        query: q,
        facet: draft.infraRole,
        extras: {},
      })
      break
  }
}

function onMitreTreePick(facetKey: string) {
  draft.facet = facetKey
  draft.tactic = ''
  applySearch()
}

function resetDimensionState() {
  Object.assign(draft, emptyDraft())
  cveList.resetFilters()
  malwareList.resetFilters()
  actorList.resetFilters()
  mitreList.resetFilters()
  toolList.resetFilters()
  infraList.resetFilters()
}

watch(
  dimension,
  () => {
    resetDimensionState()
    applySearch()
  },
  { immediate: true },
)
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        :title="activeNav?.label ?? '威胁情报库'"
        :description="
          dimension === 'actors'
            ? '攻击组织由关联的攻击知识聚合；点击组织名查看多维关联图谱。'
            : dimension === 'cve'
              ? '漏洞利用维度：从攻击知识聚合 CVE 及在野状态；点击 CVE 查看关联行业、组织、工具等与图谱。'
              : ['malware', 'tools', 'infra'].includes(dimension)
                ? '由攻击知识聚合；点击名称查看关联行业、知识条目与多维拓扑图谱。'
                : '按实体类型浏览聚合威胁对象；筛选条件填写后点击「查询」生效（与攻击知识库一致）。'
        "
      />

      <ThreatIntelFilterForm
        :dimension="dimension"
        :draft="draft"
        :tactic-options="tacticOptions"
        :tool-category-options="toolCategoryOptions"
        :infra-role-options="infraRoleOptions"
        @search="applySearch"
      />

      <section class="panel">
        <el-alert
          v-if="dimension === 'infra'"
          type="info"
          :closable="false"
          show-icon
          class="infra-hint"
          title="本页展示攻击知识中维护的基础设施（域名、IP 等及用途）。报告内抽取的 IOC 请在采集队列对应任务中查看。"
        />

        <EntityListPanel
          v-if="dimension === 'cve'"
          title="漏洞利用"
          :total="cveList.total.value"
          :page="cveList.page.value"
          :page-size="cveList.pageSize.value"
          @page-change="cveList.onPageChange"
          @size-change="cveList.onSizeChange"
        >
          <el-table v-if="cveList.total.value > 0" :data="cveList.pageData.value" stripe size="small">
            <el-table-column label="CVE" width="150">
              <template #default="{ row }">
                <el-link type="primary" @click="goEntityDetail('cve', row.slug)">{{ row.cveId }}</el-link>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
            <el-table-column label="在野" width="80">
              <template #default="{ row }">
                <el-tag :type="row.inWild ? 'danger' : 'info'" size="small">
                  {{ row.inWild ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="攻击行业" min-width="160">
              <template #default="{ row }">
                <IndustryTags v-if="row.targetIndustries.length" :industries="row.targetIndustries" />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="关联知识" width="96" align="center">
              <template #default="{ row }">
                <el-tag size="small" round effect="plain" type="primary">{{ row.knowledgeCount }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="88" align="center">
              <template #default="{ row }">
                <el-button text type="primary" size="small" @click="goEntityDetail('cve', row.slug)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </EntityListPanel>

        <EntityListPanel
          v-else-if="dimension === 'malware'"
          title="恶意软件家族"
          :total="malwareList.total.value"
          :page="malwareList.page.value"
          :page-size="malwareList.pageSize.value"
          @page-change="malwareList.onPageChange"
          @size-change="malwareList.onSizeChange"
        >
          <el-table v-if="malwareList.total.value > 0" :data="malwareList.pageData.value" stripe size="small">
            <el-table-column label="家族" width="160">
              <template #default="{ row }">
                <el-link type="primary" @click="goEntityDetail('malware', row.slug)">{{ row.name }}</el-link>
              </template>
            </el-table-column>
            <el-table-column label="分类" width="110">
              <template #default="{ row }">
                <el-tag
                  v-if="row.class"
                  size="small"
                  :type="malwareClassColor(row.class)"
                >
                  {{ malwareClassLabel(row.class) }}
                </el-tag>
                <span v-else class="muted">未分类</span>
              </template>
            </el-table-column>
            <el-table-column label="攻击行业" min-width="160">
              <template #default="{ row }">
                <IndustryTags v-if="row.targetIndustries.length" :industries="row.targetIndustries" />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="关联知识" width="96" align="center">
              <template #default="{ row }">
                <el-tag size="small" round effect="plain" type="primary">{{ row.knowledgeCount }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="88" align="center">
              <template #default="{ row }">
                <el-button text type="primary" size="small" @click="goEntityDetail('malware', row.slug)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </EntityListPanel>

        <EntityListPanel
          v-else-if="dimension === 'actors'"
          title="攻击组织"
          :total="actorList.total.value"
          :page="actorList.page.value"
          :page-size="actorList.pageSize.value"
          @page-change="actorList.onPageChange"
          @size-change="actorList.onSizeChange"
        >
          <el-table v-if="actorList.total.value > 0" :data="actorList.pageData.value" stripe size="small">
            <el-table-column label="组织/团伙" min-width="180">
              <template #default="{ row }">
                <el-link type="primary" class="actor-link" @click="goActorDetail(row.name)">
                  {{ row.name }}
                </el-link>
              </template>
            </el-table-column>
            <el-table-column label="攻击行业" min-width="200">
              <template #default="{ row }">
                <IndustryTags v-if="row.targetIndustries.length" :industries="row.targetIndustries" />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="关联知识" width="96" align="center">
              <template #default="{ row }">
                <el-tag size="small" round effect="plain" type="primary">{{ row.knowledgeCount }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="88" align="center">
              <template #default="{ row }">
                <el-button text type="primary" size="small" @click="goActorDetail(row.name)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>
        </EntityListPanel>

        <EntityListPanel
          v-else-if="dimension === 'mitre'"
          title="MITRE ATT&CK 技术"
          :total="mitreList.total.value"
          :page="mitreList.page.value"
          :page-size="mitreList.pageSize.value"
          @page-change="mitreList.onPageChange"
          @size-change="mitreList.onSizeChange"
        >
          <template #aside>
            <MitreTacticTree
              :model-value="draft.facet"
              :techniques="mitreTechniqueRows"
              @update:model-value="onMitreTreePick"
            />
          </template>
          <el-table v-if="mitreList.total.value > 0" :data="mitreList.pageData.value" stripe size="small">
            <el-table-column prop="techniqueId" label="ID" width="90" />
            <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
            <el-table-column prop="tactic" label="战术" width="140" />
            <el-table-column prop="matrix" label="矩阵" width="100" />
            <el-table-column label="知识" width="120">
              <template #default="{ row }">
                <KnowledgeLinkCell :items="row.knowledge" />
              </template>
            </el-table-column>
          </el-table>
        </EntityListPanel>

        <EntityListPanel
          v-else-if="dimension === 'tools'"
          title="攻击工具"
          :total="toolList.total.value"
          :page="toolList.page.value"
          :page-size="toolList.pageSize.value"
          @page-change="toolList.onPageChange"
          @size-change="toolList.onSizeChange"
        >
          <el-table v-if="toolList.total.value > 0" :data="toolList.pageData.value" stripe size="small">
            <el-table-column label="工具" width="160">
              <template #default="{ row }">
                <el-link type="primary" @click="goEntityDetail('tool', row.slug)">{{ row.name }}</el-link>
              </template>
            </el-table-column>
            <el-table-column prop="category" label="类别" width="120" />
            <el-table-column label="攻击行业" min-width="160">
              <template #default="{ row }">
                <IndustryTags v-if="row.targetIndustries.length" :industries="row.targetIndustries" />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="关联知识" width="96" align="center">
              <template #default="{ row }">
                <el-tag size="small" round effect="plain" type="primary">{{ row.knowledgeCount }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="88" align="center">
              <template #default="{ row }">
                <el-button text type="primary" size="small" @click="goEntityDetail('tool', row.slug)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </EntityListPanel>

        <EntityListPanel
          v-else-if="dimension === 'infra'"
          title="基础设施"
          :total="infraList.total.value"
          :page="infraList.page.value"
          :page-size="infraList.pageSize.value"
          @page-change="infraList.onPageChange"
          @size-change="infraList.onSizeChange"
        >
          <el-table v-if="infraList.total.value > 0" :data="infraList.pageData.value" stripe size="small">
            <el-table-column prop="type" label="类型" width="80" />
            <el-table-column label="值" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <el-link type="primary" @click="goEntityDetail('infra', row.slug)">{{ row.value }}</el-link>
              </template>
            </el-table-column>
            <el-table-column label="用途" width="100">
              <template #default="{ row }">
                {{ infraRoleLabel[row.role] ?? row.role }}
              </template>
            </el-table-column>
            <el-table-column label="攻击行业" min-width="140">
              <template #default="{ row }">
                <IndustryTags v-if="row.targetIndustries.length" :industries="row.targetIndustries" />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column label="关联知识" width="96" align="center">
              <template #default="{ row }">
                <el-tag size="small" round effect="plain" type="primary">{{ row.knowledgeCount }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="88" align="center">
              <template #default="{ row }">
                <el-button text type="primary" size="small" @click="goEntityDetail('infra', row.slug)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </EntityListPanel>
      </section>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.infra-hint {
  margin-bottom: 16px;
  border-radius: var(--ami-radius);
}
.panel :deep(.el-table) {
  width: 100%;
}
.muted {
  color: #94a3b8;
  font-size: 13px;
}
.actor-link {
  font-weight: 500;
}
</style>
