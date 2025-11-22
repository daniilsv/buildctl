import apiClient from './client'

export interface Token {
  id: string
  name: string
  created_at: string
  last_used_at?: string
  expires_at?: string
}

export interface CreateTokenRequest {
  name: string
  expires_at?: number
}

export interface CreateTokenResponse {
  token: string
}

export const getTokens = async (): Promise<Token[]> => {
  const { data } = await apiClient.get<Token[]>('/tokens')
  return data
}

export const createToken = async (token: CreateTokenRequest): Promise<CreateTokenResponse> => {
  const { data } = await apiClient.post<CreateTokenResponse>('/tokens', token)
  return data
}

export const deleteToken = async (id: string): Promise<void> => {
  await apiClient.delete(`/tokens/${id}`)
}

