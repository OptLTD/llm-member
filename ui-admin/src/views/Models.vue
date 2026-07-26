<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { NCard, NDataTable, NSelect, NTag, type DataTableColumns } from 'naive-ui'
import http from '../api/http'
import ListToolbar from '../components/ListToolbar.vue'

interface ModelRow {
  id: string
  name: string
  provider: string
}

const models = ref<ModelRow[]>([])
const loading = ref(false)
const provider = ref('all')

async function load() {
  loading.value = true
  try {
    const { data } = await http.get('/api/admin/models')
    models.value = data.data || []
  } finally {
    loading.value = false
  }
}

const providers = computed(() => {
  const set = new Set(models.value.map((m) => m.provider).filter(Boolean))
  return [
    { label: '全部提供商', value: 'all' },
    ...Array.from(set).map((p) => ({ label: p, value: p })),
  ]
})

const filtered = computed(() => {
  if (provider.value === 'all') return models.value
  return models.value.filter((m) => m.provider === provider.value)
})

const columns: DataTableColumns<ModelRow> = [
  { title: '模型 ID', key: 'id', ellipsis: { tooltip: true } },
  { title: '名称', key: 'name' },
  {
    title: '提供商',
    key: 'provider',
    render: (row) =>
      h(NTag, { type: 'info', size: 'small', bordered: false }, { default: () => row.provider }),
  },
  {
    title: '状态',
    key: 'status',
    render: () =>
      h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '可用' }),
  },
]

onMounted(load)
</script>

<template>
  <n-card>
    <ListToolbar :loading="loading" @refresh="load">
      <n-select v-model:value="provider" :options="providers" style="width: 200px" />
      <span class="count">共 {{ filtered.length }} 个模型</span>
    </ListToolbar>
    <n-data-table
      :columns="columns"
      :data="filtered"
      :loading="loading"
      :bordered="false"
      striped
    />
  </n-card>
</template>

<style scoped>
.count {
  font-size: 13px;
  color: #6b7280;
}
</style>
