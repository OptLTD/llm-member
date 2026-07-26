<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCard, NDataTable, type DataTableColumns } from 'naive-ui'
import http from '../api/http'
import ListToolbar from '../components/ListToolbar.vue'

const referrals = ref<any[]>([])
const campaigns = ref<any[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [r, c] = await Promise.all([
      http.get('/api/admin/referral/list'),
      http.get('/api/admin/referral/campaigns'),
    ])
    referrals.value = r.data.referrals || []
    campaigns.value = c.data.campaigns || []
  } finally {
    loading.value = false
  }
}

const referralColumns: DataTableColumns<any> = [
  { title: '邀请人', key: 'inviterId' },
  { title: '被邀请人', key: 'inviteeId' },
  { title: '状态', key: 'status' },
  { title: '奖励', key: 'rewardStatus' },
]

const campaignColumns: DataTableColumns<any> = [
  { title: '名称', key: 'name' },
  { title: '奖励类型', key: 'rewardType' },
  { title: '启用', key: 'enabled' },
]

onMounted(load)
</script>

<template>
  <div>
    <n-card title="邀请关系" style="margin-bottom: 8px">
      <ListToolbar :loading="loading" @refresh="load" />
      <n-data-table
        :columns="referralColumns"
        :data="referrals"
        :loading="loading"
        :bordered="false"
        striped
      />
    </n-card>
    <n-card title="活动">
      <ListToolbar :loading="loading" @refresh="load" />
      <n-data-table
        :columns="campaignColumns"
        :data="campaigns"
        :loading="loading"
        :bordered="false"
        striped
      />
    </n-card>
  </div>
</template>
