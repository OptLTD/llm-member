export interface AdminMenuItem {
  path: string
  label: string
  icon: string
  /** 依赖的后端模块；省略表示始终显示 */
  module?: string
}

export interface AdminMenuGroup {
  key: string
  title: string
  icon: string
  items: AdminMenuItem[]
}

/** 侧栏顶部并排固定项（非分组） */
export const adminPinnedItems: AdminMenuItem[] = [
  { path: '/dashboard', label: '仪表盘', icon: 'speedometer' },
  { path: '/settings', label: '系统设置', icon: 'settings' },
]

/** 按工作流分组；整组无可见子项时侧栏隐藏 */
export const adminMenuGroups: AdminMenuGroup[] = [
  {
    key: 'subscription',
    title: '会员与订阅',
    icon: 'people',
    items: [
      { path: '/members', label: '会员管理', icon: 'people', module: 'member' },
      { path: '/orders', label: '订单管理', icon: 'cart', module: 'member' },
      { path: '/pricing', label: '付费方案', icon: 'pricetag' },
      { path: '/devices', label: '设备管理', icon: 'desktop', module: 'device' },
    ],
  },
  {
    key: 'growth',
    title: '增长运营',
    icon: 'ticket',
    items: [
      { path: '/coupons', label: '优惠券', icon: 'ticket', module: 'coupon' },
      { path: '/referral', label: '分享裂变', icon: 'share', module: 'referral' },
      { path: '/surveys', label: '问卷调研', icon: 'document', module: 'survey' },
    ],
  },
  {
    key: 'insight',
    title: '用户洞察',
    icon: 'analytics',
    items: [
      { path: '/analytics', label: '埋点分析', icon: 'analytics', module: 'analytics' },
      { path: '/feedback', label: '用户反馈', icon: 'chatbox', module: 'feedback' },
    ],
  },
  {
    key: 'ai',
    title: 'AI 能力',
    icon: 'hardware',
    items: [
      { path: '/models', label: '模型管理', icon: 'hardware', module: 'llm-relay' },
      { path: '/logs', label: '请求日志', icon: 'list', module: 'llm-relay' },
      { path: '/testchat', label: '聊天测试', icon: 'chatbubbles', module: 'llm-relay' },
    ],
  },
]
