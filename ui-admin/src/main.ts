import { createApp } from 'vue'
import { createPinia } from 'pinia'
import naive from 'naive-ui'
import App from './App.vue'
import router from './router'
import './style.css'

createApp(App).use(createPinia()).use(router).use(naive).mount('#app')
