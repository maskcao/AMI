<script setup lang="ts">
import { useRouter } from 'vue-router'
import { buildDashboardPayload } from '@/analytics/aggregate'
import SimpleBarChart from '@/components/analytics/SimpleBarChart.vue'
import StatCard from '@/components/analytics/StatCard.vue'
import QualityScore from '@/components/QualityScore.vue'

const router = useRouter()
const dashboard = buildDashboardPayload()
const ops = dashboard.operations

function goKnowledgeList() {
  router.push({ name: 'knowledge-list' })
}

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}
</script>

<template>
  <div class="dashboard">
    <header class="page-intro">
      <h2 class="intro-heading">
        <span class="intro-title">运营总览</span>
        <span class="intro-sub">
          聚焦知识沉淀与情报采集运营指标；实体维度明细请使用左侧「威胁情报库」导航。
        </span>
      </h2>
    </header>

    <!-- Platform KPIs (no overlap with threat-intel dimension counts) -->
    <div class="kpi-grid">
      <StatCard
        label="攻击知识条目"
        :value="ops.knowledgeCount"
        accent="#3b82f6"
        hint="结构化认知（TTP / 工具 / 设施等）"
      />
      <StatCard
        label="情报源"
        :value="ops.intelSourceCount"
        accent="#8b5cf6"
        :hint="ops.sourceLabels.join(' · ') || '—'"
      />
      <StatCard
        label="情报采集量"
        :value="ops.intelTotalCount"
        accent="#06b6d4"
        :hint="`已发布 ${ops.intelPublishedCount} · 待审 ${ops.intelPendingCount}`"
      />
      <StatCard
        label="已关联知识的情报"
        :value="ops.intelLinkedCount"
        accent="#22c55e"
        :hint="`未关联 ${ops.intelTotalCount - ops.intelLinkedCount} 条`"
      />
    </div>

  <!-- Insight tags (derived, not duplicate counts) -->
    <div v-if="dashboard.ransomwareKnowledgeCount || dashboard.icsKnowledgeCount" class="insight-tags">
      <el-tag v-if="dashboard.ransomwareKnowledgeCount" type="danger" effect="plain" round>
        含勒索软件知识 {{ dashboard.ransomwareKnowledgeCount }} 条
      </el-tag>
      <el-tag v-if="dashboard.icsKnowledgeCount" type="warning" effect="plain" round>
        ICS 矩阵知识 {{ dashboard.icsKnowledgeCount }} 条
      </el-tag>
      <el-tag type="info" effect="plain" round>
        在野 CVE {{ dashboard.inWildCveCount }} / {{ dashboard.snapshot.cveCount }}
      </el-tag>
    </div>

    <!-- Charts -->
    <el-row :gutter="16" class="chart-row">
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <span>恶意软件分类</span>
            <el-tag size="small" type="danger" effect="plain" class="card-tag">知识聚合</el-tag>
          </template>
          <SimpleBarChart :items="dashboard.snapshot.malwareClassDist" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <template #header>攻击行业分布</template>
          <SimpleBarChart :items="dashboard.snapshot.industryDist" />
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" class="chart-row">
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <template #header>ATT&CK 战术分布</template>
          <SimpleBarChart :items="dashboard.snapshot.tacticDist" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <template #header>Enterprise / ICS 矩阵</template>
          <SimpleBarChart :items="dashboard.snapshot.matrixDist" />
        </el-card>
      </el-col>
    </el-row>

    <!-- Recent knowledge activity -->
    <el-card shadow="never" class="chart-card activity-card">
      <template #header>
        <div class="table-header">
          <span>最近更新的攻击知识</span>
          <el-button text type="primary" @click="goKnowledgeList">查看全部</el-button>
        </div>
      </template>
      <el-table :data="dashboard.knowledgeRows" stripe size="small">
        <el-table-column prop="id" label="ID" width="64" />
        <el-table-column label="标题" min-width="280">
          <template #default="{ row }">
            <el-link type="primary" @click="goKnowledge(row.id)">{{ row.title }}</el-link>
            <el-tag v-if="row.hasRansomware" size="small" type="danger" class="row-tag">勒索</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="attackMatrix" label="矩阵" width="100" />
        <el-table-column label="挂载情报" width="88" align="center">
          <template #default="{ row }">{{ row.intelCount }}</template>
        </el-table-column>
        <el-table-column label="质量" width="120">
          <template #default="{ row }">
            <QualityScore :score="row.qualityScore" />
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="120" />
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
.activity-card {
  margin-top: 4px;
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
.row-tag {
  margin-left: 8px;
  vertical-align: middle;
}
</style>
