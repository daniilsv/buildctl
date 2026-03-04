import apiClient from './client'

export interface SSHKey {
  id: string
  name: string
  fingerprint: string
  created_at: string
}

export interface CreateSSHKeyRequest {
  name: string
  private_key: string
}

export const getSSHKeys = async (): Promise<SSHKey[]> => {
  const { data } = await apiClient.get<SSHKey[]>('/ssh-keys')
  return data
}

export const createSSHKey = async (req: CreateSSHKeyRequest): Promise<SSHKey> => {
  const { data } = await apiClient.post<SSHKey>('/ssh-keys', req)
  return data
}

export const deleteSSHKey = async (id: string): Promise<void> => {
  await apiClient.delete(`/ssh-keys/${id}`)
}
