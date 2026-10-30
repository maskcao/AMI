<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { DashboardPayload } from '@/analytics/aggregate'
import { buildActorRows, buildCveRows, buildDashboardPayload } from '@/analytics/aggregate'
import { getAnalyticsHydrateError, hydrateAnalytics, loadDashboardPayload } from '@/api/analytics'
import { USE_API } from '@/api/http'
import { tacticLabel } from '@/constants/mitre-zh'
import SimpleBarChart from '@/components/analytics/SimpleBarChart.vue'
import StatCard from '@/components/analytics/StatCard.vue'

const router = useRouter()
const loading = ref(USE_API)
const error = ref('')
const dashboard = ref<DashboardPayload | null>(null)

const ops = () => dashboard.value?.operations

const tacticDistZh = computed(() =>
  (dashboard.value?.snapshot.tacticDist ?? []).map((item) => ({
    ...item,
    name: tacticLabel(item.name),
  })),
)

/** Top 10 threat actors by linked knowledge count. */
const topActiveActorDist = computed(() => {
  const snap = dashboard.value?.snapshot.activeActorDist
  if (snap?.length) return snap
  return buildActorRows()
    .map((a) => ({ name: a.name, count: a.knowledgeCount }))
    .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name, 'zh-CN'))
    .slice(0, 10)
})

/** In-the-wild CVEs ranked by linked knowledge count (dashboard card). */
const topInWildCveRows = computed(() =>
  buildCveRows()
    .filter((c) => c.inWild)
    .sort(
      (a, b) =>
        b.knowledgeCount - a.knowledgeCount || a.cveId.localeCompare(b.cveId),
    )
    .slice(0, 10),
)

onMounted(async () => {
  if (!USE_API) {
    dashboard.value = buildDashboardPayload()
    loading.value = false
    return
  }
  try {
    dashboard.value = await loadDashboardPayload()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载运营总览失败'
    const cached = getAnalyticsHydrateError()
    if (cached) error.value = `${error.value}（${cached}）`
  } finally {
    loading.value = false
  }
})

async function retry() {
  loading.value = true
  error.value = ''
  try {
    await hydrateAnalytics(true)
    dashboard.value = await loadDashboardPayload()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载运营总览失败'
  } finally {
    loading.value = false
  }
}

function goVulnExploit() {
  router.push({ path: '/analytics/cve' })
}

function goCveDetail(slug: string) {
  router.push({ name: 'entity-dimension-detail', params: { dimension: 'cve', entityKey: slug } })
}
</script>

<template>
  <div v-loading="loading" class="dashboard">
    <el-alert
      v-if="error"
      type="error"
      :title="error"
      show-icon
      :closable="false"
      class="load-error"
    >
      <template #default>
        <p>请确认后端已启动（<code>cd service && make run</code>）且 Vite 代理指向 8080。</p>
        <el-button type="primary" size="small" @click="retry">重试</el-button>
      </template>
    </el-alert>

    <template v-else-if="dashboard && ops()">
      <header class="page-intro">
        <h2 class="intro-heading">
          <span class="intro-title">运营总览</span>
          <span class="intro-sub">
            聚焦知识沉淀与情报采集运营指标；实体维度明细请使用左侧「威胁情报库」导航。
          </span>
        </h2>
      </header>

      <div class="kpi-grid">
        <StatCard
          label="攻击知识条目"
          :value="ops()!.knowledgeCount"
          accent="#3b82f6"
          hint="结构化认知（TTP / 工具 / 设施等）"
        />
        <StatCard
          label="情报源"
          :value="ops()!.intelSourceCount"
          accent="#8b5cf6"
          :hint="ops()!.sourceLabels.join(' · ') || '—'"
        />
        <StatCard
          label="情报采集量"
          :value="ops()!.intelTotalCount"
          accent="#06b6d4"
          :hint="`已完成 ${ops()!.intelSucceededCount} · 处理中 ${ops()!.intelProcessingCount}`"
        />
        <StatCard
          label="流水线失败"
          :value="ops()!.intelFailedCount"
          accent="#ef4444"
          :hint="`已写入知识 ${ops()!.intelLinkedCount} 条`"
        />
      </div>

      <div
        v-if="dashboard.ransomwareKnowledgeCount || dashboard.icsKnowledgeCount"
        class="insight-tags"
      >
        <el-tag v-if="dashboard.ransomwareKnowledgeCount" type="danger" effect="plain" round>
          含勒索软件知识 {{ dashboard.ransomwareKnowledgeCount }} 条
        </el-tag>
        <el-tag v-if="dashboard.icsKnowledgeCount" type="warning" effect="plain" round>
          ICS 矩阵知识 {{ dashboard.icsKnowledgeCount }} 条
        </el-tag>
        <el-tag type="info" effect="plain" round class="click-tag" @click="goVulnExploit">
          在野漏洞利用 {{ dashboard.inWildCveCount }} / {{ dashboard.snapshot.cveCount }} 个 CVE
        </el-tag>
      </div>

      <el-row :gutter="16" class="chart-row">
        <el-col :span="12">
          <el-card shadow="never" class="chart-card">
            <template #header>活跃攻击组织 Top 10</template>
            <SimpleBarChart v-if="topActiveActorDist.length" :items="topActiveActorDist" />
            <el-empty v-else description="暂无攻击组织数据" :image-size="56" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never" class="chart-card">
            <template #header>
              <span>恶意软件分类</span>
              <el-tag size="small" type="danger" effect="plain" class="card-tag">知识聚合</el-tag>
            </template>
            <SimpleBarChart :items="dashboard.snapshot.malwareClassDist" />
          </el-card>
        </el-col>
      </el-row>
      <el-row :gutter="16" class="chart-row">
        <el-col :span="12">
          <el-card shadow="never" class="chart-card">
            <template #header>
              <div class="table-header">
                <span>漏洞利用（在野）</span>
                <el-button text type="primary" @click="goVulnExploit">查看列表</el-button>
              </div>
            </template>
            <el-table
              v-if="topInWildCveRows.length"
              :data="topInWildCveRows"
              stripe
              size="small"
              class="cve-top-table"
            >
              <el-table-column label="CVE" width="148">
                <template #default="{ row }">
                  <el-link type="primary" @click="goCveDetail(row.slug)">{{ row.cveId }}</el-link>
                </template>
              </el-table-column>
              <el-table-column
                prop="description"
                label="描述"
                min-width="120"
                show-overflow-tooltip
              />
              <el-table-column label="关联知识" width="88" align="center">
                <template #default="{ row }">
                  <el-tag size="small" round effect="plain" type="danger">{{ row.knowledgeCount }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty
              v-else
              description="暂无在野 CVE"
              :image-size="56"
            />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never" class="chart-card">
            <template #header>ATT&CK 战术分布（中文）</template>
            <SimpleBarChart :items="tacticDistZh" />
          </el-card>
        </el-col>
      </el-row>
    </template>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-height: 200px;
}
.load-error {
  margin-bottom: 16px;
}
.load-error code {
  font-size: 12px;
}
.page-intro {
  margin-bottom: 20px;
}
.intro-heading {
  margin: 0;
  font-weight: inherit;
  font-size: inherit;
  line-height: 1.5;
}
.intro-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--ami-text);
  letter-spacing: -0.02em;
}
.intro-sub {
  margin-left: 12px;
  font-size: 13px;
  font-weight: 400;
  color: var(--ami-text-secondary);
}
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}
@media (max-width: 1100px) {
  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
.insight-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 20px;
}
.chart-row {
  margin-bottom: 16px;
}
.chart-card {
  background: var(--ami-surface) !important;
  height: 100%;
}
.card-tag {
  margin-left: 8px;
}
.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
.click-tag {
  cursor: pointer;
}
.cve-top-table {
  width: 100%;
}
</style>
