<script setup lang="ts">
import AppLayout from '@/components/AppLayout.vue'
import AnalyticsDashboard from '@/components/analytics/AnalyticsDashboard.vue'
import PageShell from '@/components/PageShell.vue'
</script>

<template>
  <AppLayout>
    <PageShell wide>
      <AnalyticsDashboard />
    </PageShell>
  </AppLayout>
</template>
