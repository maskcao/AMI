<script setup lang="ts">
import AttackTechniques from '@/components/AttackTechniques.vue'
import MalwareClassTags from '@/components/MalwareClassTags.vue'
import type { KnowledgeDetail } from '@/types/knowledge'

defineProps<{
  detail: KnowledgeDetail
  attackMatrix: string
}>()

const infraRoleLabel: Record<string, string> = {
  c2: 'C2',
  hosting: '托管',
  payload: '载荷分发',
  scan: '扫描源',
  other: '其他',
}
</script>

<template>
  <section class="block">
    <h3>恶意软件（含分类）</h3>
    <MalwareClassTags :items="detail.malware" />
  </section>

  <section class="block">
    <h3>攻击手法 (MITRE ATT&CK)</h3>
    <AttackTechniques :techniques="detail.techniques" :matrix="attackMatrix" />
  </section>

  <section class="block">
    <h3>攻击工具</h3>
    <el-table v-if="detail.tools.length" :data="detail.tools" size="small" stripe>
      <el-table-column prop="name" label="工具名称" min-width="140" />
      <el-table-column prop="category" label="类别" width="120" />
      <el-table-column prop="description" label="说明" min-width="240" show-overflow-tooltip />
    </el-table>
    <el-empty v-else description="暂无攻击工具" :image-size="56" />
  </section>

  <section class="block">
    <h3>攻击者基础设施</h3>
    <el-table
      v-if="detail.infrastructure.length"
      :data="detail.infrastructure"
      size="small"
      stripe
    >
      <el-table-column prop="type" label="类型" width="80" />
      <el-table-column prop="value" label="值" min-width="200" />
      <el-table-column label="用途" width="100">
        <template #default="{ row }">{{ infraRoleLabel[row.role] ?? row.role }}</template>
      </el-table-column>
      <el-table-column prop="confidence" label="置信度" width="90" />
      <el-table-column prop="note" label="备注" min-width="160" show-overflow-tooltip />
    </el-table>
    <el-empty v-else description="暂无基础设施情报" :image-size="56" />
  </section>

  <section class="block">
    <h3>利用漏洞</h3>
    <el-table
      v-if="detail.vulnerabilities.length"
      :data="detail.vulnerabilities"
      size="small"
      stripe
    >
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
      <el-table-column label="在野利用" width="100">
        <template #default="{ row }">
          <el-tag :type="row.inWild ? 'danger' : 'info'" size="small">
            {{ row.inWild ? '是' : '否' }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-else description="暂无关联漏洞" :image-size="56" />
  </section>
</template>

<style scoped>
.block {
  margin-bottom: 28px;
}
.block h3 {
  margin: 0 0 12px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}
</style>
