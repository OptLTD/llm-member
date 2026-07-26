<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { NCard, NForm, NFormItem, NInput, NButton, NSelect, NSwitch, NSpace } from 'naive-ui'
import http from '../api/http'
import { notify } from '../utils/notify'

const model = ref('')
const message = ref('Hello')
const reply = ref('')
const loading = ref(false)
const stream = ref(true)
const modelOptions = ref<{ label: string; value: string }[]>([])
const replyEl = ref<HTMLElement | null>(null)

async function loadModels() {
  try {
    const { data } = await http.get('/api/admin/models')
    const list = (data.data || []) as { id: string; name?: string }[]
    modelOptions.value = list.map((m) => ({
      label: m.name ? `${m.name} (${m.id})` : m.id,
      value: m.id,
    }))
    if (!model.value && list.length) {
      model.value = list[0].id
    }
  } catch {
    // 模型列表失败时仍允许手填
  }
}

async function ensureApiKey(force = false): Promise<string> {
  if (!force) {
    const cached = localStorage.getItem('apiKey')
    if (cached) return cached
  }
  const { data } = await http.post('/api/admin/me/api-key')
  const key = data.apiKey as string
  localStorage.setItem('apiKey', key)
  return key
}

async function scrollReply() {
  await nextTick()
  if (replyEl.value) {
    replyEl.value.scrollTop = replyEl.value.scrollHeight
  }
}

async function postChat(apiKey: string, body: object) {
  return fetch('/v1/chat/completions', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${apiKey}`,
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify(body),
  })
}

async function readNonStream(res: Response) {
  const data = await res.json()
  if (!res.ok) {
    throw new Error(data.error || data.message || `HTTP ${res.status}`)
  }
  reply.value = data.choices?.[0]?.message?.content || JSON.stringify(data, null, 2)
}

async function readStream(res: Response) {
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const data = await res.json()
      msg = data.error || data.message || msg
    } catch {
      /* ignore */
    }
    throw new Error(msg)
  }
  if (!res.body) throw new Error('浏览器不支持流式响应')

  const reader = res.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  reply.value = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const parts = buffer.split('\n')
    buffer = parts.pop() || ''

    for (const line of parts) {
      const trimmed = line.trim()
      if (!trimmed.startsWith('data:')) continue
      const payload = trimmed.slice(5).trim()
      if (!payload || payload === '[DONE]') continue
      try {
        const chunk = JSON.parse(payload)
        if (chunk.error?.message) {
          throw new Error(chunk.error.message)
        }
        const delta = chunk.choices?.[0]?.delta?.content
        if (delta) {
          reply.value += delta
          await scrollReply()
        }
      } catch (e: any) {
        if (e instanceof SyntaxError) continue
        throw e
      }
    }
  }
}

async function send() {
  if (!model.value.trim()) {
    notify.warning('请填写模型')
    return
  }
  if (!message.value.trim()) {
    notify.warning('请填写消息')
    return
  }

  loading.value = true
  reply.value = stream.value ? '' : '…'
  try {
    let apiKey = await ensureApiKey()
    const body = {
      model: model.value,
      stream: stream.value,
      messages: [{ role: 'user', content: message.value }],
    }

    let res = await postChat(apiKey, body)
    if (res.status === 401) {
      apiKey = await ensureApiKey(true)
      res = await postChat(apiKey, body)
    }

    if (stream.value) {
      await readStream(res)
    } else {
      await readNonStream(res)
    }
    await scrollReply()
  } catch (e: any) {
    notify.error(e.message || e.response?.data?.error || '请求失败')
    if (!reply.value || reply.value === '…') reply.value = ''
  } finally {
    loading.value = false
  }
}

onMounted(loadModels)
</script>

<template>
  <div class="chat-layout">
    <n-card title="请求" class="pane">
      <n-form label-placement="top">
        <n-form-item label="模型">
          <n-select
            v-if="modelOptions.length"
            v-model:value="model"
            filterable
            tag
            :options="modelOptions"
            placeholder="选择或输入模型"
          />
          <n-input v-else v-model:value="model" placeholder="模型 ID" />
        </n-form-item>
        <n-form-item label="流式输出 (stream)">
          <n-space align="center">
            <n-switch v-model:value="stream" />
            <span class="hint">{{ stream ? 'SSE 逐字返回' : '等待完整响应' }}</span>
          </n-space>
        </n-form-item>
        <n-form-item label="消息">
          <n-input
            v-model:value="message"
            type="textarea"
            :rows="12"
            placeholder="输入要发送的内容"
          />
        </n-form-item>
        <n-button type="primary" block :loading="loading" @click="send">发送</n-button>
      </n-form>
    </n-card>

    <n-card title="回复" class="pane">
      <div ref="replyEl" class="reply-box">
        <pre v-if="reply" class="reply">{{ reply }}</pre>
        <div v-else class="empty">发送后在此显示模型回复</div>
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.chat-layout {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) minmax(320px, 1fr);
  gap: 8px;
  align-items: stretch;
  min-height: calc(100vh - 80px);
}

.pane {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.pane :deep(.n-card__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.hint {
  font-size: 13px;
  color: #6b7280;
}

.reply-box {
  flex: 1;
  min-height: 360px;
  max-height: calc(100vh - 160px);
  overflow: auto;
  padding: 6px 8px;
  border-radius: 8px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
}

.reply {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.6;
  color: #111827;
  font-family: ui-sans-serif, system-ui, -apple-system, sans-serif;
}

.empty {
  color: #9ca3af;
  font-size: 13px;
  padding: 8px 0;
}

@media (max-width: 900px) {
  .chat-layout {
    grid-template-columns: 1fr;
    min-height: auto;
  }

  .reply-box {
    max-height: 420px;
  }
}
</style>
