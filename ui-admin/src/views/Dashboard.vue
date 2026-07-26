<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NGrid, NGi, NStatistic, NCard } from 'naive-ui'
import http from '../api/http'

const stats = ref<any>({})

onMounted(async () => {
  const { data } = await http.get('/api/admin/stats')
  stats.value = data
})
</script>

<template>
  <div>
        <n-grid cols="1 s:2 l:4" responsive="screen" :x-gap="8" :y-gap="8">
      <n-gi><n-card><n-statistic label="总会员" :value="stats.totalMembers || 0" /></n-card></n-gi>
      <n-gi><n-card><n-statistic label="付费会员" :value="stats.paidMembers || 0" /></n-card></n-gi>
      <n-gi><n-card><n-statistic label="总订单" :value="stats.totalOrders || 0" /></n-card></n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="总收入" :value="stats.totalRevenue || 0">
            <template #prefix>¥</template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi><n-card><n-statistic label="本月新增" :value="stats.monthlyNewMembers || 0" /></n-card></n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="本月收入" :value="stats.monthlyRevenue || 0">
            <template #prefix>¥</template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi><n-card><n-statistic label="API 请求(月)" :value="stats.totalRequests || 0" /></n-card></n-gi>
      <n-gi>
        <n-card>
          <n-statistic label="成功率" :value="stats.successRate || 0">
            <template #suffix>%</template>
          </n-statistic>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>
