<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NCard, NDataTable, NInput, NButton, type DataTableColumns } from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'
import ListToolbar from '../components/ListToolbar.vue'

const devices = ref<any[]>([])
const userId = ref('')
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await http.get('/api/admin/devices', {
      params: { userId: userId.value || undefined },
    })
    devices.value = data.devices || []
  } finally {
    loading.value = false
  }
}

async function unbind(id: number) {
  await http.delete(`/api/admin/devices/${id}`)
  notify.success('已解绑')
  load()
}

const columns: DataTableColumns<any> = [
  { title: '用户ID', key: 'userId', width: 90 },
  { title: '设备ID', key: 'deviceId' },
  { title: '名称', key: 'deviceName' },
  { title: '平台', key: 'platform' },
  { title: '最后活跃', key: 'lastSeenAt' },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) =>
      h(
        NButton,
        { size: 'small', type: 'error', onClick: () => unbind(row.id) },
        { default: () => '解绑' },
      ),
  },
]

onMounted(load)
</script>

<template>
  <n-card>
    <ListToolbar :loading="loading" @refresh="load">
      <n-input
        v-model:value="userId"
        placeholder="按用户ID筛选"
        clearable
        style="width: 200px"
        @keyup.enter="load"
        @clear="load"
      />
      <n-button @click="load">筛选</n-button>
    </ListToolbar>
    <n-data-table :columns="columns" :data="devices" :loading="loading" :bordered="false" striped />
  </n-card>
</template>
