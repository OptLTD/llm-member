<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCard, NDataTable, type DataTableColumns } from 'naive-ui'
import http from '../api/http'
import ListToolbar from '../components/ListToolbar.vue'

const logs = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await http.post('/api/admin/logs', { page: page.value, size: 20 })
    logs.value = data.data || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

const columns: DataTableColumns<any> = [
  { title: '模型', key: 'model' },
  { title: '提供商', key: 'provider' },
  { title: '状态', key: 'status', width: 90 },
  { title: '耗时(ms)', key: 'duration', width: 100 },
  { title: '时间', key: 'reqTime' },
]

onMounted(load)
</script>

<template>
  <n-card>
    <ListToolbar
      v-model:page="page"
      :item-count="total"
      :loading="loading"
      show-pagination
      @refresh="load"
    />
    <n-data-table :columns="columns" :data="logs" :loading="loading" :bordered="false" striped />
  </n-card>
</template>
