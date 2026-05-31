<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DIMENSION_LIST_PATH, getEntityProfile } from '@/analytics/entity-profiles'
import { useEntityTopologyNodes } from '@/composables/useEntityTopology'
import AppLayout from '@/components/AppLayout.vue'
import AttackTechniques from '@/components/AttackTechniques.vue'
import IndustryTags from '@/components/IndustryTags.vue'
import MalwareClassTags from '@/components/MalwareClassTags.vue'
import PageShell from '@/components/PageShell.vue'
import EntityTopologyGraph from '@/components/analytics/EntityTopologyGraph.vue'
import { malwareClassLabel, malwareClassColor } from '@/constants/malware'
import type { EntityDimensionKind, EntityDimensionProfile } from '@/types/entity-profile'
import type { TopologyProfileSlice } from '@/composables/useEntityTopology'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const activeTab = ref('graph')
const profile = ref<EntityDimensionProfile | null>(null)

const dimension = computed(() => route.params.dimension as EntityDimensionKind)
const topologySource = computed<TopologyProfileSlice | null>(() => profile.value)
const topologyNodes = useEntityTopologyNodes(topologySource)

const infraRoleLabel: Record<string, string> = {
  c2: 'C2',
  hosting: '托管',
  payload: '载荷分发',
  scan: '扫描源',
  other: '其他',
}

const listPath = computed(() => DIMENSION_LIST_PATH[dimension.value] ?? '/analytics/cve')

async function load() {
  loading.value = true
  try {
    profile.value = getEntityProfile(dimension.value, route.params.entityKey as string)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push(listPath.value)
}

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}

watch([dimension, () => route.params.entityKey], load)
onMounted(load)
</script>

<template>
  <AppLayout>
    <PageShell v-loading="loading" wide>
      <template v-if="profile">
        <div class="detail-top">
          <el-button text type="primary" @click="goBack">← 返回{{ profile.subtitle }}列表</el-button>
          <h1 class="detail-title">{{ profile.title }}</h1>
          <el-tag v-if="profile.inWild != null" :type="profile.inWild ? 'danger' : 'info'" size="small">
            {{ profile.inWild ? '在野利用' : '未发现在野' }}
          </el-tag>
          <el-tag v-if="profile.malwareClass" size="small" :type="malwareClassColor(profile.malwareClass)">
            {{ malwareClassLabel(profile.malwareClass) }}
          </el-tag>
          <el-tag v-if="profile.toolCategory" size="small" type="info">{{ profile.toolCategory }}</el-tag>
          <el-tag v-if="profile.infraRole" size="small" type="warning">
            {{ infraRoleLabel[profile.infraRole] ?? profile.infraRole }}
          </el-tag>
          <p v-if="profile.summary" class="summary">{{ profile.summary }}</p>
        </div>

        <el-card shadow="never" class="meta-card">
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="类型">{{ profile.subtitle }}</el-descriptions-item>
            <el-descriptions-item label="攻击行业" :span="2">
              <IndustryTags :industries="profile.targetIndustries" />
            </el-descriptions-item>
            <el-descriptions-item label="关联知识">
              <strong>{{ profile.knowledgeCount }}</strong> 条
            </el-descriptions-item>
            <el-descriptions-item label="支撑情报">{{ profile.intelReportCount }} 份</el-descriptions-item>
            <el-descriptions-item label="ATT&CK 矩阵">
              <el-tag v-for="m in profile.matrices" :key="m" size="small" class="matrix-tag">{{ m }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never">
          <el-tabs v-model="activeTab">
            <el-tab-pane label="关联图谱" name="graph">
              <EntityTopologyGraph
                :center-label="profile.title.length > 28 ? profile.title.slice(0, 26) + '…' : profile.title"
                :center-subtitle="profile.subtitle"
                :dimension-nodes="topologyNodes"
                :knowledge="profile.knowledge"
              />
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

            <el-tab-pane v-if="profile.threatActors.length" label="攻击组织" name="actors">
              <el-tag
                v-for="a in profile.threatActors"
                :key="a"
                class="actor-tag"
                size="small"
                effect="plain"
              >
                {{ a }}
              </el-tag>
            </el-tab-pane>

            <el-tab-pane v-if="profile.techniques.length" label="攻击手法" name="techniques">
              <AttackTechniques
                :techniques="profile.techniques"
                :matrix="profile.matrices[0] ?? 'Enterprise'"
              />
            </el-tab-pane>

            <el-tab-pane v-if="profile.tools.length" label="攻击工具" name="tools">
              <el-table :data="profile.tools" stripe size="small">
                <el-table-column prop="name" label="工具" min-width="140" />
                <el-table-column prop="category" label="类别" width="120" />
                <el-table-column prop="description" label="说明" min-width="240" show-overflow-tooltip />
              </el-table>
            </el-tab-pane>

            <el-tab-pane v-if="profile.infrastructure.length" label="基础设施" name="infra">
              <el-table :data="profile.infrastructure" stripe size="small">
                <el-table-column prop="type" label="类型" width="80" />
                <el-table-column prop="value" label="值" min-width="200" />
                <el-table-column label="用途" width="100">
                  <template #default="{ row }">{{ infraRoleLabel[row.role] ?? row.role }}</template>
                </el-table-column>
                <el-table-column prop="confidence" label="置信度" width="90" />
              </el-table>
            </el-tab-pane>

            <el-tab-pane v-if="profile.malware.length" label="恶意软件" name="malware">
              <MalwareClassTags :items="profile.malware" />
            </el-tab-pane>

            <el-tab-pane v-if="profile.vulnerabilities.length" label="漏洞利用" name="cve">
              <el-table :data="profile.vulnerabilities" stripe size="small">
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
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </template>

      <el-result v-else-if="!loading" icon="warning" title="未找到该实体">
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
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.detail-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  flex: 1 1 100%;
}
.summary {
  margin: 8px 0 0;
  flex: 1 1 100%;
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
.actor-tag {
  margin: 0 8px 8px 0;
}
</style>
