<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCard, NDataTable, NInput, NButton, type DataTableColumns } from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'
import ListToolbar from '../components/ListToolbar.vue'

const surveys = ref<any[]>([])
const loading = ref(false)
const form = ref({ title: '', description: '', status: 'draft' })

async function load() {
  loading.value = true
  try {
    const { data } = await http.get('/api/admin/surveys')
    surveys.value = data.surveys || []
  } finally {
    loading.value = false
  }
}

async function create() {
  await http.post('/api/admin/surveys', form.value)
  notify.success('问卷已创建')
  form.value = { title: '', description: '', status: 'draft' }
  load()
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '标题', key: 'title' },
  { title: '状态', key: 'status' },
]

onMounted(load)
</script>

<template>
  <n-card>
    <ListToolbar :loading="loading" @refresh="load">
      <n-input v-model:value="form.title" placeholder="标题" style="width: 200px" />
      <n-input v-model:value="form.description" placeholder="描述" style="width: 260px" />
      <n-button type="primary" @click="create">创建</n-button>
    </ListToolbar>
    <n-data-table :columns="columns" :data="surveys" :loading="loading" :bordered="false" striped />
  </n-card>
</template>
