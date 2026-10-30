import { createRouter, createWebHistory } from 'vue-router'
import { hydrateAnalytics } from '@/api/analytics'
import { USE_API } from '@/api/http'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/dashboard' },
    {
      path: '/knowledge',
      name: 'knowledge-list',
      component: () => import('@/views/KnowledgeList.vue'),
    },
    {
      path: '/knowledge/:id',
      name: 'knowledge-detail',
      component: () => import('@/views/KnowledgeDetail.vue'),
    },
    {
      path: '/dashboard',
      name: 'analytics-dashboard',
      component: () => import('@/views/AnalyticsDashboardPage.vue'),
    },
    {
      path: '/analytics',
      redirect: '/analytics/cve',
    },
    {
      path: '/analytics/overview',
      redirect: '/dashboard',
    },
    {
      path: '/analytics/ioc',
      redirect: '/analytics/infra',
    },
    {
      path: '/analytics/actor/:actorKey',
      name: 'threat-actor-detail',
      component: () => import('@/views/ThreatActorDetailPage.vue'),
    },
    {
      path: '/analytics/entity/:dimension/:entityKey',
      name: 'entity-dimension-detail',
      component: () => import('@/views/EntityDimensionDetailPage.vue'),
    },
    {
      path: '/analytics/:dimension',
      name: 'analytics-dimension',
      component: () => import('@/views/AnalyticsHub.vue'),
    },
    {
      path: '/intel',
      name: 'intel-list',
      component: () => import('@/views/IntelList.vue'),
    },
    {
      path: '/intel/:id',
      name: 'intel-detail',
      component: () => import('@/views/IntelDetail.vue'),
    },
    {
      path: '/ioc',
      name: 'ioc-search',
      component: () => import('@/views/IocSearch.vue'),
    },
    {
      path: '/settings',
      redirect: '/settings/collect',
    },
    {
      path: '/settings/collect',
      name: 'settings-collect',
      component: () => import('@/views/settings/SettingsCollectPage.vue'),
    },
    {
      path: '/settings/parser',
      name: 'settings-parser',
      component: () => import('@/views/settings/SettingsParserPage.vue'),
    },
    {
      path: '/settings/provider',
      name: 'settings-provider',
      component: () => import('@/views/settings/SettingsProviderPage.vue'),
    },
    {
      path: '/settings/model',
      redirect: '/settings/provider',
    },
    {
      path: '/datacenter',
      redirect: '/datacenter/logs',
    },
    {
      path: '/datacenter/logs',
      name: 'datacenter-logs',
      component: () => import('@/views/datacenter/RuntimeLogPage.vue'),
    },
    {
      path: '/datacenter/reports',
      name: 'datacenter-reports',
      component: () => import('@/views/datacenter/IncidentReportPage.vue'),
    },
  ],
})

router.beforeEach(async (to) => {
  if (!USE_API) return true
  const needsBundle =
    to.path === '/dashboard' ||
    to.path.startsWith('/analytics') ||
    to.name === 'threat-actor-detail'
  if (!needsBundle) return true
  try {
    await hydrateAnalytics()
  } catch (e) {
    console.error('[ami] analytics hydrate on navigation failed:', e)
  }
  return true
})

export default router
