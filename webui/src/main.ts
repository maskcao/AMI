import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import { hydrateAnalytics } from '@/api/analytics'
import { USE_API } from '@/api/http'
import './styles/main.css'

async function bootstrap() {
  if (USE_API) {
    try {
      await hydrateAnalytics()
    } catch (e) {
      console.error('[ami] analytics bundle preload failed:', e)
    }
  }

  const app = createApp(App)
  app.use(router)
  app.use(ElementPlus, { locale: zhCn })
  app.mount('#app')
}

void bootstrap()
