<script setup lang="ts">
import { NButton, NPagination, NIcon, NSpace } from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'

withDefaults(
  defineProps<{
    page?: number
    pageSize?: number
    itemCount?: number
    showPagination?: boolean
    loading?: boolean
  }>(),
  {
    page: 1,
    pageSize: 20,
    itemCount: 0,
    showPagination: false,
    loading: false,
  },
)

const emit = defineEmits<{
  'update:page': [number]
  refresh: []
}>()

function onPage(p: number) {
  emit('update:page', p)
  emit('refresh')
}
</script>

<template>
  <div class="list-toolbar">
    <div class="left">
      <n-space align="center" :wrap="true" :size="8">
        <slot />
      </n-space>
    </div>
    <div class="right">
      <n-pagination
        v-if="showPagination"
        :page="page"
        :page-size="pageSize"
        :item-count="itemCount"
        size="small"
        @update:page="onPage"
      />
      <n-button quaternary circle :loading="loading" title="刷新" @click="emit('refresh')">
        <template #icon>
          <n-icon :component="RefreshOutline" />
        </template>
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.list-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
  min-height: 30px;
}

.left {
  min-width: 0;
  flex: 1;
}

.right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
</style>
