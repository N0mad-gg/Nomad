import { create } from 'zustand'

interface Server { id: string; name: string; invite_code: string; icon_url?: string }
interface Channel { id: string; name: string; position: number }
interface Message { id: string; user_id: string; username: string; content: string; created_at: string }

interface AppState {
  servers: Server[]
  channels: Channel[]
  messages: Message[]
  activeServerId: string | null
  activeChannelId: string | null
  setServers: (servers: Server[]) => void
  setChannels: (channels: Channel[]) => void
  setMessages: (messages: Message[]) => void
  addMessage: (message: Message) => void
  setActiveServer: (id: string | null) => void
  setActiveChannel: (id: string | null) => void
}

export const useAppStore = create<AppState>((set) => ({
  servers: [],
  channels: [],
  messages: [],
  activeServerId: null,
  activeChannelId: null,
  setServers: (servers) => set({ servers }),
  setChannels: (channels) => set({ channels }),
  setMessages: (messages) => set({ messages }),
  addMessage: (message) => set((s) => ({ messages: [...s.messages, message] })),
  setActiveServer: (id) => set({ activeServerId: id, activeChannelId: null }),
  setActiveChannel: (id) => set({ activeChannelId: id }),
}))
