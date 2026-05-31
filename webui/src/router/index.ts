import { createRouter, createWebHistory } from 'vue-router'

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
  ],
})

export default router
