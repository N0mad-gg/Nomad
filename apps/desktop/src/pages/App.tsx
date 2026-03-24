import { useEffect } from 'react'
import { useAuthStore } from '../store/auth'
import { useAppStore } from '../store/app'
import { getServers, getChannels, getMessages } from '../api/client'
import Login from './Login'
import ServerList from '../components/layout/ServerList'
import ChannelList from '../components/layout/ChannelList'
import ChatArea from '../components/chat/ChatArea'

export default function App() {
  const { token } = useAuthStore()
  const { activeServerId, activeChannelId, setServers, setChannels, setMessages } = useAppStore()

  useEffect(() => {
    if (!token) return
    getServers().then((r) => setServers(r.data))
  }, [token])

  useEffect(() => {
    if (!activeServerId) return
    getChannels(activeServerId).then((r) => setChannels(r.data))
  }, [activeServerId])

  useEffect(() => {
    if (!activeServerId || !activeChannelId) return
    getMessages(activeServerId, activeChannelId).then((r) => setMessages(r.data.reverse()))
  }, [activeChannelId])

  if (!token) return <Login />

  return (
    <div style={{ display: 'flex', height: '100vh', background: '#111', color: '#fff' }}>
      <ServerList />
      {activeServerId && <ChannelList />}
      {activeChannelId && <ChatArea />}
      {!activeChannelId && (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#555' }}>
          {activeServerId ? 'Select a channel' : 'Select a server'}
        </div>
      )}
    </div>
  )
}
