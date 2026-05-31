<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { searchIoc } from '@/api/intel'
import AppLayout from '@/components/AppLayout.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'

const router = useRouter()
const query = ref('')
const loading = ref(false)
const results = ref<Awaited<ReturnType<typeof searchIoc>>>([])

async function onSearch() {
  loading.value = true
  try {
    results.value = await searchIoc(query.value)
  } finally {
    loading.value = false
  }
}

function goIntel(id: number) {
  router.push({ name: 'intel-detail', params: { id: String(id) } })
}

function goKnowledge(id: number) {
  router.push({ name: 'knowledge-detail', params: { id: String(id) } })
}
</script>

<template>
  <AppLayout>
    <PageShell>
      <PageHeader
        title="IOC 查询"
        description="辅助入口：命中后跳转情报与攻击知识。平台主检索请使用攻击知识库。"
      />

      <el-card shadow="never" class="search-card">
        <el-input
          v-model="query"
          placeholder="IP / 域名 / 哈希片段…"
          clearable
          class="search-input"
          @keyup.enter="onSearch"
        >
          <template #append>
            <el-button type="primary" :loading="loading" @click="onSearch">搜索</el-button>
          </template>
        </el-input>
        <p class="hint">Demo 示例：尝试 <code>198.51.100.10</code> 或 <code>malicious-cdn</code></p>
      </el-card>

      <el-card shadow="never" v-loading="loading" class="result-card">
        <el-table v-if="results.length" :data="results" stripe class="full-table">
          <el-table-column label="IOC 类型" prop="ioc.type" width="100" />
          <el-table-column label="值" prop="ioc.value" min-width="280" show-overflow-tooltip />
          <el-table-column label="置信度" prop="ioc.confidence" width="100" />
          <el-table-column label="来源情报" min-width="280" show-overflow-tooltip>
            <template #default="{ row }">
              <el-link type="primary" @click="goIntel(row.intelId)">{{ row.intelTitle }}</el-link>
            </template>
          </el-table-column>
          <el-table-column label="关联攻击知识" min-width="160">
            <template #default="{ row }">
              <el-link
                v-for="kid in row.knowledgeIds"
                :key="kid"
                type="primary"
                class="kid-link"
                @click="goKnowledge(kid)"
              >
                #{{ kid }}
              </el-link>
              <span v-if="!row.knowledgeIds.length" class="muted">-</span>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else-if="query && !loading" description="无匹配 IOC" />
        <el-empty v-else description="输入 IOC 开始查询" :image-size="80" />
      </el-card>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.search-card,
.result-card {
  background: var(--ami-surface) !important;
  margin-bottom: 16px;
}
.search-input {
  max-width: 560px;
}
.hint {
  margin: 12px 0 0;
  font-size: 13px;
  color: #909399;
}
.full-table {
  width: 100%;
}
.kid-link {
  margin-right: 8px;
}
.muted {
  color: #c0c4cc;
}
</style>
