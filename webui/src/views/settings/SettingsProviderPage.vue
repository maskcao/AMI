<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import {
  createModelProvider,
  deleteModelProvider,
  listModelProviders,
  patchModelProvider,
  testModelProvider,
} from '@/api/settings'
import type { ModelProviderSummary } from '@/types/settings'

const searchQ = ref('')
const allRows = ref<ModelProviderSummary[]>([])
const rows = ref<ModelProviderSummary[]>([])
const dialogVisible = ref(false)
const isCreate = ref(false)
const editId = ref('')
const editForm = reactive({
  name: '',
  modelId: '',
  baseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
  apiKey: '',
  enabled: true,
  defaultFor: 'parse, enrich',
})
const apiKeyConfigured = ref(false)

const dialogTitle = computed(() => (isCreate.value ? '新建模型服务' : '编辑模型服务'))

async function load() {
  allRows.value = await listModelProviders()
  applyFilter()
}

function applyFilter() {
  const q = searchQ.value.trim().toLowerCase()
  rows.value = q
    ? allRows.value.filter((p) =>
        [p.name, p.modelId, p.baseUrl].some((s) => (s ?? '').toLowerCase().includes(q)),
      )
    : [...allRows.value]
}

function applyArkPreset(kind: 'inference' | 'coding') {
  if (kind === 'coding') {
    editForm.baseUrl = 'https://ark.cn-beijing.volces.com/api/coding/v3'
    if (!editForm.modelId || editForm.modelId === 'auto') {
      editForm.modelId = 'ark-code-latest'
    }
  } else {
    editForm.baseUrl = 'https://ark.cn-beijing.volces.com/api/v3'
    if (editForm.modelId === 'ark-code-latest') {
      editForm.modelId = ''
    }
  }
  ElMessage.info('已填入火山方舟模板，请粘贴 Hermes 中相同的 API Key 与接入点 ID')
}

function resetForm() {
  editForm.name = ''
  editForm.modelId = ''
  editForm.baseUrl = 'https://ark.cn-beijing.volces.com/api/v3'
  editForm.apiKey = ''
  editForm.enabled = true
  editForm.defaultFor = 'parse, enrich'
  apiKeyConfigured.value = false
}

function openCreate() {
  isCreate.value = true
  editId.value = ''
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: ModelProviderSummary) {
  isCreate.value = false
  editId.value = row.id
  editForm.name = row.name
  editForm.modelId = row.modelId ?? ''
  editForm.baseUrl = row.baseUrl ?? ''
  editForm.apiKey = ''
  editForm.enabled = row.enabled
  editForm.defaultFor = row.defaultFor ?? 'parse, enrich'
  apiKeyConfigured.value = !!row.apiKeyConfigured
  dialogVisible.value = true
}

async function runTest(row: ModelProviderSummary) {
  const res = await testModelProvider(row.id)
  if (res.ok) {
    ElMessage.success(`「${row.name}」连通正常`)
  } else {
    ElMessage.error(res.error ?? '连通失败，请检查 API Key 与接入点 ID')
  }
}

async function confirmDelete(row: ModelProviderSummary) {
  try {
    await ElMessageBox.confirm(`确定删除模型服务「${row.name}」？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
    await deleteModelProvider(row.id)
    ElMessage.success('已删除')
    await load()
  } catch (e) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

async function save() {
  if (!editForm.name.trim()) {
    ElMessage.warning('请填写配置名称')
    return
  }
  try {
    const payload = {
      name: editForm.name.trim(),
      modelId: editForm.modelId.trim(),
      baseUrl: editForm.baseUrl.trim(),
      enabled: editForm.enabled,
      defaultFor: editForm.defaultFor.trim(),
    }
    if (isCreate.value) {
      await createModelProvider({
        ...payload,
        apiKey: editForm.apiKey.trim(),
      })
      ElMessage.success('已创建')
    } else {
      const patch: Record<string, unknown> = { ...payload }
      if (editForm.apiKey.trim()) {
        patch.apiKey = editForm.apiKey.trim()
      }
      await patchModelProvider(editId.value, patch)
      ElMessage.success('已保存')
    }
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="模型服务"
        description="OpenAI 兼容网关（含火山方舟）；配置名称、接入点 ID、Base URL 与 API Key 保存在本地数据库。"
      />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="applyFilter">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="名称 / 模型 ID" style="width: 200px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="applyFilter">查询</el-button>
            <el-button type="success" @click="openCreate">新建</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" class="table-card">
        <el-table :data="rows" stripe size="small" class="settings-table">
          <el-table-column prop="name" label="配置名称" min-width="140" />
          <el-table-column prop="modelId" label="模型 / 接入点 ID" min-width="200" show-overflow-tooltip />
          <el-table-column prop="baseUrl" label="Base URL" min-width="180" show-overflow-tooltip />
          <el-table-column label="密钥" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.apiKeyConfigured ? 'success' : 'info'" size="small">
                {{ row.apiKeyConfigured ? '已配置' : '未配置' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="启用" width="72" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" min-width="160" fixed="right" align="right">
            <template #default="{ row }">
              <div class="row-actions">
                <el-button size="small" @click.stop="runTest(row)">测试</el-button>
                <el-button size="small" @click.stop="openEdit(row)">编辑</el-button>
                <el-button size="small" type="danger" @click.stop="confirmDelete(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-dialog
        v-model="dialogVisible"
        :title="dialogTitle"
        width="560px"
        align-center
        append-to-body
        destroy-on-close
        :close-on-click-modal="false"
      >
        <el-form label-width="120px">
          <el-form-item label="配置名称" required>
            <el-input v-model="editForm.name" placeholder="如：火山方舟生产" />
          </el-form-item>
          <el-form-item label="接入点 / 模型 ID" required>
            <el-input
              v-model="editForm.modelId"
              placeholder="火山方舟控制台中的接入点 ID（ep-xxx 或 ark-xxx）"
            />
          </el-form-item>
          <el-form-item label="Base URL" required>
            <el-input
              v-model="editForm.baseUrl"
              placeholder="https://ark.cn-beijing.volces.com/api/v3"
            />
            <p class="field-hint">
              与 Hermes 保持一致：推理用 <code>/api/v3</code> + 控制台接入点 ID；编程计划用
              <code>/api/coding/v3</code> + Hermes 中的 model 名。勿用 <code>/api/plan</code>。
            </p>
            <div class="preset-row">
              <el-button size="small" @click="applyArkPreset('inference')">填入：推理 /api/v3</el-button>
              <el-button size="small" @click="applyArkPreset('coding')">填入：Hermes 编程 /api/coding/v3</el-button>
            </div>
          </el-form-item>
          <el-form-item label="API Key" :required="isCreate">
            <el-input
              v-model="editForm.apiKey"
              type="password"
              show-password
              :placeholder="apiKeyConfigured && !isCreate ? '已配置，留空则不修改' : '火山方舟控制台获取的 API Key'"
            />
          </el-form-item>
          <el-form-item label="默认用途">
            <el-input v-model="editForm.defaultFor" placeholder="parse, enrich" />
          </el-form-item>
          <el-form-item label="启用">
            <el-switch v-model="editForm.enabled" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="save">保存</el-button>
        </template>
      </el-dialog>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.table-card {
  background: var(--ami-surface) !important;
}
.settings-table {
  width: 100%;
}
.row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: flex-end;
}
.preset-row {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.field-hint {
  margin: 6px 0 0;
  font-size: 12px;
  color: var(--ami-text-secondary);
  line-height: 1.5;
}
.field-hint code {
  font-size: 11px;
}
</style>
