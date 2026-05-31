<script setup lang="ts">
import { computed, markRaw, onMounted, ref, watch, type Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Collection,
  Connection,
  Cpu,
  DataLine,
  Document,
  Download,
  List,
  Reading,
  Setting,
} from '@element-plus/icons-vue'
import { THREAT_LIB_NAV_ITEMS } from '@/constants/threat-lib-nav'

const route = useRoute()
const router = useRouter()

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

function toggleThreatLib() {
  threatLibOpen.value = !threatLibOpen.value
}

function toggleSettings() {
  settingsOpen.value = !settingsOpen.value
}

const primaryNavItems = [
  { path: '/dashboard', label: '运营总览', icon: markRaw(DataLine) },
  { path: '/knowledge', label: '攻击知识库', icon: markRaw(Reading) },
]

const ingestQueueNav = { path: '/intel', label: '采集队列', icon: markRaw(List) }

const settingsChildren: { path: string; label: string; icon: Component }[] = [
  { path: '/settings/collect', label: '采集配置', icon: markRaw(Download) },
  { path: '/settings/parser', label: '解析模板', icon: markRaw(Document) },
  { path: '/settings/provider', label: '模型服务', icon: markRaw(Cpu) },
]

const activeTop = computed(() => {
  if (route.path === '/dashboard') return '/dashboard'
  if (route.path.startsWith('/knowledge')) return '/knowledge'
  if (route.path.startsWith('/intel')) return '/intel'
  return ''
})

const activeThreatDim = computed(() => {
  if (route.path.startsWith('/analytics/actor/')) return '/analytics/actors'
  if (route.path.startsWith('/analytics/entity/')) {
    const dim = route.params.dimension as string
    if (dim === 'tool') return '/analytics/tools'
    if (dim) return `/analytics/${dim}`
  }
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

/** Warm route chunk before user picks a dimension (does not block menu toggle). */
onMounted(() => {
  void import('@/views/AnalyticsHub.vue')
})
</script>

<template>
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
        v-for="item in primaryNavItems"
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
        <button type="button" class="nav-item nav-group-head" @click="toggleThreatLib">
          <el-icon class="nav-icon"><Collection /></el-icon>
          <span class="nav-label">威胁情报库</span>
          <span class="nav-chevron" :class="{ open: threatLibOpen }">›</span>
        </button>
        <div class="nav-sub">
          <button
            v-for="child in THREAT_LIB_NAV_ITEMS"
            :key="child.key"
            type="button"
            class="nav-sub-item"
            :class="{ active: activeThreatDim === `/analytics/${child.key}` }"
            @click="go(`/analytics/${child.key}`)"
          >
            <el-icon class="sub-icon"><component :is="child.icon" /></el-icon>
            <span class="sub-label">{{ child.label }}</span>
          </button>
        </div>
      </div>

      <button
        type="button"
        class="nav-item"
        :class="{ active: isTopActive(ingestQueueNav.path) }"
        @click="go(ingestQueueNav.path)"
      >
        <el-icon class="nav-icon"><component :is="ingestQueueNav.icon" /></el-icon>
        <span class="nav-label">{{ ingestQueueNav.label }}</span>
      </button>

      <div class="nav-group" :class="{ open: settingsOpen, active: settingsGroupActive }">
        <button type="button" class="nav-item nav-group-head" @click="toggleSettings">
          <el-icon class="nav-icon"><Setting /></el-icon>
          <span class="nav-label">系统配置</span>
          <span class="nav-chevron" :class="{ open: settingsOpen }">›</span>
        </button>
        <div class="nav-sub">
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
</template>

<style scoped>
.aside {
  background: var(--ami-sidebar);
  color: #e2e8f0;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 100vh;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  contain: layout style paint;
  flex-shrink: 0;
}
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 20px 16px;
  flex-shrink: 0;
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
  overflow-x: hidden;
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
  transition: background 0.12s, color 0.12s;
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
  transition: transform 0.12s;
  opacity: 0.7;
}
.nav-chevron.open {
  transform: rotate(90deg);
}
.nav-group:not(.open) > .nav-sub {
  display: none;
}
.nav-group.open > .nav-sub {
  display: flex;
  flex-direction: column;
  gap: 2px;
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
  transition: background 0.12s, color 0.12s;
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
  flex-shrink: 0;
}
</style>
