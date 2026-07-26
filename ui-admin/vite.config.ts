import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  base: '/admin/',
  build: {
    outDir: resolve(__dirname, '../ui-front/admin/dist'),
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/v1': 'http://localhost:8080',
    },
  },
})
