<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import {
  NCard, NDataTable, NButton, NModal, NForm,
  NInput, NSwitch, NSpace, NTag, NDivider, NGrid, NFormItemGi, useDialog,
} from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'
import { useProjectsStore } from '../stores/projects'

const dialog = useDialog()
const projects = useProjectsStore()

const loading = ref(false)
const showModal = ref(false)
const editing = ref(false)
const form = ref<any>({
  id: 0, name: '', slug: '', status: 'active',
  settings: { appName: '', authCallback: '', llm: { provider: '', apiKey: '', baseUrl: '', defaultModel: '', models: '' } },
})

function renderStatus(row: any) {
  if (row.status === 'active') {
    return h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '启用' })
  }
  return h(NTag, { type: 'warning', size: 'small', bordered: false }, { default: () => '停用' })
}

function renderActions(row: any) {
  return h(NSpace, { size: 4 }, {
    default: () => [
      h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
      h(NButton, { size: 'small', type: 'error', quaternary: true, onClick: () => remove(row) }, { default: () => '删除' }),
    ],
  })
}

const columns = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name' },
  { title: 'Slug', key: 'slug' },
  { title: '状态', key: 'status', render: renderStatus },
  { title: '操作', key: 'actions', render: renderActions },
]

function openEdit(row: any) {
  editing.value = true
  const st = row.settings || {}
  const llm = st.llm || {}
  form.value = {
    id: row.id, name: row.name, slug: row.slug, status: row.status,
    settings: {
      appName: st.appName || '',
      authCallback: st.authCallback || '',
      llm: {
        provider: llm.provider || '', apiKey: llm.apiKey || '',
        baseUrl: llm.baseUrl || '', defaultModel: llm.defaultModel || '',
        models: Array.isArray(llm.models) ? llm.models.join(',') : (llm.models || ''),
      },
    },
  }
  showModal.value = true
}

function openCreate() {
  editing.value = false
  form.value = { id: 0, name: '', slug: '', status: 'active', settings: { appName: '', authCallback: '', llm: { provider: '', apiKey: '', baseUrl: '', defaultModel: '', models: '' } } }
  showModal.value = true
}

async function save() {
  loading.value = true
  try {
    const payload = {
      name: form.value.name,
      slug: form.value.slug,
      status: form.value.status,
      appName: form.value.settings.appName,
      authCallback: form.value.settings.authCallback,
      llm: {
        provider: form.value.settings.llm.provider,
        apiKey: form.value.settings.llm.apiKey,
        baseUrl: form.value.settings.llm.baseUrl,
        defaultModel: form.value.settings.llm.defaultModel,
        models: (form.value.settings.llm.models || '').split(',').map((s: string) => s.trim()).filter(Boolean),
      },
    }
    if (editing.value) {
      await http.put(`/api/admin/projects/${form.value.id}`, payload)
      if (payload.appName || payload.authCallback || payload.llm.apiKey) {
        await http.put(`/api/admin/projects/${form.value.id}/settings`, {
          appName: payload.appName,
          authCallback: payload.authCallback,
          llm: payload.llm,
        })
      }
    } else {
      const { data } = await http.post('/api/admin/projects', payload)
      if (payload.appName || payload.authCallback || payload.llm.apiKey) {
        await http.put(`/api/admin/projects/${data.project.id}/settings`, {
          appName: payload.appName,
          authCallback: payload.authCallback,
          llm: payload.llm,
        })
      }
    }
    notify.success('已保存')
    showModal.value = false
    await projects.load()
  } catch (e: any) {
    notify.error(e.response?.data?.error || '保存失败')
  } finally {
    loading.value = false
  }
}

function remove(row: any) {
  dialog.warning({
    title: '删除项目',
    content: `确定删除项目「${row.name}」吗？其下数据将不可见。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await http.delete(`/api/admin/projects/${row.id}`)
        notify.success('已删除')
        await projects.load()
      } catch (e: any) {
        notify.error(e.response?.data?.error || '删除失败')
      }
    },
  })
}

onMounted(async () => {
  await projects.load()
})
</script>

<template>
  <n-card title="项目管理" style="margin-top: 0">
    <template #header-extra>
      <n-button type="primary" size="small" @click="openCreate">新建项目</n-button>
    </template>
    <n-data-table :columns="columns" :data="projects.projects" :bordered="false" size="small" />
  </n-card>

  <n-modal v-model:show="showModal" preset="card" :title="editing ? '编辑项目' : '新建项目'" style="width: 640px">
    <n-form label-placement="top" size="small">
      <n-grid :cols="2" :x-gap="16" responsive="screen">
        <n-form-item-gi label="名称">
          <n-input v-model:value="form.name" placeholder="项目名称" />
        </n-form-item-gi>
        <n-form-item-gi label="Slug（唯一标识）">
          <n-input v-model:value="form.slug" placeholder="如 my-app" :disabled="editing" />
        </n-form-item-gi>
      </n-grid>

      <n-divider>项目设置</n-divider>
      <n-grid :cols="2" :x-gap="16" responsive="screen">
        <n-form-item-gi label="应用名称（AppName）">
          <n-input v-model:value="form.settings.appName" placeholder="覆盖全局 APP_NAME" />
        </n-form-item-gi>
        <n-form-item-gi label="认证回调 URL">
          <n-input v-model:value="form.settings.authCallback" placeholder="覆盖全局 AUTH_CALLBACK" />
        </n-form-item-gi>
      </n-grid>

      <n-divider>LLM 上游（覆盖全局 LLM_*）</n-divider>
      <n-grid :cols="2" :x-gap="16" responsive="screen">
        <n-form-item-gi label="Provider">
          <n-input v-model:value="form.settings.llm.provider" placeholder="如 openai / deepseek" />
        </n-form-item-gi>
        <n-form-item-gi label="API Key">
          <n-input v-model:value="form.settings.llm.apiKey" placeholder="sk-..." />
        </n-form-item-gi>
        <n-form-item-gi label="Base URL">
          <n-input v-model:value="form.settings.llm.baseUrl" placeholder="https://api.openai.com/v1" />
        </n-form-item-gi>
        <n-form-item-gi label="默认模型">
          <n-input v-model:value="form.settings.llm.defaultModel" placeholder="如 gpt-4o-mini" />
        </n-form-item-gi>
        <n-form-item-gi :span="2" label="模型列表（逗号分隔）">
          <n-input v-model:value="form.settings.llm.models" placeholder="gpt-4o-mini,gpt-4o" />
        </n-form-item-gi>
      </n-grid>
    </n-form>
    <template #footer>
      <div class="modal-footer">
        <div class="footer-left">
          <n-switch
            v-model:value="form.status"
            checked-value="active"
            unchecked-value="disabled"
            size="small"
          />
          <span class="status-text">{{ form.status === 'active' ? '启用' : '停用' }}</span>
        </div>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="loading" @click="save">保存</n-button>
        </n-space>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-text {
  font-size: 13px;
  color: #6b7280;
}
</style>
