import apiClient from './client'

export interface Build {
  id: string
  project_id: string
  project_name?: string
  branch_id: string
  branch_name?: string
  commit_hash: string
  commit_message?: string
  status: string
  started_at: string
  finished_at?: string
  created_at: string
  logs?: Log[]
}

export interface Log {
  id: string
  status: string
  log_message: string
  artifact_s3_key?: string
  artifact_url?: string
  created_at: string
}

export interface Artifact {
  id: string
  build_id: string
  log_id?: string
  project_id: string
  branch_id: string
  commit_hash: string
  filename: string
  s3_key: string
  size_bytes: number
  content_type?: string
  artifact_type: 'file' | 'container_image'
  image_name?: string
  image_tag?: string
  image_digest?: string
  created_at: string
  public_url: string
}

// Helper to check if string is UUID
const isUUID = (str: string): boolean => {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
  return uuidRegex.test(str)
}

export const getBuilds = async (projectNameOrId?: string, branchNameOrId?: string): Promise<Build[]> => {
  const params = new URLSearchParams()
  if (projectNameOrId) {
    if (isUUID(projectNameOrId)) {
      params.append('project_id', projectNameOrId)
    } else {
      params.append('project_name', projectNameOrId)
    }
  }
  if (branchNameOrId) {
    if (isUUID(branchNameOrId)) {
      params.append('branch_id', branchNameOrId)
    } else {
      params.append('branch_name', branchNameOrId)
    }
  }
  const queryString = params.toString()
  const url = queryString ? `/builds?${queryString}` : '/builds'
  const { data } = await apiClient.get<Build[]>(url)
  return data
}

export const getBuild = async (id: string): Promise<Build> => {
  const { data } = await apiClient.get<Build>(`/builds/${id}`)
  return data
}

export const getBuildArtifacts = async (buildId: string): Promise<Artifact[]> => {
  const { data } = await apiClient.get<Artifact[]>(`/builds/${buildId}/artifacts`)
  return data
}

export const deleteArtifact = async (artifactId: string): Promise<void> => {
  await apiClient.delete(`/artifacts/${artifactId}`)
}

export const deleteBuildArtifacts = async (buildId: string): Promise<void> => {
  await apiClient.delete(`/builds/${buildId}/artifacts`)
}
