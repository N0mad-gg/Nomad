import { useState } from 'react'
import { useAppStore } from '../../store/app'
import { createChannel, getChannels } from '../../api/client'
import { useAuthStore } from '../../store/auth'

export default function ChannelList() {
  const { channels, activeServerId, activeChannelId, setActiveChannel, setChannels } = useAppStore()
  const { logout } = useAuthStore()
  const [input, setInput] = useState('')
  const [adding, setAdding] = useState(false)

  const serverName = useAppStore((s) => s.servers.find((sv) => sv.id === s.activeServerId)?.name)

  const addChannel = async () => {
    if (!activeServerId || !input.trim()) return
    await createChannel(activeServerId, input.trim())
    const res = await getChannels(activeServerId)
    setChannels(res.data)
    setInput('')
    setAdding(false)
  }

  return (
    <div style={{ width: 240, background: '#161622', display: 'flex', flexDirection: 'column' }}>
      <div style={{ padding: '16px 12px', borderBottom: '1px solid #222', fontWeight: 700 }}>
        {serverName ?? 'Server'}
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '8px 0' }}>
        <div style={{ padding: '4px 12px', fontSize: 11, color: '#555', textTransform: 'uppercase', letterSpacing: 1, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          Channels
          <span onClick={() => setAdding(true)} style={{ cursor: 'pointer', fontSize: 16, color: '#7c3aed' }}>+</span>
        </div>
        {channels.map((ch) => (
          <div key={ch.id} onClick={() => setActiveChannel(ch.id)}
            style={{
              padding: '6px 16px', cursor: 'pointer', borderRadius: 6, margin: '2px 8px',
              background: activeChannelId === ch.id ? '#2a2a3e' : 'transparent',
              color: activeChannelId === ch.id ? '#fff' : '#888',
            }}>
            # {ch.name}
          </div>
        ))}
        {adding && (
          <div style={{ padding: '4px 8px', display: 'flex', gap: 4 }}>
            <input autoFocus value={input} onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && addChannel()}
              placeholder="channel-name"
              style={{ flex: 1, padding: '6px 8px', borderRadius: 6, border: '1px solid #333', background: '#181825', color: '#fff', fontSize: 13 }} />
          </div>
        )}
      </div>
      <div style={{ padding: 12, borderTop: '1px solid #222' }}>
        <button onClick={logout} style={{ width: '100%', padding: 8, borderRadius: 8, border: 'none', background: '#333', color: '#aaa', cursor: 'pointer', fontSize: 13 }}>
          Logout
        </button>
      </div>
    </div>
  )
}
