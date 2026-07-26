<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NDataTable, NCard, NButton, type DataTableColumns } from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'
import ListToolbar from '../components/ListToolbar.vue'

const users = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await http.post('/api/admin/users', { page: page.value, size: 20 })
    users.value = data.data || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

async function toggle(id: number) {
  await http.post(`/api/admin/users/${id}/toggle`)
  notify.success('状态已更新')
  load()
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户名', key: 'username' },
  { title: '邮箱', key: 'email' },
  { title: '套餐', key: 'userPlan' },
  {
    title: '状态',
    key: 'isActive',
    render: (row) => (row.isActive ? '正常' : '禁用'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row) =>
      h(NButton, { size: 'small', onClick: () => toggle(row.id) }, { default: () => '切换状态' }),
  },
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
    <n-data-table :columns="columns" :data="users" :loading="loading" :bordered="false" striped />
  </n-card>
</template>
