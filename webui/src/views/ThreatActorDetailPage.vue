<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getThreatActorProfileBySlug } from '@/analytics/actors'
import AppLayout from '@/components/AppLayout.vue'
import AttackTechniques from '@/components/AttackTechniques.vue'
import IndustryTags from '@/components/IndustryTags.vue'
import MalwareClassTags from '@/components/MalwareClassTags.vue'
import PageShell from '@/components/PageShell.vue'
import ThreatActorGraph from '@/components/analytics/ThreatActorGraph.vue'
import KnowledgeLinkCell from '@/components/analytics/KnowledgeLinkCell.vue'
import { malwareClassLabel } from '@/constants/malware'
import type { ThreatActorProfile } from '@/types/threat-actor'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const activeTab = ref('graph')
const profile = ref<ThreatActorProfile | null>(null)

const actorSlug = computed(() => route.params.actorKey as string)

const infraRoleLabel: Record<string, string> = {
  c2: 'C2',
  hosting: '托管',
  payload: '载荷分发',
  scan: '扫描源',
  other: '其他',
}

async function load() {
  loading.value = true
  try {
    profile.value = getThreatActorProfileBySlug(actorSlug.value)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push({ path: '/analytics/actors' })
}

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}

watch(actorSlug, load)
onMounted(load)
</script>

<template>
  <AppLayout>
    <PageShell v-loading="loading" wide>
      <template v-if="profile">
        <div class="detail-top">
          <el-button text type="primary" @click="goBack">← 返回攻击组织列表</el-button>
          <h1 class="detail-title">{{ profile.name }}</h1>
          <div v-if="profile.aliases.length" class="alias-row">
            <span class="alias-label">别名</span>
            <el-tag v-for="a in profile.aliases" :key="a" size="small" effect="plain" round>{{ a }}</el-tag>
          </div>
          <p v-if="profile.summary" class="summary">{{ profile.summary }}</p>
        </div>

        <el-card shadow="never" class="meta-card">
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="攻击行业" :span="2">
              <IndustryTags :industries="profile.targetIndustries" />
            </el-descriptions-item>
            <el-descriptions-item label="关联知识">
              <strong>{{ profile.knowledgeCount }}</strong> 条
            </el-descriptions-item>
            <el-descriptions-item label="支撑情报">
              {{ profile.intelReportCount }} 份
            </el-descriptions-item>
            <el-descriptions-item label="ATT&CK 矩阵">
              <el-tag v-for="m in profile.matrices" :key="m" size="small" class="matrix-tag">{{ m }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="活动窗口">
              {{ profile.firstSeen?.slice(0, 10) ?? '—' }} ~ {{ profile.lastSeen?.slice(0, 10) ?? '—' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="关联图谱" name="graph">
              <ThreatActorGraph :profile="profile" />
            </el-tab-pane>

            <el-tab-pane label="攻击知识" name="knowledge">
              <el-table :data="profile.knowledge" stripe size="small">
                <el-table-column prop="knowledgeId" label="ID" width="72" />
                <el-table-column label="标题" min-width="320">
                  <template #default="{ row }">
                    <el-link type="primary" @click="goKnowledge(row.knowledgeId)">
                      {{ row.knowledgeTitle }}
                    </el-link>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>

            <el-tab-pane label="攻击手法" name="techniques">
              <AttackTechniques
                :techniques="profile.techniques"
                :matrix="profile.matrices[0] ?? 'Enterprise'"
              />
            </el-tab-pane>

            <el-tab-pane label="攻击工具" name="tools">
              <el-table v-if="profile.tools.length" :data="profile.tools" stripe size="small">
                <el-table-column prop="name" label="工具" min-width="140" />
                <el-table-column prop="category" label="类别" width="120" />
                <el-table-column prop="description" label="说明" min-width="240" show-overflow-tooltip />
                <el-table-column label="来源知识" width="120">
                  <template #default>
                    <KnowledgeLinkCell :items="profile.knowledge" />
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无关联工具" />
            </el-tab-pane>

            <el-tab-pane label="基础设施" name="infra">
              <el-table v-if="profile.infrastructure.length" :data="profile.infrastructure" stripe size="small">
                <el-table-column prop="type" label="类型" width="80" />
                <el-table-column prop="value" label="值" min-width="200" />
                <el-table-column label="用途" width="100">
                  <template #default="{ row }">{{ infraRoleLabel[row.role] ?? row.role }}</template>
                </el-table-column>
                <el-table-column prop="confidence" label="置信度" width="90" />
                <el-table-column prop="note" label="备注" min-width="160" show-overflow-tooltip />
              </el-table>
              <el-empty v-else description="暂无基础设施" />
            </el-tab-pane>

            <el-tab-pane label="恶意软件" name="malware">
              <MalwareClassTags v-if="profile.malware.length" :items="profile.malware" />
              <el-table v-if="profile.malware.length" :data="profile.malware" stripe size="small" class="malware-table">
                <el-table-column prop="name" label="家族" />
                <el-table-column label="分类">
                  <template #default="{ row }">{{ malwareClassLabel(row.class) }}</template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无恶意软件" />
            </el-tab-pane>

            <el-tab-pane label="利用漏洞" name="cve">
              <el-table v-if="profile.vulnerabilities.length" :data="profile.vulnerabilities" stripe size="small">
                <el-table-column prop="cveId" label="CVE" width="140">
                  <template #default="{ row }">
                    <el-link
                      :href="`https://nvd.nist.gov/vuln/detail/${row.cveId}`"
                      target="_blank"
                      type="primary"
                    >
                      {{ row.cveId }}
                    </el-link>
                  </template>
                </el-table-column>
                <el-table-column prop="description" label="描述" min-width="260" show-overflow-tooltip />
                <el-table-column label="在野" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.inWild ? 'danger' : 'info'" size="small">
                      {{ row.inWild ? '是' : '否' }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无关联漏洞" />
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </template>

      <el-result v-else-if="!loading" icon="warning" title="未找到该攻击组织">
        <template #extra>
          <el-button type="primary" @click="goBack">返回列表</el-button>
        </template>
      </el-result>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.detail-top {
  margin-bottom: 16px;
}
.detail-title {
  margin: 8px 0 0;
  font-size: 22px;
  font-weight: 700;
}
.alias-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}
.alias-label {
  font-size: 12px;
  color: #909399;
}
.summary {
  margin: 12px 0 0;
  font-size: 14px;
  line-height: 1.7;
  color: #606266;
}
.meta-card {
  margin-bottom: 16px;
}
.matrix-tag {
  margin-right: 6px;
}
.malware-table {
  margin-top: 16px;
}
</style>
