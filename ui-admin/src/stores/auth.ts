import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref('')

  async function login(user: string, password: string) {
    const { data } = await http.post('/api/admin/signin', { username: user, password })
    token.value = data.token
    username.value = data.user?.username || user
    localStorage.setItem('token', data.token)
    return data
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('apiKey')
  }

  async function fetchCurrent() {
    const { data } = await http.get('/api/admin/current')
    username.value = data.user?.username || ''
    return data
  }

  return { token, username, login, logout, fetchCurrent }
})
