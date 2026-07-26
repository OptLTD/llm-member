<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NInput, NButton, NForm, NFormItem } from 'naive-ui'
import { useAuthStore } from '../stores/auth'
import { notify } from '../utils/notify'

const username = ref('')
const password = ref('')
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()

async function submit() {
  if (!username.value.trim() || !password.value) {
    notify.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    notify.success('登录成功')
    router.push('/dashboard')
  } catch (e: any) {
    notify.error(e.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <aside class="login-brand">
      <div class="brand-content">
        <div class="brand-logo">
          <img src="/favicon.svg" alt="Logo" />
        </div>
        <h1>会员管理后台</h1>
        <p class="brand-desc">统一管理会员、订单、设备与运营数据，助力业务高效运转。</p>
        <ul class="brand-features">
          <li>会员与订单一站式管理</li>
          <li>多模块按需启用，灵活扩展</li>
          <li>数据可视化与运营分析</li>
        </ul>
      </div>
      <div class="brand-footer">Option Member Admin</div>
    </aside>

    <main class="login-main">
      <div class="login-form-wrap">
        <div class="form-header">
          <h2>欢迎回来</h2>
          <p>输入账号密码，进入管理控制台</p>
        </div>

        <n-form class="login-form" @submit.prevent="submit">
          <n-form-item label="用户名">
            <n-input
              v-model:value="username"
              placeholder="admin"
              size="large"
              autocomplete="username"
            />
          </n-form-item>
          <n-form-item label="密码">
            <n-input
              v-model:value="password"
              type="password"
              show-password-on="click"
              placeholder="请输入密码"
              size="large"
              autocomplete="current-password"
              @keyup.enter="submit"
            />
          </n-form-item>
          <n-button type="primary" attr-type="submit" size="large" block :loading="loading" class="submit-btn">
            登录
          </n-button>
        </n-form>

        <p class="form-footer">仅限授权管理员访问</p>
      </div>
    </main>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  min-height: 100vh;
}

.login-brand {
  position: relative;
  flex: 0 0 44%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px 56px;
  color: #fff;
  background:
    radial-gradient(ellipse 80% 60% at 20% 10%, rgba(99, 102, 241, 0.35), transparent),
    radial-gradient(ellipse 60% 50% at 80% 90%, rgba(139, 92, 246, 0.25), transparent),
    linear-gradient(145deg, #1e1b4b 0%, #312e81 45%, #3730a3 100%);
  overflow: hidden;
}

.login-brand::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: linear-gradient(to bottom, rgba(0, 0, 0, 0.6), transparent);
  pointer-events: none;
}

.brand-content {
  position: relative;
  z-index: 1;
  max-width: 420px;
}

.brand-logo {
  width: 52px;
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 28px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.brand-logo img {
  width: 32px;
  height: 32px;
}

.login-brand h1 {
  font-size: 32px;
  font-weight: 700;
  line-height: 1.25;
  letter-spacing: -0.02em;
  color: #fff;
  margin-bottom: 16px;
}

.brand-desc {
  margin: 0 0 32px;
  font-size: 15px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.72);
}

.brand-features {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.brand-features li {
  position: relative;
  padding-left: 22px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
}

.brand-features li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 7px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #a5b4fc;
}

.brand-footer {
  position: relative;
  z-index: 1;
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.35);
}

.login-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 64px;
  background: #fff;
}

.login-form-wrap {
  width: 100%;
  max-width: 360px;
}

.form-header {
  margin-bottom: 40px;
}

.form-header h2 {
  font-size: 26px;
  font-weight: 600;
  letter-spacing: -0.02em;
  margin-bottom: 8px;
  color: #111827;
}

.form-header p {
  margin: 0;
  font-size: 14px;
  color: #6b7280;
}

.submit-btn {
  margin-top: 8px;
}

.form-footer {
  margin: 32px 0 0;
  font-size: 12px;
  color: #9ca3af;
  text-align: center;
}

@media (max-width: 900px) {
  .login-page {
    flex-direction: column;
  }

  .login-brand {
    flex: none;
    padding: 36px 28px 32px;
  }

  .login-brand h1 {
    font-size: 26px;
  }

  .brand-features,
  .brand-footer {
    display: none;
  }

  .login-main {
    padding: 32px 24px 48px;
  }
}
</style>
