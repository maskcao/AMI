<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Download, EditPen, Search, UploadFilled, View } from '@element-plus/icons-vue'
import { marked } from 'marked'
import {
  deleteIncidentReport,
  fetchIncidentDocxHtml,
  fetchIncidentMarkdown,
  incidentDownloadUrl,
  listIncidentReports,
  updateIncidentReportTime,
  uploadIncidentReport,
  type IncidentReport,
} from '@/api/incident'
import { USE_API } from '@/api/http'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'

const loading = ref(false)
const uploading = ref(false)
const items = ref<IncidentReport[]>([])
const searchInput = ref('')
const appliedQ = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

const pendingFiles = ref<File[]>([])
const uploadDialogVisible = ref(false)
const uploadReportTime = ref('')
const editTimeDialogVisible = ref(false)
const editingRow = ref<IncidentReport | null>(null)
const editReportTime = ref('')

const filteredItems = computed(() => {
  const q = appliedQ.value.trim().toLowerCase()
  if (!q) return items.value
  const tokens = q.split(/\s+/).filter(Boolean)
  return items.value.filter((it) => {
    const target = `${it.title}\n${it.filename}`.toLowerCase()
    return tokens.every((t) => target.includes(t))
  })
})

const pagedItems = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredItems.value.slice(start, start + pageSize.value)
})

function doSearch() {
  appliedQ.value = searchInput.value
  currentPage.value = 1
}

function resetSearch() {
  searchInput.value = ''
  appliedQ.value = ''
  currentPage.value = 1
}

const previewVisible = ref(false)
const previewTitle = ref('')
const previewKind = ref<'md' | 'docx'>('md')
const previewHtml = ref('')
const previewLoading = ref(false)

async function load() {
  if (!USE_API) {
    items.value = []
    return
  }
  loading.value = true
  try {
    const res = await listIncidentReports()
    items.value = res.items
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载报告列表失败')
  } finally {
    loading.value = false
  }
}

async function onUpload(file: File) {
  const ext = file.name.slice(file.name.lastIndexOf('.')).toLowerCase()
  if (ext !== '.md' && ext !== '.markdown' && ext !== '.docx') {
    ElMessage.error('仅支持 .md / .docx 文件')
    return false
  }
  pendingFiles.value.push(file)
  uploadDialogVisible.value = true
  return false
}

function cancelUpload() {
  pendingFiles.value = []
  uploadReportTime.value = ''
  uploadDialogVisible.value = false
}

async function confirmUpload() {
  const files = pendingFiles.value
  if (files.length === 0) return
  uploading.value = true
  let okCount = 0
  try {
    for (const file of files) {
      try {
        await uploadIncidentReport(file, uploadReportTime.value || '')
        okCount++
      } catch (e) {
        ElMessage.error(`${file.name}：${e instanceof Error ? e.message : '上传失败'}`)
      }
    }
    if (okCount > 0) ElMessage.success(`已上传 ${okCount} 个报告`)
    cancelUpload()
    currentPage.value = 1
    await load()
  } finally {
    uploading.value = false
  }
}

function openEditTime(row: IncidentReport) {
  editingRow.value = row
  editReportTime.value = row.reportTime || ''
  editTimeDialogVisible.value = true
}

async function saveEditTime() {
  if (!editingRow.value) return
  try {
    await updateIncidentReportTime(editingRow.value.id, editReportTime.value || '')
    ElMessage.success('报告时间已更新')
    editTimeDialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '更新失败')
  }
}

async function openPreview(row: IncidentReport) {
  const ext = (row.ext || '').toLowerCase()
  if (ext !== '.md' && ext !== '.markdown' && ext !== '.docx') {
    ElMessage.warning('该文件类型不支持在线预览')
    return
  }
  previewTitle.value = row.title
  previewKind.value = ext === '.docx' ? 'docx' : 'md'
  previewHtml.value = ''
  previewVisible.value = true
  previewLoading.value = true
  try {
    if (previewKind.value === 'md') {
      const text = await fetchIncidentMarkdown(row.id)
      previewHtml.value = await marked.parse(text)
    } else {
      previewHtml.value = await fetchIncidentDocxHtml(row.id)
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '预览加载失败')
    previewVisible.value = false
  } finally {
    previewLoading.value = false
  }
}

async function confirmDelete(row: IncidentReport) {
  try {
    await ElMessageBox.confirm(`确定删除报告「${row.title}」？`, '删除确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteIncidentReport(row.id)
    ElMessage.success('已删除')
    currentPage.value = 1
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

function extTagType(ext: string) {
  if (ext === '.docx') return 'warning'
  return 'success'
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

function formatTime(iso: string) {
  return iso ? iso.replace('T', ' ').slice(0, 19) : ''
}

onMounted(() => {
  void load()
})
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="安全事件报告"
        description="上传 Word（.docx）或 Markdown（.md）格式的安全事件报告，支持在线预览与原件下载。"
      />

      <el-alert
        v-if="!USE_API"
        type="warning"
        :closable="false"
        show-icon
        title="当前为 Mock 模式，安全事件报告需连接 API 服务（VITE_USE_API=true）。"
        class="api-hint"
      />

      <FilterCard>
        <div class="toolbar">
          <div class="toolbar-left">
            <el-input
              v-model="searchInput"
              clearable
              placeholder="输入关键词模糊搜索，空格分隔多个词"
              :prefix-icon="Search"
              class="search-input"
              @keyup.enter="doSearch"
              @clear="resetSearch"
            />
            <el-button type="primary" :icon="Search" @click="doSearch">搜索</el-button>
            <el-button @click="resetSearch">重置</el-button>
          </div>
          <el-upload
            multiple
            :show-file-list="false"
            :auto-upload="false"
            accept=".md,.markdown,.docx"
            :on-change="(f: any) => f.raw && onUpload(f.raw as File)"
          >
            <el-button type="primary" :icon="UploadFilled" :loading="uploading">上传报告</el-button>
            <template #tip>
              <span class="upload-tip">支持 .md / .docx，单个不超过 50MB</span>
            </template>
          </el-upload>
        </div>
      </FilterCard>

      <el-card shadow="never" v-loading="loading" class="table-card">
        <el-table :data="pagedItems" stripe size="small" empty-text="暂无报告">
          <el-table-column prop="title" label="报告标题" min-width="220" show-overflow-tooltip />
          <el-table-column prop="filename" label="文件名" min-width="200" show-overflow-tooltip />
          <el-table-column label="报告时间" width="150" align="center">
            <template #default="{ row }">
              <span :class="{ 'time-empty': !row.reportTime }">{{ row.reportTime || '未填写' }}</span>
              <el-button size="small" link class="time-edit-btn" :icon="EditPen" @click="openEditTime(row)" />
            </template>
          </el-table-column>
          <el-table-column label="类型" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="extTagType(row.ext)" size="small">{{ row.ext }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="大小" width="100" align="right">
            <template #default="{ row }">{{ formatSize(row.sizeBytes) }}</template>
          </el-table-column>
          <el-table-column label="上传时间" width="168">
            <template #default="{ row }">{{ formatTime(row.uploadedAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="200" align="center">
            <template #default="{ row }">
              <el-button size="small" type="primary" link :icon="View" @click="openPreview(row)">预览</el-button>
              <a :href="incidentDownloadUrl(row.id)" class="dl-link">
                <el-button size="small" link :icon="Download">下载</el-button>
              </a>
              <el-button size="small" type="danger" link :icon="Delete" @click="confirmDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-row">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="filteredItems.length"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            small
          />
        </div>
      </el-card>

      <el-dialog
        v-model="uploadDialogVisible"
        title="填写报告信息"
        width="420px"
        :close-on-click-modal="false"
        @close="cancelUpload"
      >
        <div class="pending-files">
          <div v-for="f in pendingFiles" :key="f.name + f.size" class="pending-file">{{ f.name }}</div>
        </div>
        <el-form label-width="90px">
          <el-form-item label="报告时间">
            <el-date-picker
              v-model="uploadReportTime"
              type="date"
              value-format="YYYY-MM-DD"
              placeholder="选择事件发生/报告时间（可留空）"
              style="width: 100%"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="cancelUpload">取消</el-button>
          <el-button type="primary" :loading="uploading" @click="confirmUpload">确认上传</el-button>
        </template>
      </el-dialog>

      <el-dialog v-model="editTimeDialogVisible" title="修改报告时间" width="380px">
        <el-form label-width="90px">
          <el-form-item label="报告时间">
            <el-date-picker
              v-model="editReportTime"
              type="date"
              value-format="YYYY-MM-DD"
              placeholder="选择事件发生/报告时间（可留空）"
              style="width: 100%"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="editTimeDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveEditTime">保存</el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="previewVisible"
        :title="`预览：${previewTitle}`"
        width="72%"
        top="5vh"
        destroy-on-close
      >
        <div v-loading="previewLoading" class="preview-body">
          <div class="preview-content" v-html="previewHtml"></div>
        </div>
        <template #footer>
          <el-button @click="previewVisible = false">关闭</el-button>
        </template>
      </el-dialog>
    </PageShell>
  </AppLayout>
</template>

<style scoped>
.api-hint {
  margin-bottom: 12px;
}
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.search-input {
  width: 280px;
}
.upload-tip {
  margin-left: 10px;
  font-size: 12px;
  color: #94a3b8;
}
.table-card {
  background: var(--ami-surface) !important;
}
.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
.dl-link {
  color: inherit;
  text-decoration: none;
}
.time-empty {
  color: #94a3b8;
}
.time-edit-btn {
  margin-left: 4px;
}
.pending-files {
  margin-bottom: 14px;
  max-height: 140px;
  overflow-y: auto;
}
.pending-file {
  font-size: 13px;
  color: #475569;
  padding: 4px 0;
}
.preview-body {
  min-height: 200px;
  max-height: 70vh;
  overflow-y: auto;
}
.preview-content {
  line-height: 1.7;
  word-break: break-word;
}
.preview-content :deep(h1),
.preview-content :deep(h2),
.preview-content :deep(h3) {
  margin: 16px 0 8px;
}
.preview-content :deep(table) {
  border-collapse: collapse;
}
.preview-content :deep(th),
.preview-content :deep(td) {
  border: 1px solid #d1d5db;
  padding: 4px 8px;
}
.preview-content :deep(img) {
  max-width: 100%;
}
</style>
