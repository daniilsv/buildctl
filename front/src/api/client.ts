import axios from 'axios'
import { getToken, removeToken } from '../utils/auth'

const apiClient = axios.create({
  baseURL: '/api/v1',
})

apiClient.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 || error.response?.status === 403) {
      removeToken()
      window.location.href = '/auth/login'
    }
    return Promise.reject(error)
  }
)

export default apiClient

