<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createCollectSource,
  deleteCollectSource,
  listCollectSources,
  patchCollectSource,
  runCollectSource,
  testCollectArticle,
  testCollectList,
  unlockCollectBatch,
  COLLECT_TYPE_LABELS,
  type CollectSourceDetail,
  type CollectSourceConfig,
} from '@/api/settings'
import type { PilotQualityCheck } from '@/types/settings'
import AppLayout from '@/components/AppLayout.vue'
import FilterCard from '@/components/FilterCard.vue'
import PageHeader from '@/components/PageHeader.vue'
import PageShell from '@/components/PageShell.vue'
import type { CollectSourceType } from '@/types/settings'

const searchQ = ref('')
const filterType = ref<CollectSourceType | ''>('')
const filterEnabled = ref<'' | 'yes' | 'no'>('')

const typeOptions = Object.entries(COLLECT_TYPE_LABELS).map(([value, label]) => ({ value, label }))
const createTypeOptions = typeOptions.filter((o) =>
  ['direct_url', 'paginated_list', 'threat_portal', 'rss', 'scheduled_url', 'wechat_feed', 'manual_upload'].includes(
    o.value,
  ),
)

/** Ghost / WordPress style: list + /page/2/, /page/3/ … */
function pathTemplateFromListUrl(listUrl: string): string {
  const base = listUrl.trim().replace(/\/+$/, '')
  if (!base) return ''
  return `${base}/page/{page}/`
}

function applyPathTemplateFromList() {
  const t = pathTemplateFromListUrl(editForm.listUrl)
  if (!t) {
    ElMessage.warning('请先填写列表 URL')
    return
  }
  editForm.pathTemplate = t
  editForm.paginationType = 'path'
  if (editForm.startPage < 2) editForm.startPage = 2
  ElMessage.success('已生成 …/page/{page}/ 模板')
}

const allRows = ref<CollectSourceDetail[]>([])
const rows = ref<CollectSourceDetail[]>([])
const dialogVisible = ref(false)
const isCreate = ref(false)
const editId = ref('')
const editForm = reactive({
  type: 'rss' as CollectSourceType,
  name: '',
  enabled: true,
  sourceLabel: '',
  scheduleLabel: '',
  feedUrl: '',
  articleUrlsText: '',
  listUrl: '',
  articleBase: '',
  linkMustContain: '',
  linkExcludeText: '',
  linkPreferRegex: '',
  paginationType: 'path' as 'none' | 'path' | 'scroll' | 'wordpress_ajax',
  pathTemplate: '',
  startPage: 2,
  maxPages: 3,
  stopOnEmptyList: true,
  maxItems: 5,
  delayMs: 500,
  titleBlocklistText: '',
  titleExtractRegexText: '',
})
const runningId = ref('')
const pilotDialogVisible = ref(false)
const pilotRow = ref<CollectSourceDetail | null>(null)
const pilotArticleUrl = ref('')
const pilotArticleLoading = ref(false)
const pilotListLoading = ref(false)
const pilotUnlockLoading = ref(false)
const pilotListPreview = ref<string[]>([])
const pilotListCount = ref(0)
const lastArticleChecks = ref<PilotQualityCheck[]>([])

const dialogTitle = computed(() => (isCreate.value ? '新建采集源' : '编辑采集源'))
const formType = computed(() => {
  if (isCreate.value) return editForm.type
  return allRows.value.find((r) => r.id === editId.value)?.type ?? editForm.type
})

function parseRowConfig(row: CollectSourceDetail): CollectSourceConfig {
  const c = row.config
  if (typeof c === 'string') {
    try {
      return JSON.parse(c) as CollectSourceConfig
    } catch {
      return {}
    }
  }
  return c ?? {}
}

async function loadSources() {
  allRows.value = await listCollectSources()
  applyFilter()
}

function applyFilter() {
  let list = [...allRows.value]
  const q = searchQ.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (r) =>
        r.name.toLowerCase().includes(q) ||
        r.sourceLabel.toLowerCase().includes(q) ||
        r.typeLabel.includes(q),
    )
  }
  if (filterType.value) list = list.filter((r) => r.type === filterType.value)
  if (filterEnabled.value === 'yes') list = list.filter((r) => r.enabled)
  if (filterEnabled.value === 'no') list = list.filter((r) => !r.enabled)
  rows.value = list
}

function resetForm() {
  editForm.type = 'rss'
  editForm.name = ''
  editForm.enabled = true
  editForm.sourceLabel = ''
  editForm.scheduleLabel = ''
  editForm.feedUrl = ''
  editForm.articleUrlsText = ''
  editForm.listUrl = ''
  editForm.articleBase = ''
  editForm.linkMustContain = ''
  editForm.linkExcludeText = ''
  editForm.linkPreferRegex = ''
  editForm.paginationType = 'path'
  editForm.pathTemplate = ''
  editForm.startPage = 2
  editForm.maxPages = 3
  editForm.stopOnEmptyList = true
  editForm.maxItems = 5
  editForm.delayMs = 500
  editForm.titleBlocklistText = ''
  editForm.titleExtractRegexText = ''
}

function loadTitleExtractFields(cfg: CollectSourceConfig) {
  editForm.titleBlocklistText = (cfg.title_blocklist ?? []).join('\n')
  editForm.titleExtractRegexText = (cfg.title_extract ?? [])
    .map((r) => r.regex?.trim() ?? '')
    .filter(Boolean)
    .join('\n')
}

function appendTitleExtractConfig(cfg: CollectSourceConfig) {
  const blocklist = editForm.titleBlocklistText
    .split(/[,，\n]/)
    .map((s) => s.trim())
    .filter(Boolean)
  const rules = editForm.titleExtractRegexText
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)
    .map((regex) => ({ regex }))
  if (blocklist.length) cfg.title_blocklist = blocklist
  if (rules.length) cfg.title_extract = rules
}

function openCreate() {
  isCreate.value = true
  editId.value = ''
  resetForm()
  dialogVisible.value = true
}

function defaultPilotArticleUrl(cfg: CollectSourceConfig): string {
  const u = cfg.article_urls?.find((x) => x.trim())
  return u?.trim() ?? ''
}

function pilotArticleDone(row: CollectSourceDetail): boolean {
  return !!row.pilotArticlePassedAt?.trim()
}

function pilotListDone(row: CollectSourceDetail): boolean {
  if (!row.needsListPilot) return true
  return !!row.pilotListPassedAt?.trim()
}

function canUnlockBatch(row: CollectSourceDetail): boolean {
  return pilotArticleDone(row) && pilotListDone(row) && !row.batchUnlocked
}

function openPilot(row: CollectSourceDetail) {
  if (row.type === 'manual_upload') return
  pilotRow.value = row
  pilotArticleUrl.value = defaultPilotArticleUrl(parseRowConfig(row))
  lastArticleChecks.value = []
  pilotListPreview.value = []
  pilotListCount.value = 0
  pilotDialogVisible.value = true
}

function openEdit(row: CollectSourceDetail) {
  isCreate.value = false
  editId.value = row.id
  editForm.type = row.type
  editForm.name = row.name
  editForm.enabled = row.enabled
  editForm.sourceLabel = row.sourceLabel
  editForm.scheduleLabel = row.scheduleLabel ?? ''
  const cfg = parseRowConfig(row)
  editForm.feedUrl = cfg.feed_url ?? ''
  editForm.articleUrlsText = (cfg.article_urls ?? []).join('\n')
  editForm.listUrl = cfg.list_url ?? cfg.entry_url ?? ''
  editForm.articleBase = cfg.article_base ?? ''
  const lex = cfg.link_extract
  editForm.linkMustContain = lex?.must_contain ?? cfg.link_selector ?? ''
  editForm.linkExcludeText = (lex?.exclude_contains ?? []).join(', ')
  editForm.linkPreferRegex = lex?.prefer_path_regex ?? ''
  editForm.paginationType = cfg.pagination?.type ?? 'path'
  editForm.pathTemplate = cfg.pagination?.path_template ?? ''
  editForm.startPage = cfg.pagination?.start_page ?? 2
  editForm.maxPages = cfg.pagination?.max_pages ?? 3
  editForm.stopOnEmptyList = cfg.pagination?.stop_on_empty_list ?? true
  editForm.maxItems = cfg.max_items_per_run ?? 5
  editForm.delayMs = cfg.delay_ms ?? 500
  loadTitleExtractFields(cfg)
  dialogVisible.value = true
}

function mergeCollectSourceRow(updated: CollectSourceDetail) {
  const i = allRows.value.findIndex((r) => r.id === updated.id)
  if (i >= 0) {
    allRows.value[i] = { ...allRows.value[i], ...updated }
  }
  if (pilotRow.value?.id === updated.id) {
    pilotRow.value = { ...pilotRow.value, ...updated }
  }
  applyFilter()
}

async function runPilotArticle() {
  const row = pilotRow.value
  if (!row) return
  if (!pilotArticleUrl.value.trim()) {
    ElMessage.warning('请填写试点文章 URL')
    return
  }
  pilotArticleLoading.value = true
  try {
    const res = await testCollectArticle(row.id, {
      url: pilotArticleUrl.value.trim(),
      waitPipeline: true,
    })
    lastArticleChecks.value = res.checks ?? []
    mergeCollectSourceRow(res.collectSource)
    if (res.passed) {
      ElMessage.success(`单条试采通过（情报 #${res.intelId}）`)
    } else {
      ElMessage.warning(`单条试采完成但未通过质检（情报 #${res.intelId}）`)
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '单条试采失败')
  } finally {
    pilotArticleLoading.value = false
  }
}

async function runPilotList() {
  const row = pilotRow.value
  if (!row) return
  const cfg = parseRowConfig(row)
  const maxPages = cfg.pagination?.type === 'path' ? (cfg.pagination.max_pages ?? 2) : 1
  pilotListLoading.value = true
  try {
    const res = await testCollectList(row.id, { maxPages })
    mergeCollectSourceRow(res.collectSource)
    pilotListCount.value = res.linkCount ?? 0
    pilotListPreview.value = res.sampleUrls ?? []
    if (res.skipped) {
      ElMessage.success(res.message ?? '无需分页试采')
      return
    }
    if (res.passed) {
      ElMessage.success(`分页试采通过：发现 ${res.linkCount ?? 0} 条链接`)
    } else {
      ElMessage.warning('分页试采未发现文章链接，请检查列表 URL / 抽链规则')
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '分页试采失败')
  } finally {
    pilotListLoading.value = false
  }
}

async function runUnlockBatch() {
  const row = pilotRow.value
  if (!row) return
  pilotUnlockLoading.value = true
  try {
    const res = await unlockCollectBatch(row.id)
    mergeCollectSourceRow(res.collectSource)
    ElMessage.success('已解锁批量运行与定时调度')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '解锁失败')
  } finally {
    pilotUnlockLoading.value = false
  }
}

const isListType = computed(() =>
  ['paginated_list', 'threat_portal', 'scheduled_url', 'wechat_feed'].includes(formType.value),
)

function buildConfig(): CollectSourceConfig {
  if (formType.value === 'rss') {
    const cfg: CollectSourceConfig = {
      feed_url: editForm.feedUrl.trim(),
      max_items_per_run: editForm.maxItems,
    }
    appendTitleExtractConfig(cfg)
    return cfg
  }
  if (formType.value === 'direct_url') {
    const article_urls = editForm.articleUrlsText
      .split('\n')
      .map((s) => s.trim())
      .filter(Boolean)
    const cfg: CollectSourceConfig = {
      article_urls,
      max_items_per_run: editForm.maxItems,
      timeout_sec: 30,
    }
    appendTitleExtractConfig(cfg)
    return cfg
  }
  const exclude_contains = editForm.linkExcludeText
    .split(/[,，]/)
    .map((s) => s.trim())
    .filter(Boolean)
  const cfg: CollectSourceConfig = {
    list_url: editForm.listUrl.trim(),
    article_base: editForm.articleBase.trim() || undefined,
    link_extract: {
      must_contain: editForm.linkMustContain.trim() || undefined,
      exclude_contains: exclude_contains.length ? exclude_contains : undefined,
      prefer_path_regex: editForm.linkPreferRegex.trim() || undefined,
    },
    max_items_per_run: editForm.maxItems,
    delay_ms: editForm.delayMs,
  }
  if (isListType.value && editForm.paginationType !== 'none') {
    cfg.pagination = {
      type: editForm.paginationType,
      path_template: editForm.pathTemplate.trim() || undefined,
      start_page: editForm.startPage,
      max_pages: editForm.maxPages,
      stop_on_empty_list: editForm.stopOnEmptyList,
    }
  }
  appendTitleExtractConfig(cfg)
  return cfg
}

async function save() {
  if (!editForm.name.trim()) {
    ElMessage.warning('请填写名称')
    return
  }
  try {
    const body = {
      name: editForm.name.trim(),
      enabled: editForm.enabled,
      sourceLabel: editForm.sourceLabel.trim(),
      scheduleLabel: editForm.scheduleLabel.trim() || undefined,
      config: buildConfig(),
    }
    if (isCreate.value) {
      await createCollectSource({
        ...body,
        type: editForm.type,
      })
      ElMessage.success('已创建')
    } else {
      await patchCollectSource(editId.value, body)
      ElMessage.success('已保存；试点状态已重置，请在列表中重新试采')
    }
    dialogVisible.value = false
    await loadSources()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function confirmDelete(row: CollectSourceDetail) {
  try {
    await ElMessageBox.confirm(`确定删除采集源「${row.name}」？删除后不可恢复。`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
    await deleteCollectSource(row.id)
    ElMessage.success('已删除')
    await loadSources()
  } catch (e) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

async function confirmRun(row: CollectSourceDetail) {
  if (!row.batchUnlocked) {
    ElMessage.warning('请先完成上线前测试并解锁批量运行')
    openPilot(row)
    return
  }
  try {
    await ElMessageBox.confirm(
      `立即执行采集源「${row.name}」？将从配置的目标拉取新情报并进入流水线。`,
      '运行确认',
      { type: 'info', confirmButtonText: '开始运行', cancelButtonText: '取消' },
    )
    await runNow(row)
  } catch (e) {
    if (e === 'cancel' || e === 'close') return
  }
}

async function runNow(row: CollectSourceDetail) {
  runningId.value = row.id
  try {
    const res = await runCollectSource(row.id)
    if (res.error) {
      ElMessage.warning(
        `采集完成：导入 ${res.imported ?? 0}，跳过 ${res.skipped ?? 0}，失败 ${res.failed ?? 0}；${res.error}`,
      )
    } else {
      ElMessage.success(`导入 ${res.imported ?? 0} 条，跳过 ${res.skipped ?? 0} 条`)
    }
    await loadSources()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '运行失败')
  } finally {
    runningId.value = ''
  }
}

function statusTagType(status: string) {
  if (status === 'success') return 'success'
  if (status === 'partial') return 'warning'
  if (status === 'failed') return 'danger'
  return 'info'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    success: '成功',
    partial: '部分成功',
    failed: '失败',
    skipped: '跳过',
    never: '未运行',
  }
  return map[status] ?? status
}

onMounted(() => {
  void loadSources()
})
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <PageHeader
        title="采集配置"
        description="编辑保存采集规则；列表中进行上线前试采，通过并解锁后再批量运行或依赖定时调度。"
      />

      <FilterCard>
        <el-form :inline="true" @submit.prevent="applyFilter">
          <el-form-item label="关键词">
            <el-input v-model="searchQ" clearable placeholder="名称 / 来源标签" style="width: 200px" />
          </el-form-item>
          <el-form-item label="采集类型">
            <el-select v-model="filterType" clearable placeholder="全部" style="width: 160px">
              <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="启用">
            <el-select v-model="filterEnabled" clearable placeholder="全部" style="width: 100px">
              <el-option label="是" value="yes" />
              <el-option label="否" value="no" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="applyFilter">查询</el-button>
            <el-button type="success" @click="openCreate">新建</el-button>
          </el-form-item>
        </el-form>
      </FilterCard>

      <el-card shadow="never" class="table-card">
        <el-table :data="rows" stripe size="small" class="settings-table">
          <el-table-column prop="name" label="名称" min-width="180" />
          <el-table-column prop="typeLabel" label="类型" width="140" />
          <el-table-column prop="sourceLabel" label="来源标签" width="100" />
          <el-table-column prop="scheduleLabel" label="调度" width="110" />
          <el-table-column label="启用" width="72" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small" effect="light">
                {{ row.enabled ? '是' : '否' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="上线前测试" min-width="200">
            <template #default="{ row }">
              <div v-if="row.type === 'manual_upload'" class="pilot-cell-muted">—</div>
              <div v-else class="pilot-cell">
                <el-space wrap :size="4">
                  <el-tag :type="pilotArticleDone(row) ? 'success' : 'info'" size="small" effect="plain">
                    单条{{ pilotArticleDone(row) ? '✓' : '—' }}
                  </el-tag>
                  <el-tag
                    v-if="row.needsListPilot"
                    :type="pilotListDone(row) ? 'success' : 'info'"
                    size="small"
                    effect="plain"
                  >
                    分页{{ pilotListDone(row) ? '✓' : '—' }}
                  </el-tag>
                  <el-tag :type="row.batchUnlocked ? 'success' : 'warning'" size="small" effect="plain">
                    {{ row.batchUnlocked ? '已解锁' : '未解锁' }}
                  </el-tag>
                </el-space>
                <el-button size="small" type="warning" plain class="pilot-open-btn" @click.stop="openPilot(row)">
                  试点
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="上次运行" width="100">
            <template #default="{ row }">
              {{ row.lastRunAt ? row.lastRunAt.slice(0, 10) : '—' }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.lastStatus)" size="small" effect="light">
                {{ statusLabel(row.lastStatus) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" min-width="300" fixed="right" align="right">
            <template #default="{ row }">
              <div class="row-actions">
                <el-button size="small" @click.stop="openEdit(row)">编辑</el-button>
                <el-button
                  size="small"
                  type="primary"
                  plain
                  :loading="runningId === row.id"
                  :disabled="row.type === 'manual_upload' || !row.batchUnlocked"
                  @click.stop="confirmRun(row)"
                >
                  运行
                </el-button>
                <el-button size="small" type="danger" @click.stop="confirmDelete(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-dialog
        v-model="dialogVisible"
        :title="dialogTitle"
        width="720px"
        align-center
        append-to-body
        destroy-on-close
        :close-on-click-modal="false"
        class="collect-source-dialog"
      >
        <el-form label-width="96px" class="collect-source-form">
          <el-row v-if="isCreate" :gutter="16">
            <el-col :span="24">
              <el-form-item label="采集类型" required>
                <el-select v-model="editForm.type" style="width: 100%">
                  <el-option
                    v-for="opt in createTypeOptions"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="名称" required>
                <el-input v-model="editForm.name" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="来源标签">
                <el-input v-model="editForm.sourceLabel" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="启用">
                <el-switch v-model="editForm.enabled" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="调度说明">
                <el-input v-model="editForm.scheduleLabel" placeholder="如每 6 小时" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row v-if="formType === 'rss'" :gutter="16">
            <el-col :span="24">
              <el-form-item label="Feed URL">
                <el-input v-model="editForm.feedUrl" type="textarea" :rows="2" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row v-else-if="formType === 'direct_url'" :gutter="16">
            <el-col :span="24">
              <el-form-item label="文章 URL">
                <el-input
                  v-model="editForm.articleUrlsText"
                  type="textarea"
                  :rows="4"
                  placeholder="每行一个详情页地址"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <template v-else-if="isListType">
            <el-row :gutter="16">
              <el-col :span="24">
                <el-form-item label="列表 URL" required>
                  <el-input v-model="editForm.listUrl" type="textarea" :rows="2" placeholder="如 tag 列表首页" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="站点根 URL">
                  <el-input v-model="editForm.articleBase" placeholder="相对链接解析" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="链接须含">
                  <el-input v-model="editForm.linkMustContain" placeholder="域名或 URL 片段，如 example.com" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="24">
                <el-form-item label="排除片段">
                  <el-input v-model="editForm.linkExcludeText" placeholder="/tag/, /page/, /feed" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="24">
                <el-form-item label="路径正则">
                  <el-input
                    v-model="editForm.linkPreferRegex"
                    clearable
                    placeholder="可选，Go 正则；如 /[0-9]{4}/ 匹配路径中的年份段"
                  />
                  <p class="field-hint">
                    可选。对<strong>文章详情 URL</strong>做优先匹配（非列表翻页）。留空则仅依赖「链接须含」与「排除片段」。列表
                    <code>/page/2/</code> 由下方「路径模板」生成；排除片段含 <code>/page/</code> 可避免误采翻页链。
                  </p>
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="分页方式">
                  <el-select v-model="editForm.paginationType" style="width: 100%">
                    <el-option label="仅首页" value="none" />
                    <el-option label="路径分页 (/page/2/)" value="path" />
                    <el-option label="WordPress AJAX 加载更多" value="wordpress_ajax" />
                    <el-option label="滚动加载（需 Playwright）" value="scroll" disabled />
                  </el-select>
                  <p v-if="editForm.paginationType === 'wordpress_ajax'" class="field-hint">
                    用于 Unit42 等 WordPress 标签页：列表由 <code>admin-ajax.php</code> 分页加载，无需填写路径模板。
                  </p>
                </el-form-item>
              </el-col>
              <el-col v-if="editForm.paginationType === 'path'" :span="12">
                <el-form-item label="空页停止">
                  <el-switch v-model="editForm.stopOnEmptyList" />
                </el-form-item>
              </el-col>
            </el-row>
            <template v-if="editForm.paginationType === 'path'">
              <el-row :gutter="16">
                <el-col :span="24">
                  <el-form-item label="路径模板">
                    <div class="path-template-row">
                      <el-input
                        v-model="editForm.pathTemplate"
                        placeholder="{列表URL去尾斜杠}/page/{page}/"
                      />
                      <el-button type="primary" plain @click="applyPathTemplateFromList">按列表 URL 生成</el-button>
                    </div>
                    <p class="field-hint">
                      占位符 <code>{page}</code> 会替换为 2、3、4…，生成
                      <code>/page/2/</code>、<code>/page/3/</code>。第 1 页用「列表 URL」，不从模板生成。示例：起始页
                      2、追加页数 2 → 访问 <code>…/page/2/</code> 与 <code>…/page/3/</code>。
                    </p>
                  </el-form-item>
                </el-col>
              </el-row>
              <el-row :gutter="16">
                <el-col :span="12">
                  <el-form-item label="起始页">
                    <el-input-number v-model="editForm.startPage" :min="2" :max="100" class="num-full" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="追加页数">
                    <el-input-number v-model="editForm.maxPages" :min="1" :max="20" class="num-full" />
                    <p class="field-hint field-hint--inline">从起始页起连续 N 页（非「末页页码」）</p>
                  </el-form-item>
                </el-col>
              </el-row>
            </template>
            <el-row v-if="formType === 'threat_portal'" :gutter="16">
              <el-col :span="24">
                <el-alert
                  type="info"
                  :closable="false"
                  show-icon
                  title="威胁门户与分页列表共用配置；滚动分页将在后续版本通过 Playwright 支持。"
                />
              </el-col>
            </el-row>
          </template>
          <template v-if="formType !== 'manual_upload'">
            <el-divider content-position="left">标题解析</el-divider>
            <el-row :gutter="16">
              <el-col :span="24">
                <el-form-item label="标题黑名单">
                  <el-input
                    v-model="editForm.titleBlocklistText"
                    type="textarea"
                    :rows="2"
                    placeholder="每行或逗号分隔；命中则不作为文章标题，如：奇安信威胁情报中心"
                  />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="24">
                <el-form-item label="标题正则">
                  <el-input
                    v-model="editForm.titleExtractRegexText"
                    type="textarea"
                    :rows="4"
                    placeholder="每行一条 Go 正则，须含捕获组；按顺序匹配 HTML，优先于 &lt;title&gt;"
                  />
                  <p class="field-hint">
                    写入采集源 <code>config_json.title_extract</code>。Nuxt 博客示例（奇安信 TI）：<br />
                    <code>(?is)e\.title\s*=\s*"((?:\\.|[^"\\])*)"</code><br />
                    <code>(?is)class="title-detail"[^&gt;]*&gt;\s*([^&lt;]+?)(?:\s*&lt;/|\s+By\s)</code>
                  </p>
                </el-form-item>
              </el-col>
            </el-row>
          </template>
          <el-row v-if="formType !== 'manual_upload'" :gutter="16">
            <el-col :span="12">
              <el-form-item label="单次上限">
                <el-input-number v-model="editForm.maxItems" :min="1" :max="50" class="num-full" />
              </el-form-item>
            </el-col>
            <el-col v-if="isListType" :span="12">
              <el-form-item label="间隔 ms">
                <el-input-number v-model="editForm.delayMs" :min="0" :max="10000" :step="100" class="num-full" />
              </el-form-item>
            </el-col>
          </el-row>

        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="save">保存</el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="pilotDialogVisible"
        :title="pilotRow ? `上线前测试 — ${pilotRow.name}` : '上线前测试'"
        width="640px"
        align-center
        append-to-body
        destroy-on-close
        class="collect-pilot-dialog"
      >
        <template v-if="pilotRow">
          <p class="field-hint pilot-intro">
            基于当前已保存的配置试采（与「编辑」分离）。顺序：单条文章 → 分页/列表（若需要）→ 解锁后列表「运行」与定时调度生效。
          </p>
          <el-form label-width="88px">
            <el-form-item label="试点 URL">
              <el-input
                v-model="pilotArticleUrl"
                placeholder="代表性文章页；直采型可留空用配置首条 URL"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="pilotArticleLoading" @click="runPilotArticle">
                ① 单条试采
              </el-button>
              <el-button
                v-if="pilotRow.needsListPilot"
                type="warning"
                plain
                :loading="pilotListLoading"
                @click="runPilotList"
              >
                ② 分页试采
              </el-button>
            </el-form-item>
          </el-form>
          <div v-if="lastArticleChecks.length" class="pilot-checks">
            <div
              v-for="c in lastArticleChecks"
              :key="c.id"
              class="pilot-check-row"
              :class="{ ok: c.ok, fail: !c.ok }"
            >
              <span>{{ c.ok ? '✓' : '✗' }}</span>
              <span>{{ c.label }}</span>
              <span v-if="c.detail" class="pilot-check-detail">{{ c.detail }}</span>
            </div>
          </div>
          <div v-if="pilotListCount > 0" class="pilot-list-preview">
            <p class="field-hint">发现 {{ pilotListCount }} 条链接，样例：</p>
            <ul>
              <li v-for="u in pilotListPreview" :key="u">{{ u }}</li>
            </ul>
          </div>
          <div class="pilot-dialog-footer">
            <el-space wrap>
              <el-tag :type="pilotArticleDone(pilotRow) ? 'success' : 'info'" size="small">
                单条 {{ pilotArticleDone(pilotRow) ? '已通过' : '未完成' }}
              </el-tag>
              <el-tag
                v-if="pilotRow.needsListPilot"
                :type="pilotListDone(pilotRow) ? 'success' : 'info'"
                size="small"
              >
                分页 {{ pilotListDone(pilotRow) ? '已通过' : '未完成' }}
              </el-tag>
              <el-tag :type="pilotRow.batchUnlocked ? 'success' : 'warning'" size="small">
                批量 {{ pilotRow.batchUnlocked ? '已解锁' : '未解锁' }}
              </el-tag>
            </el-space>
            <el-button
              type="success"
              :disabled="!canUnlockBatch(pilotRow)"
              :loading="pilotUnlockLoading"
              @click="runUnlockBatch"
            >
              ③ 解锁批量运行
            </el-button>
          </div>
        </template>
        <template #footer>
          <el-button @click="pilotDialogVisible = false">关闭</el-button>
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
.collect-source-form :deep(.el-form-item) {
  margin-bottom: 14px;
}
.collect-source-form :deep(.num-full) {
  width: 100%;
}
.collect-source-form :deep(.el-input-number.num-full .el-input__wrapper) {
  width: 100%;
}
.field-hint {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}
.field-hint code {
  font-size: 11px;
  padding: 0 4px;
  background: var(--el-fill-color-light);
  border-radius: 3px;
}
.pilot-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}
.pilot-cell-muted {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
.pilot-open-btn {
  margin-top: 2px;
}
.pilot-dialog-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
}
.pilot-intro {
  margin-bottom: 12px;
}
.pilot-checks {
  margin: 0 0 12px 12px;
  font-size: 12px;
}
.pilot-check-row {
  display: flex;
  gap: 8px;
  line-height: 1.6;
  color: var(--el-color-danger);
}
.pilot-check-row.ok {
  color: var(--el-color-success);
}
.pilot-check-detail {
  color: var(--el-text-color-secondary);
}
.pilot-list-preview ul {
  margin: 4px 0 12px 20px;
  font-size: 12px;
  word-break: break-all;
}
.pilot-list-row {
  margin-top: 8px;
}
.path-template-row {
  display: flex;
  gap: 8px;
  width: 100%;
}
.path-template-row .el-input {
  flex: 1;
}
.field-hint--inline {
  margin-top: 4px;
}
</style>
