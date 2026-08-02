import axios from 'axios'

const http = axios.create({ baseURL: '' })

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  const projectId = localStorage.getItem('projectId')
  if (projectId && Number(projectId) > 0) {
    config.headers['X-Project-Id'] = projectId
  }
  return config
})

http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const url = String(err.config?.url || '')
      // /v1 用的是用户 API Key，失败时不要清掉管理员登录态
      if (!url.includes('/v1/')) {
        localStorage.removeItem('token')
        if (!location.pathname.includes('/login')) {
          location.href = '/admin/login'
        }
      }
    }
    return Promise.reject(err)
  }
)

export default http
