<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NCard, NDescriptions, NDescriptionsItem, NSwitch, NButton, NSpin } from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'
import { useModulesStore } from '../stores/modules'

const MODULE_LABELS: Record<string, string> = {
  member: '会员 / 支付',
  device: '设备管理',
  analytics: '埋点分析',
  coupon: '优惠券',
  referral: '分享裂变',
  survey: '问卷调研',
  feedback: '用户反馈',
  'llm-relay': 'LLM 中继',
}

const info = ref<any>({})
const available = ref<string[]>([])
const enabled = ref<Set<string>>(new Set())
const saving = ref(false)
const loading = ref(true)
const modulesStore = useModulesStore()

const dirty = computed(() => {
  const current = [...enabled.value].sort().join(',')
  const saved = (info.value.modules || []).slice().sort().join(',')
  return current !== saved
})

async function load() {
  loading.value = true
  try {
    const { data } = await http.get('/api/admin/system')
    info.value = data || {}
    available.value = data.available || []
    enabled.value = new Set(data.modules || [])
  } finally {
    loading.value = false
  }
}

function isOn(name: string) {
  return enabled.value.has(name)
}

function toggle(name: string, on: boolean) {
  const next = new Set(enabled.value)
  if (on) next.add(name)
  else next.delete(name)
  enabled.value = next
}

async function save() {
  saving.value = true
  try {
    const { data } = await http.put('/api/admin/system/modules', {
      modules: [...enabled.value],
    })
    info.value = { ...info.value, modules: data.modules }
    enabled.value = new Set(data.modules || [])
    await modulesStore.load()
    notify.success('模块已保存，立即生效')
  } catch (e: any) {
    notify.error(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <n-spin :show="loading">
    <n-card title="运行信息">
      <n-descriptions label-placement="left" :column="1" bordered>
        <n-descriptions-item label="应用名称">{{ info.appName || '-' }}</n-descriptions-item>
        <n-descriptions-item label="运行模式">{{ info.appMode || '-' }}</n-descriptions-item>
        <n-descriptions-item label="数据库">{{ info.dbDriver || '-' }}</n-descriptions-item>
        <n-descriptions-item label="版本">{{ info.version || '-' }}</n-descriptions-item>
      </n-descriptions>
    </n-card>

    <n-card title="功能模块" style="margin-top: 8px">
      <div class="modules-grid">
        <div v-for="m in available" :key="m" class="module-item">
          <div class="module-text">
            <div class="module-name">{{ MODULE_LABELS[m] || m }}</div>
            <div class="module-id">{{ m }}</div>
          </div>
          <n-switch :value="isOn(m)" @update:value="(v) => toggle(m, v)" />
        </div>
      </div>
      <div class="footer">
        <p class="hint">开关写入数据库，菜单与接口立即生效，无需重启。</p>
        <n-button type="primary" :disabled="!dirty" :loading="saving" @click="save">保存</n-button>
      </div>
    </n-card>
  </n-spin>
</template>

<style scoped>
.modules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 6px;
}

.module-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 6px 8px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fafafa;
}

.module-text {
  min-width: 0;
}

.module-name {
  font-weight: 500;
  font-size: 14px;
  color: #111827;
}

.module-id {
  margin-top: 2px;
  font-size: 12px;
  color: #9ca3af;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 8px;
}

.hint {
  margin: 0;
  font-size: 13px;
  color: #6b7280;
}
</style>
