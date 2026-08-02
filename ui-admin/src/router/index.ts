import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'login',
      meta: { title: '登录' },
      component: () => import('../views/Login.vue'),
    },
    {
      path: '/',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/dashboard' },
        {
          path: 'dashboard',
          meta: { title: '仪表板', subtitle: '核心运营指标一览' },
          component: () => import('../views/Dashboard.vue'),
        },
        {
          path: 'members',
          meta: { title: '会员管理', subtitle: '查看和管理注册用户' },
          component: () => import('../views/Members.vue'),
        },
        {
          path: 'orders',
          meta: { title: '订单管理', subtitle: '查看和管理用户订单' },
          component: () => import('../views/Orders.vue'),
        },
        {
          path: 'models',
          meta: { title: '模型管理', subtitle: 'LLM Relay 可用模型' },
          component: () => import('../views/Models.vue'),
        },
        {
          path: 'settings',
          meta: { title: '系统设置', subtitle: '部署环境与已启用模块' },
          component: () => import('../views/Settings.vue'),
        },
        {
          path: 'projects',
          meta: { title: '项目管理', subtitle: '多项目（租户）管理与切换' },
          component: () => import('../views/Projects.vue'),
        },
        {
          path: 'logs',
          meta: { title: '请求日志', subtitle: 'LLM Relay 调用记录' },
          component: () => import('../views/Logs.vue'),
        },
        {
          path: 'pricing',
          meta: { title: '付费方案', subtitle: '配置会员套餐与配额' },
          component: () => import('../views/Pricing.vue'),
        },
        {
          path: 'testchat',
          meta: { title: '聊天测试', subtitle: '快速验证 LLM Relay' },
          component: () => import('../views/TestChat.vue'),
        },
        {
          path: 'devices',
          meta: { title: '设备管理', subtitle: '查看用户绑定设备' },
          component: () => import('../views/Devices.vue'),
        },
        {
          path: 'analytics',
          meta: { title: '埋点分析', subtitle: '站点访问与事件统计' },
          component: () => import('../views/Analytics.vue'),
        },
        {
          path: 'coupons',
          meta: { title: '优惠券', subtitle: '创建和管理券码' },
          component: () => import('../views/Coupons.vue'),
        },
        {
          path: 'referral',
          meta: { title: '分享裂变', subtitle: '邀请关系与营销活动' },
          component: () => import('../views/Referral.vue'),
        },
        {
          path: 'surveys',
          meta: { title: '问卷调研', subtitle: '创建和管理问卷' },
          component: () => import('../views/Surveys.vue'),
        },
        {
          path: 'feedback',
          meta: { title: '用户反馈', subtitle: '处理客户反馈工单' },
          component: () => import('../views/Feedback.vue'),
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) return '/login'
  if (to.path === '/login' && token) return '/dashboard'
})

export default router
