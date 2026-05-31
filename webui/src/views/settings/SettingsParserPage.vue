<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import {
  createParserProfile,
  deleteParserProfile,
  listParserProfiles,
  patchParserProfile,
} from '@/api/settings'
import type { ParserProfileSummary } from '@/types/settings'

const searchQ = ref('')
const allRows = ref<ParserProfileSummary[]>([])
const rows = ref<ParserProfileSummary[]>([])
const dialogVisible = ref(false)
const isCreate = ref(false)
const editId = ref('')
const editForm = reactive({
  name: '',
  description: '',
})

const dialogTitle = computed(() => (isCreate.value ? '新建解析模板' : '编辑解析模板'))

async function load() {
  allRows.value = await listParserProfiles()
  applyFilter()
}

function applyFilter() {
  const q = searchQ.value.trim().toLowerCase()
  rows.value = q
    ? allRows.value.filter((p) =>
        [p.name, p.description].some((s) => (s ?? '').toLowerCase().includes(q)),
      )
    : [...allRows.value]
}

function openCreate() {
  isCreate.value = true
  editId.value = ''
  editForm.name = ''
  editForm.description = ''
  dialogVisible.value = true
}

function openEdit(row: ParserProfileSummary) {
  isCreate.value = false
  editId.value = row.id
  editForm.name = row.name
  editForm.description = row.description ?? ''
  dialogVisible.value = true
}

async function save() {
  if (!editForm.name.trim()) {
    ElMessage.warning('请填写模板 ID')
    return
  }
  try {
    if (isCreate.value) {
      await createParserProfile({
        name: editForm.name.trim(),
        description: editForm.description.trim(),
      })
      ElMessage.success('已创建')
    } else {
      await patchParserProfile(editId.value, {
        name: editForm.name.trim(),
        description: editForm.description.trim(),
      })
      ElMessage.success('已保存')
    }
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function confirmDelete(row: ParserProfileSummary) {
  if (row.usedBySources > 0) {
    ElMessage.warning(`该模板仍被 ${row.usedBySources} 个采集源引用，无法删除。`)
    return
  }
  try {
    await ElMessageBox.confirm(`确定删除解析模板「${row.name}」？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
    await deleteParserProfile(row.id)
    ElMessage.success('已删除')
    await load()
  } catch (e) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
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
        title="解析模板"
        description="HTML / 微信等正文抽取规则；支持新建、编辑、删除。"
      />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="applyFilter">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="模板 ID / 说明" style="width: 200px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="applyFilter">查询</el-button>
            <el-button type="success" @click="openCreate">新建</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" class="table-card">
        <el-table :data="rows" stripe size="small" class="settings-table">
          <el-table-column prop="name" label="模板 ID" width="200" />
          <el-table-column prop="description" label="说明" min-width="280" />
          <el-table-column prop="usedBySources" label="引用采集源" width="110" align="center" />
          <el-table-column label="操作" min-width="140" fixed="right" align="right">
            <template #default="{ row }">
              <div class="row-actions">
                <el-button size="small" @click.stop="openEdit(row)">编辑</el-button>
                <el-button
                  size="small"
                  type="danger"
                  :disabled="row.usedBySources > 0"
                  @click.stop="confirmDelete(row)"
                >
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-dialog
        v-model="dialogVisible"
        :title="dialogTitle"
        width="480px"
        align-center
        append-to-body
        destroy-on-close
        :close-on-click-modal="false"
      >
        <el-form label-width="90px">
          <el-form-item label="模板 ID" required>
            <el-input v-model="editForm.name" placeholder="如 generic_html_v2" :disabled="!isCreate" />
          </el-form-item>
          <el-form-item label="说明">
            <el-input v-model="editForm.description" type="textarea" :rows="3" />
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
</style>
