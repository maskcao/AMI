<script setup lang="ts">
import { computed } from 'vue'
import EntityTopologyGraph from '@/components/analytics/EntityTopologyGraph.vue'
import { useEntityTopologyNodes, type TopologyProfileSlice } from '@/composables/useEntityTopology'
import type { ThreatActorProfile } from '@/types/threat-actor'

const props = defineProps<{
  profile: ThreatActorProfile
}>()

const topologySource = computed<TopologyProfileSlice | null>(() => {
  const p = props.profile
  return {
    dimension: 'actor',
    knowledgeCount: p.knowledgeCount,
    targetIndustries: p.targetIndustries,
    threatActors: [p.name, ...p.aliases.filter((a) => a !== p.name)],
    techniques: p.techniques,
    tools: p.tools,
    infrastructure: p.infrastructure,
    malware: p.malware,
    vulnerabilities: p.vulnerabilities,
  }
})

const topologyNodes = useEntityTopologyNodes(topologySource)

const centerLabel = computed(() => {
  const n = props.profile.name
  return n.length > 22 ? `${n.slice(0, 20)}…` : n
})
</script>

<template>
  <EntityTopologyGraph
    :center-label="centerLabel"
    center-subtitle="攻击组织"
    :dimension-nodes="topologyNodes"
    :knowledge="profile.knowledge"
  />
</template>
