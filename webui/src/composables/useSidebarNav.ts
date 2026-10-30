import { ref } from 'vue'

// AppLayout 随页面组件重挂载，分组展开状态必须放在组件外才能跨路由保留
const threatLibOpen = ref(false)
const icsOpen = ref(false)
const dataCenterOpen = ref(false)
const settingsOpen = ref(false)

export function useSidebarNav() {
  return { threatLibOpen, icsOpen, dataCenterOpen, settingsOpen }
}
