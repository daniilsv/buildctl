import apiClient from './client'

export interface Branch {
  id: string
  project_id: string
  name: string
  last_successful_commit?: string
  last_successful_at?: string
  settings: Record<string, any>
  created_at: string
}

export interface CreateBranchRequest {
  name: string
  settings?: Record<string, any>
}

export interface UpdateBranchRequest {
  settings: Record<string, any>
}

export const getBranches = async (projectName: string): Promise<Branch[]> => {
  const { data } = await apiClient.get<Branch[]>(`/projects/${projectName}/branches`)
  return data
}

export const getBranch = async (projectName: string, branchName: string): Promise<Branch> => {
  const { data } = await apiClient.get<Branch>(`/projects/${projectName}/branches/${branchName}`)
  return data
}

export const createBranch = async (projectName: string, branch: CreateBranchRequest): Promise<Branch> => {
  const { data } = await apiClient.post<Branch>(`/projects/${projectName}/branches`, branch)
  return data
}

export const updateBranch = async (projectName: string, branchName: string, branch: UpdateBranchRequest): Promise<Branch> => {
  const { data } = await apiClient.patch<Branch>(`/projects/${projectName}/branches/${branchName}`, branch)
  return data
}

export const getBuildsByBranch = async (projectName: string, branchName: string) => {
  const { data } = await apiClient.get(`/projects/${projectName}/branches/${branchName}/builds`)
  return data
}

export const testBranchTelegramNotification = async (projectName: string, branchName: string, chatID: string, threadID?: string): Promise<{ status: string; message: string }> => {
  const { data } = await apiClient.post<{ status: string; message: string }>(`/projects/${projectName}/branches/${branchName}/test-telegram`, {
    chat_id: chatID,
    thread_id: threadID || undefined,
  })
  return data
}

export const testBranchWebhook = async (projectName: string, branchName: string, webhookURL: string): Promise<{ status: string; message: string }> => {
  const { data } = await apiClient.post<{ status: string; message: string }>(`/projects/${projectName}/branches/${branchName}/test-webhook`, {
    webhook_url: webhookURL,
  })
  return data
}

export const testBranchB24Notification = async (projectName: string, branchName: string, typeKey: string): Promise<{ status: string; message: string }> => {
  const { data } = await apiClient.post<{ status: string; message: string }>(`/projects/${projectName}/branches/${branchName}/test-b24`, {
    type_key: typeKey,
  })
  return data
}

export const redeployBranch = async (projectName: string, branchName: string): Promise<{ status: string; message: string }> => {
  const { data } = await apiClient.post<{ status: string; message: string }>(`/projects/${projectName}/branches/${branchName}/redeploy`)
  return data
}

