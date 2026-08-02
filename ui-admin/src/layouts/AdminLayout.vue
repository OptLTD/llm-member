<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NButton,
  NIcon,
  useDialog,
  type MenuOption,
} from 'naive-ui'
import {
  SpeedometerOutline,
  ChatbubblesOutline,
  ListOutline,
  PeopleOutline,
  CartOutline,
  HardwareChipOutline,
  PricetagOutline,
  SettingsOutline,
  DesktopOutline,
  TicketOutline,
  ShareSocialOutline,
  DocumentTextOutline,
  AnalyticsOutline,
  ChatboxEllipsesOutline,
  LogOutOutline,
} from '@vicons/ionicons5'
import { useAuthStore } from '../stores/auth'
import { useModulesStore } from '../stores/modules'
import { useProjectsStore } from '../stores/projects'
import { adminMenuGroups, adminPinnedItems, type AdminMenuItem } from '../config/menu'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const mods = useModulesStore()
const projects = useProjectsStore()
const dialog = useDialog()

const iconMap: Record<string, any> = {
  speedometer: SpeedometerOutline,
  chatbubbles: ChatbubblesOutline,
  list: ListOutline,
  people: PeopleOutline,
  cart: CartOutline,
  hardware: HardwareChipOutline,
  pricetag: PricetagOutline,
  settings: SettingsOutline,
  desktop: DesktopOutline,
  ticket: TicketOutline,
  share: ShareSocialOutline,
  document: DocumentTextOutline,
  analytics: AnalyticsOutline,
  chatbox: ChatboxEllipsesOutline,
}

onMounted(async () => {
  await mods.load()
  await projects.load()
})

async function onProjectChange(id: number) {
  projects.select(id)
  await mods.load()
}

const projectMenuOptions = computed(() => [
  ...projects.projects.map((p) => ({
    label: p.name,
    key: `project:${p.id}`,
  })),
  { type: 'divider', key: 'divider' },
  { label: '项目管理', key: 'manage' },
])

function onProjectMenuSelect(key: string) {
  if (key === 'manage') {
    router.push('/projects')
    return
  }
  if (key.startsWith('project:')) {
    const id = Number(key.slice('project:'.length))
    if (id > 0) onProjectChange(id)
  }
}


function renderIcon(name: string) {
  const icon = iconMap[name]
  return icon ? () => h(NIcon, null, { default: () => h(icon) }) : undefined
}

function isVisible(item: AdminMenuItem) {
  if (!item.module) return true
  if (!mods.loaded) return true
  return mods.has(item.module)
}

const menuOptions = computed<MenuOption[]>(() =>
  adminMenuGroups
    .map((group) => {
      const children = group.items.filter(isVisible).map((item) => ({
        label: item.label,
        key: item.path,
        icon: renderIcon(item.icon),
      }))
      if (!children.length) return null
      return {
        label: group.title,
        key: `group:${group.key}`,
        children,
      } as MenuOption
    })
    .filter((x): x is MenuOption => x != null),
)

const expandedKeys = ref<string[]>([])
const expandInited = ref(false)

function groupKeyForPath(path: string) {
  for (const opt of menuOptions.value) {
    if (opt.children?.some((c) => c.key === path)) return String(opt.key)
  }
  return null
}

watch(
  menuOptions,
  (opts) => {
    if (!opts.length) return
    const valid = new Set(opts.map((o) => String(o.key)))
    const next = new Set(expandedKeys.value.filter((k) => valid.has(k)))
    if (!expandInited.value) {
      next.add(String(opts[0].key))
      expandInited.value = true
    }
    const activeGroup = groupKeyForPath(route.path)
    if (activeGroup) next.add(activeGroup)
    expandedKeys.value = [...next]
  },
  { immediate: true },
)

watch(
  () => route.path,
  (path) => {
    const g = groupKeyForPath(path)
    if (g && !expandedKeys.value.includes(g)) {
      expandedKeys.value = [...expandedKeys.value, g]
    }
  },
)

const activeKey = computed(() => route.path)

const pageTitle = computed(() => (route.meta.title as string) || '')
const pageSubtitle = computed(() => (route.meta.subtitle as string) || '')

function onMenuSelect(key: string) {
  if (key.startsWith('group:')) return
  router.push(key)
}

function logout() {
  dialog.warning({
    title: '确认退出',
    content: '确定要退出当前账号吗？',
    positiveText: '退出',
    negativeText: '取消',
    onPositiveClick: () => {
      auth.logout()
      router.push('/login')
    },
  })
}
</script>

<template>
  <n-layout has-sider class="admin-layout">
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="0"
      :width="240"
      :native-scrollbar="true"
      class="sider"
    >
      <div class="sider-shell">
        <div class="sider-body">
          <div class="brand">
            <div class="brand-left">
              <img src="/favicon.svg" alt="" class="brand-icon" />
              <span class="brand-name">Member</span>
            </div>
            <n-dropdown
              v-if="projects.loaded"
              trigger="click"
              :options="projectMenuOptions"
              :value="`project:${projects.currentId}`"
              @select="onProjectMenuSelect"
            >
              <span class="brand-project">
                {{ projects.current()?.name || '选择项目' }}
                <n-icon size="14"><ChevronDownOutline /></n-icon>
              </span>
            </n-dropdown>
          </div>

          <n-menu
            inverted
            :value="activeKey"
            :options="menuOptions"
            v-model:expanded-keys="expandedKeys"
            :indent="22"
            :root-indent="12"
            @update:value="onMenuSelect"
          />
        </div>

        <div class="sider-footer">
          <div class="pinned-row">
            <button
              v-for="item in adminPinnedItems"
              :key="item.path"
              type="button"
              class="pinned-item"
              :class="{ active: activeKey === item.path }"
              @click="router.push(item.path)"
            >
              <n-icon :component="iconMap[item.icon]" :size="18" />
              <span>{{ item.label }}</span>
            </button>
          </div>
          <div class="footer-user">
            <span class="username">{{ auth.username || '管理员' }}</span>
            <n-button quaternary size="small" @click="logout">
              <template #icon>
                <n-icon><LogOutOutline /></n-icon>
              </template>
              退出
            </n-button>
          </div>
        </div>
      </div>
    </n-layout-sider>

    <n-layout>
      <n-layout-header bordered class="header">
        <div class="header-left">
          <h1 class="header-title">{{ pageTitle }}</h1>
          <span v-if="pageSubtitle" class="header-subtitle">
            {{ pageSubtitle }}
          </span>
        </div>
      </n-layout-header>
      <n-layout-content class="content">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<style scoped>
.admin-layout {
  height: 100vh;
  min-height: 100vh;
}

.sider {
  background: #111827 !important;
  height: 100vh;
}

/* 禁止外层滚动，只让 sider-body 滚，底部才能钉住 */
.sider :deep(.n-layout-sider-scroll-container) {
  height: 100% !important;
  max-height: 100vh !important;
  overflow: hidden !important;
  display: flex !important;
  flex-direction: column !important;
}

.sider-shell {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}

.sider-body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  padding-bottom: 8px;
}

.sider :deep(.n-menu) {
  background: transparent;
  margin-top: 4px;
}

.sider :deep(.n-menu-item-content) {
  border-radius: 8px;
  margin: 2px 10px;
}

.sider :deep(.n-menu-item-content--selected) {
  background: rgba(99, 102, 241, 0.18) !important;
}

.brand {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 14px 12px;
  color: #f9fafb;
  font-weight: 600;
  font-size: 16px;
}

.brand-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.brand-icon {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}

.brand-name {
  white-space: nowrap;
  font-size: 16px;
  font-weight: 600;
  color: #f9fafb;
}

.brand-project {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 500;
  color: #e5e7eb;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 3px 8px;
  cursor: pointer;
  transition: background 0.15s ease;
  flex-shrink: 1;
}

.brand-project:hover {
  background: rgba(255, 255, 255, 0.14);
}

.pinned-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 0;
}

.pinned-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 8px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.04);
  color: #e5e7eb;
  font-size: 12px;
  line-height: 1.2;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
}

.pinned-item :deep(.n-icon) {
  display: inline-flex;
  flex-shrink: 0;
}

.pinned-item span {
  white-space: nowrap;
}

.pinned-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #f9fafb;
}

.pinned-item.active {
  background: rgba(99, 102, 241, 0.22);
  border-color: rgba(99, 102, 241, 0.45);
  color: #fff;
}

.sider-footer {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  background: #0f1623;
  z-index: 2;
}

.footer-user {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 4px;
}

.sider-footer :deep(.n-button) {
  color: #e5e7eb !important;
}

.sider-footer :deep(.n-button .n-icon) {
  color: #e5e7eb !important;
}

.sider-footer :deep(.n-button:hover) {
  color: #fff !important;
  background: rgba(255, 255, 255, 0.08) !important;
}

.username {
  font-size: 13px;
  color: #e5e7eb;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header {
  height: 48px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  background: #fff;
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}



.header-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #111827;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.header-subtitle {
  font-size: 12px;
  color: #9ca3af;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.content {
  padding: 16px;
  background: #f3f4f6;
  min-height: calc(100vh - 48px);
}

/* 正文卡片圆角；纵向堆叠的卡片间距（不影响网格内横排卡片） */
.content :deep(.n-card) {
  border-radius: 8px;
}

.content :deep(.n-grid + .n-card) {
  margin-top: 8px;
}

.content :deep(.plans-grid > .n-card),
.content :deep(.n-grid .n-card) {
  margin-top: 0;
}
</style>

<style>
.sider .n-layout-sider-border {
  border-color: rgba(255, 255, 255, 0.06) !important;
}

/* 分组标题：浅色字，适配深色侧栏 */
.sider .n-menu.n-menu--inverted > .n-submenu > .n-menu-item-content,
.sider .n-menu > .n-submenu > .n-menu-item-content {
  margin: 10px 10px 2px !important;
  padding-left: 10px !important;
  min-height: 32px !important;
  background: transparent !important;
}

.sider .n-menu > .n-submenu > .n-menu-item-content .n-menu-item-content-header,
.sider .n-menu > .n-submenu > .n-menu-item-content .n-menu-item-content-header a {
  color: #f3f4f6 !important;
  opacity: 1 !important;
  font-size: 13px !important;
  font-weight: 600 !important;
  letter-spacing: 0.02em;
}

.sider .n-menu > .n-submenu > .n-menu-item-content .n-menu-item-content__arrow {
  color: #d1d5db !important;
  font-size: 12px !important;
  opacity: 1 !important;
}

.sider .n-menu > .n-submenu > .n-menu-item-content:hover {
  background: rgba(255, 255, 255, 0.06) !important;
}

.sider .n-menu > .n-submenu > .n-menu-item-content:hover .n-menu-item-content-header,
.sider .n-menu > .n-submenu > .n-menu-item-content:hover .n-menu-item-content__arrow {
  color: #fff !important;
}

/* 子菜单项 */
.sider .n-menu .n-submenu-children .n-menu-item-content {
  margin: 1px 10px 1px 14px !important;
  min-height: 38px !important;
}

.sider .n-menu .n-submenu-children .n-menu-item-content-header {
  color: #e5e7eb !important;
  font-size: 13.5px !important;
  font-weight: 450 !important;
  letter-spacing: 0;
  opacity: 1 !important;
}

.sider .n-menu .n-submenu-children .n-menu-item-content__icon {
  color: #c4b5fd !important;
}

.sider .n-menu .n-submenu-children .n-menu-item-content:hover {
  background: rgba(255, 255, 255, 0.06) !important;
}

.sider .n-menu .n-submenu-children .n-menu-item-content--selected {
  background: rgba(99, 102, 241, 0.2) !important;
}

.sider .n-menu .n-submenu-children .n-menu-item-content--selected .n-menu-item-content-header,
.sider .n-menu .n-submenu-children .n-menu-item-content--selected .n-menu-item-content__icon {
  color: #fff !important;
}
</style>
