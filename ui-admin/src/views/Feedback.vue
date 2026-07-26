<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import {
  NCard,
  NGrid,
  NGi,
  NStatistic,
  NDataTable,
  NButton,
  NSpace,
  type DataTableColumns,
} from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'
import ListToolbar from '../components/ListToolbar.vue'

const list = ref<any[]>([])
const summary = ref<any>({})
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [l, s] = await Promise.all([
      http.get('/api/admin/feedbacks'),
      http.get('/api/admin/feedbacks/summary'),
    ])
    list.value = l.data.feedbacks || []
    summary.value = s.data || {}
  } finally {
    loading.value = false
  }
}

async function update(row: any, status: string) {
  await http.put(`/api/admin/feedbacks/${row.id}`, { status })
  notify.success('已更新')
  load()
}

const columns: DataTableColumns<any> = [
  { title: '类型', key: 'type', width: 100 },
  { title: '标题', key: 'title' },
  { title: '评分', key: 'rating', width: 70 },
  { title: '状态', key: 'status', width: 100 },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row) =>
      h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => update(row, 'processing') }, { default: () => '处理中' }),
          h(
            NButton,
            { size: 'small', type: 'success', onClick: () => update(row, 'closed') },
            { default: () => '关闭' },
          ),
        ],
      }),
  },
]

onMounted(load)
</script>

<template>
  <div>
    <n-grid cols="1 s:3" responsive="screen" :x-gap="8" :y-gap="8" style="margin-bottom: 8px">
      <n-gi><n-card><n-statistic label="待处理" :value="summary.open || 0" /></n-card></n-gi>
      <n-gi><n-card><n-statistic label="已关闭" :value="summary.closed || 0" /></n-card></n-gi>
      <n-gi><n-card><n-statistic label="平均评分" :value="summary.avgRating || 0" /></n-card></n-gi>
    </n-grid>
    <n-card>
      <ListToolbar :loading="loading" @refresh="load" />
      <n-data-table :columns="columns" :data="list" :loading="loading" :bordered="false" striped />
    </n-card>
  </div>
</template>
