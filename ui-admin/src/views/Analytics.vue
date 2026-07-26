<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCard, NGrid, NGi, NStatistic, NInput, NButton } from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'

const overview = ref<any>({})

async function load() {
  const { data } = await http.get('/api/admin/analytics/overview')
  overview.value = data || {}
}

async function copyEmbed() {
  const code = overview.value.embedCode || ''
  if (!code) return
  await navigator.clipboard.writeText(code)
  notify.success('嵌入代码已复制')
}

onMounted(load)
</script>

<template>
  <div>
    <n-grid cols="1 s:2 l:4" responsive="screen" :x-gap="8" :y-gap="8" style="margin-bottom: 8px">
      <n-gi><n-card><n-statistic label="今日PV" :value="overview.todayPV || 0" /></n-card></n-gi>
      <n-gi><n-card><n-statistic label="今日UV" :value="overview.todayUV || 0" /></n-card></n-gi>
      <n-gi><n-card><n-statistic label="今日事件" :value="overview.todayEvents || 0" /></n-card></n-gi>
      <n-gi><n-card><n-statistic label="7日事件" :value="overview.weekEvents || 0" /></n-card></n-gi>
    </n-grid>
    <n-card title="嵌入代码">
      <p class="hint">单实例部署无需配置 siteKey，引入脚本即可自动上报。</p>
      <n-input type="textarea" :rows="3" readonly :value="overview.embedCode || ''" />
      <n-button type="primary" style="margin-top: 6px" @click="copyEmbed">复制代码</n-button>
    </n-card>
  </div>
</template>

<style scoped>
.hint {
  margin: 0 0 6px;
  color: #6b7280;
  font-size: 13px;
}
</style>
