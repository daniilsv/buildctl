import apiClient from './client'

export interface Project {
  id: string
  name: string
  title: string
  repository_url: string
  repository_type: string
  settings: Record<string, any>
  created_at: string
}

export const getProjects = async (): Promise<Project[]> => {
  const { data } = await apiClient.get<Project[]>('/projects')
  return data
}

export const getProject = async (idOrName: string): Promise<Project> => {
  const { data } = await apiClient.get<Project>(`/projects/${idOrName}`)
  return data
}

export const getProjectByName = async (name: string): Promise<Project> => {
  const { data } = await apiClient.get<Project>(`/projects/${name}`)
  return data
}

export const createProject = async (project: Partial<Project>): Promise<Project> => {
  const { data } = await apiClient.post<Project>('/projects', project)
  return data
}

export const updateProject = async (name: string, project: Partial<Project>): Promise<Project> => {
  const { data } = await apiClient.patch<Project>(`/projects/${name}`, project)
  return data
}

export const deleteProject = async (name: string): Promise<void> => {
  await apiClient.delete(`/projects/${name}`)
}

export const testProjectNotifications = async (name: string): Promise<{ status: string; message: string }> => {
  const { data } = await apiClient.post<{ status: string; message: string }>(`/projects/${name}/test-notifications`)
  return data
}

export const testProjectWebhooks = async (name: string): Promise<{ status: string; message: string }> => {
  const { data } = await apiClient.post<{ status: string; message: string }>(`/projects/${name}/test-webhooks`)
  return data
}

