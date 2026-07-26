import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'

/** 兼容旧别名 llm / llm-api → llm-relay */
function normalizeModule(name: string) {
  const n = name.trim().toLowerCase()
  if (n === 'llm' || n === 'llm-api') return 'llm-relay'
  return n
}

export const useModulesStore = defineStore('modules', () => {
  const modules = ref<string[]>([])
  const loaded = ref(false)

  async function load() {
    try {
      const { data } = await http.get('/api/admin/modules')
      modules.value = (data.modules || []).map(normalizeModule)
    } catch {
      modules.value = []
    } finally {
      loaded.value = true
    }
  }

  function has(name: string) {
    return modules.value.includes(normalizeModule(name))
  }

  return { modules, loaded, load, has }
})
