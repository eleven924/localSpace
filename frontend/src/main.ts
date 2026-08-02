import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useJobsStore } from './store/modules/jobs'
import { useNotificationsStore } from './store/modules/notifications'
import { useThemeStore } from './store/modules/theme'
import './assets/styles/main.css'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)

app.mount('#app')

// Initialize theme after app is mounted
const themeStore = useThemeStore()
themeStore.loadThemeFromBackend()

const jobsStore = useJobsStore()
jobsStore.initialize()

const notificationsStore = useNotificationsStore()
notificationsStore.initialize()
