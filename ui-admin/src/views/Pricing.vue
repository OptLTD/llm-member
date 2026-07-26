<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
  NButton,
  NTag,
  NSpace,
} from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'

interface Plan {
  plan: string
  name: string
  brief: string
  price: number
  usage: string
  period: string
  enabled: boolean
  features: string[]
  quotaType?: string
  maxDevices?: number
  /** UI only: features as newline text */
  featuresText?: string
}

const plans = ref<Plan[]>([])
const saving = ref<string | null>(null)

const ordered = computed(() => {
  const order = ['basic', 'extra', 'ultra', 'super']
  return [...plans.value].sort(
    (a, b) => order.indexOf(a.plan) - order.indexOf(b.plan) || a.plan.localeCompare(b.plan),
  )
})

async function load() {
  const { data } = await http.get('/api/setup/pricing')
  plans.value = (data.plans || []).map((p: Plan) => ({
    ...p,
    features: Array.isArray(p.features) ? p.features : [],
    featuresText: (Array.isArray(p.features) ? p.features : []).join('\n'),
    maxDevices: p.maxDevices ?? 0,
    quotaType: p.quotaType || 'tokens',
  }))
}

async function save(plan: Plan) {
  saving.value = plan.plan
  try {
    const payload = {
      plan: plan.plan,
      name: plan.name,
      brief: plan.brief,
      price: plan.price,
      usage: plan.usage,
      period: plan.period,
      enabled: !!plan.enabled,
      features: (plan.featuresText || '')
        .split('\n')
        .map((s) => s.trim())
        .filter(Boolean),
      quotaType: plan.quotaType || 'tokens',
      maxDevices: plan.maxDevices ?? 0,
    }
    await http.put(`/api/setup/pricing/${plan.plan}`, payload)
    plan.features = payload.features
    notify.success('已保存')
  } catch (e: any) {
    notify.error(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = null
  }
}

onMounted(load)
</script>

<template>
  <div class="plans-grid">
    <n-card v-for="plan in ordered" :key="plan.plan" class="plan-card" :class="{ disabled: !plan.enabled }">
      <template #header>
        <div class="card-header">
          <div class="title-wrap">
            <span class="title">{{ plan.name || plan.plan }}</span>
            <n-tag size="small" :bordered="false" type="default">{{ plan.plan }}</n-tag>
          </div>
          <n-space align="center" :size="8">
            <span class="switch-label">{{ plan.enabled ? '已启用' : '已禁用' }}</span>
            <n-switch v-model:value="plan.enabled" size="small" />
          </n-space>
        </div>
      </template>

      <n-form label-placement="top" size="small">
        <n-form-item label="名称">
          <n-input v-model:value="plan.name" />
        </n-form-item>
        <n-form-item label="简介">
          <n-input v-model:value="plan.brief" type="textarea" :rows="2" />
        </n-form-item>
        <div class="row-2">
          <n-form-item label="价格">
            <n-input-number v-model:value="plan.price" :min="0" class="full" />
          </n-form-item>
          <n-form-item label="周期">
            <n-input v-model:value="plan.period" placeholder="如 1d / 7d / 1m / 1y" />
          </n-form-item>
        </div>
        <n-form-item label="用量">
          <n-input v-model:value="plan.usage" placeholder="如 1k tokens / 2k requests" />
        </n-form-item>
        <div class="row-2">
          <n-form-item label="配额类型">
            <n-input v-model:value="plan.quotaType" placeholder="tokens | generic" />
          </n-form-item>
          <n-form-item label="设备上限">
            <n-input-number v-model:value="plan.maxDevices" :min="0" class="full" placeholder="0=不限" />
          </n-form-item>
        </div>
        <n-form-item label="特性（每行一条）">
          <n-input v-model:value="plan.featuresText" type="textarea" :rows="4" />
        </n-form-item>
        <n-button type="primary" block :loading="saving === plan.plan" @click="save(plan)">
          保存
        </n-button>
      </n-form>
    </n-card>
  </div>
</template>

<style scoped>
.plans-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
  align-items: start;
}

.plans-grid :deep(.n-card) {
  margin-top: 0 !important;
  align-self: start;
  width: 100%;
}

@media (max-width: 1280px) {
  .plans-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .plans-grid {
    grid-template-columns: 1fr;
  }
}

.plan-card.disabled {
  opacity: 0.72;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.title {
  font-weight: 600;
  font-size: 15px;
  color: #111827;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.switch-label {
  font-size: 12px;
  color: #6b7280;
  white-space: nowrap;
}

.row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.full {
  width: 100%;
}
</style>
