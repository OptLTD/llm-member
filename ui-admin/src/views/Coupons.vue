<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  NCard,
  NDataTable,
  NInput,
  NInputNumber,
  NSelect,
  NButton,
  type DataTableColumns,
} from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'
import ListToolbar from '../components/ListToolbar.vue'

const coupons = ref<any[]>([])
const loading = ref(false)
const form = ref({ code: '', type: 'percent', value: 10, maxUses: 100, status: 'active' })

async function load() {
  loading.value = true
  try {
    const { data } = await http.get('/api/admin/coupons')
    coupons.value = data.coupons || []
  } finally {
    loading.value = false
  }
}

async function create() {
  await http.post('/api/admin/coupons', form.value)
  notify.success('券码已创建')
  load()
}

const typeOptions = [
  { label: '百分比', value: 'percent' },
  { label: '固定', value: 'fixed' },
  { label: '延期', value: 'plan_extend' },
]

const columns: DataTableColumns<any> = [
  { title: '券码', key: 'code' },
  { title: '类型', key: 'type' },
  { title: '面值', key: 'value' },
  { title: '已用', key: 'usedCount' },
  { title: '上限', key: 'maxUses' },
  { title: '状态', key: 'status' },
]

onMounted(load)
</script>

<template>
  <n-card>
    <ListToolbar :loading="loading" @refresh="load">
      <n-input v-model:value="form.code" placeholder="券码" style="width: 140px" />
      <n-select v-model:value="form.type" :options="typeOptions" style="width: 120px" />
      <n-input-number v-model:value="form.value" placeholder="面值" style="width: 110px" />
      <n-input-number v-model:value="form.maxUses" placeholder="上限" style="width: 110px" />
      <n-button type="primary" @click="create">创建</n-button>
    </ListToolbar>
    <n-data-table :columns="columns" :data="coupons" :loading="loading" :bordered="false" striped />
  </n-card>
</template>
