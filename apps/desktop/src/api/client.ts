import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export const api = axios.create({ baseURL: `${BASE_URL}/api/v1` })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Auth
export const register = (username: string, email: string, password: string) =>
  api.post('/auth/register', { username, email, password })

export const login = (email: string, password: string) =>
  api.post('/auth/login', { email, password })

// Servers
export const getServers = () => api.get('/servers')
export const createServer = (name: string) => api.post('/servers', { name })
export const joinServer = (inviteCode: string) => api.post(`/servers/join/${inviteCode}`)
export const deleteServer = (id: string) => api.delete(`/servers/${id}`)

// Channels
export const getChannels = (serverId: string) => api.get(`/servers/${serverId}/channels`)
export const createChannel = (serverId: string, name: string) =>
  api.post(`/servers/${serverId}/channels`, { name })

// Messages
export const getMessages = (serverId: string, channelId: string) =>
  api.get(`/servers/${serverId}/channels/${channelId}/messages`)
export const sendMessage = (serverId: string, channelId: string, content: string) =>
  api.post(`/servers/${serverId}/channels/${channelId}/messages`, { content })
