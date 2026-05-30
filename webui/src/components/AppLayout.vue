<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Aim,
  Collection,
  Connection,
  Cpu,
  DataLine,
  Document,
  Download,
  OfficeBuilding,
  Reading,
  SetUp,
  Setting,
  User,
  Warning,
} from '@element-plus/icons-vue'
import { buildDimensionNav } from '@/analytics/aggregate'

const route = useRoute()
const router = useRouter()

/** Submenus collapsed by default; expand only when route is under that group. */
const threatLibOpen = ref(false)
const settingsOpen = ref(false)

watch(
  () => route.path,
  (path) => {
    threatLibOpen.value = path.startsWith('/analytics/')
    settingsOpen.value = path.startsWith('/settings')
  },
  { immediate: true },
)

const topMenuItems = [
  { path: '/dashboard', label: '运营总览', icon: DataLine },
  { path: '/knowledge', label: '攻击知识库', icon: Reading },
  { path: '/intel', label: '情报工作台', icon: Document },
]

const dimensionIcons: Record<string, typeof Warning> = {
  cve: Warning,
  malware: Cpu,
  actors: User,
  mitre: Aim,
  tools: SetUp,
  infra: OfficeBuilding,
}

const threatLibChildren = buildDimensionNav().filter((d) => d.key !== 'overview')

const settingsChildren = [
  { path: '/settings/collect', label: '采集配置', icon: Download },
  { path: '/settings/parser', label: '解析模板', icon: Document },
  { path: '/settings/provider', label: '模型服务', icon: Cpu },
]

const activeTop = computed(() => {
  if (route.path === '/dashboard') return '/dashboard'
  if (route.path.startsWith('/knowledge')) return '/knowledge'
  if (route.path.startsWith('/intel')) return '/intel'
  return ''
})

const activeThreatDim = computed(() => {
  if (!route.path.startsWith('/analytics/')) return ''
  const dim = route.params.dimension as string
  return dim && dim !== 'overview' ? `/analytics/${dim}` : ''
})

const activeSettingsPath = computed(() => {
  if (route.path.startsWith('/settings/collect')) return '/settings/collect'
  if (route.path.startsWith('/settings/parser')) return '/settings/parser'
  if (route.path.startsWith('/settings/provider') || route.path === '/settings/model') {
    return '/settings/provider'
  }
  return ''
})

const threatLibGroupActive = computed(() => route.path.startsWith('/analytics/'))
const settingsGroupActive = computed(() => route.path.startsWith('/settings'))

function go(path: string) {
  router.push(path)
}

function isTopActive(path: string) {
  return activeTop.value === path
}
</script>

<template>
  <el-container class="app-layout">
    <el-aside width="232px" class="aside">
      <div class="brand">
        <div class="brand-mark">A</div>
        <div class="brand-text">
          <span class="brand-title">AMI</span>
          <span class="brand-sub">攻击知识情报平台</span>
        </div>
      </div>

      <nav class="nav">
        <button
          v-for="item in topMenuItems"
          :key="item.path"
          type="button"
          class="nav-item"
          :class="{ active: isTopActive(item.path) }"
          @click="go(item.path)"
        >
          <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
          <span class="nav-label">{{ item.label }}</span>
        </button>

        <div class="nav-group" :class="{ open: threatLibOpen, active: threatLibGroupActive }">
          <button type="button" class="nav-item nav-group-head" @click="threatLibOpen = !threatLibOpen">
            <el-icon class="nav-icon"><Collection /></el-icon>
            <span class="nav-label">威胁情报库</span>
            <span class="nav-chevron" :class="{ open: threatLibOpen }">›</span>
          </button>
          <div v-show="threatLibOpen" class="nav-sub">
            <button
              v-for="child in threatLibChildren"
              :key="child.key"
              type="button"
              class="nav-sub-item"
              :class="{ active: activeThreatDim === `/analytics/${child.key}` }"
              @click="go(`/analytics/${child.key}`)"
            >
              <el-icon class="sub-icon"><component :is="dimensionIcons[child.key]" /></el-icon>
              <span class="sub-label">{{ child.label }}</span>
              <span class="sub-count">{{ child.count }}</span>
            </button>
          </div>
        </div>

        <div class="nav-group" :class="{ open: settingsOpen, active: settingsGroupActive }">
          <button type="button" class="nav-item nav-group-head" @click="settingsOpen = !settingsOpen">
            <el-icon class="nav-icon"><Setting /></el-icon>
            <span class="nav-label">系统配置</span>
            <span class="nav-chevron" :class="{ open: settingsOpen }">›</span>
          </button>
          <div v-show="settingsOpen" class="nav-sub">
            <button
              v-for="child in settingsChildren"
              :key="child.path"
              type="button"
              class="nav-sub-item"
              :class="{ active: activeSettingsPath === child.path }"
              @click="go(child.path)"
            >
              <el-icon class="sub-icon"><component :is="child.icon" /></el-icon>
              <span class="sub-label">{{ child.label }}</span>
            </button>
          </div>
        </div>
      </nav>

      <div class="aside-footer">
        <el-icon><Connection /></el-icon>
        <span>Demo · Mock 数据</span>
      </div>
    </el-aside>
    <el-main class="main">
      <slot />
    </el-main>
  </el-container>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
}
.aside {
  background: var(--ami-sidebar);
  color: #e2e8f0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
}
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 20px 16px;
}
.brand-mark {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 18px;
  color: #fff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}
.brand-title {
  display: block;
  font-size: 17px;
  font-weight: 700;
  color: #f8fafc;
}
.brand-sub {
  display: block;
  margin-top: 2px;
  font-size: 11px;
  color: #94a3b8;
}
.nav {
  flex: 1;
  min-height: 0;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
}
.nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 11px 14px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: #cbd5e1;
  font-size: 14px;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s, color 0.15s;
}
.nav-item:hover {
  background: var(--ami-sidebar-hover);
  color: #f1f5f9;
}
.nav-item.active {
  background: var(--ami-sidebar-active);
  color: #fff;
  box-shadow: inset 3px 0 0 #3b82f6;
}
.nav-group.active .nav-group-head {
  color: #f1f5f9;
}
.nav-group-head {
  margin-bottom: 2px;
}
.nav-chevron {
  margin-left: auto;
  font-size: 16px;
  transition: transform 0.2s;
  opacity: 0.7;
}
.nav-chevron.open {
  transform: rotate(90deg);
}
.nav-sub {
  padding: 4px 0 8px 8px;
  margin-left: 12px;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
}
.nav-sub-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #94a3b8;
  font-size: 13px;
  cursor: pointer;
  text-align: left;
  transition: background 0.15s, color 0.15s;
}
.nav-sub-item:hover {
  background: var(--ami-sidebar-hover);
  color: #e2e8f0;
}
.nav-sub-item.active {
  background: var(--ami-sidebar-active);
  color: #fff;
}
.sub-icon {
  font-size: 15px;
}
.sub-label {
  flex: 1;
}
.sub-count {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.1);
  font-variant-numeric: tabular-nums;
}
.nav-sub-item.active .sub-count {
  background: rgba(255, 255, 255, 0.2);
}
.nav-icon {
  font-size: 18px;
}
.nav-label {
  flex: 1;
}
.aside-footer {
  padding: 16px 20px;
  font-size: 11px;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 6px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}
.main {
  padding: 0;
  background: var(--ami-bg);
  overflow: auto;
}
</style>
