<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCard, NDataTable, type DataTableColumns } from 'naive-ui'
import http from '../api/http'
import ListToolbar from '../components/ListToolbar.vue'

const orders = ref<any[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await http.post('/api/admin/orders', { page: page.value, size: 20 })
    orders.value = data.data || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

const columns: DataTableColumns<any> = [
  { title: '订单号', key: 'orderId' },
  { title: '用户ID', key: 'userId', width: 90 },
  { title: '套餐', key: 'payPlan' },
  { title: '金额', key: 'amount' },
  { title: '支付方式', key: 'method' },
  { title: '状态', key: 'status' },
  { title: '券码', key: 'couponCode' },
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
    <n-data-table :columns="columns" :data="orders" :loading="loading" :bordered="false" striped />
  </n-card>
</template>
