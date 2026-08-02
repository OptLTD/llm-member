import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'

export interface ProjectItem {
  id: number
  name: string
  slug: string
  status: string
  appName?: string
}

export const useProjectsStore = defineStore('projects', () => {
  const projects = ref<ProjectItem[]>([])
  const currentId = ref<number>(Number(localStorage.getItem('projectId') || 0) || 0)
  const loaded = ref(false)

  const current = (): ProjectItem | undefined =>
    projects.value.find((p) => p.id === currentId.value)

  async function load() {
    try {
      const { data } = await http.get('/api/admin/projects')
      projects.value = (data.projects || []).map((p: any) => ({
        id: p.id,
        name: p.name,
        slug: p.slug,
        status: p.status,
        appName: p.settings?.appName,
      }))
      if (!projects.value.some((p) => p.id === currentId.value)) {
        currentId.value = projects.value[0]?.id || 0
      }
    } catch {
      projects.value = []
    } finally {
      loaded.value = true
    }
  }

  function select(id: number) {
    currentId.value = id
    localStorage.setItem('projectId', String(id))
  }

  return { projects, currentId, current, loaded, load, select }
})
