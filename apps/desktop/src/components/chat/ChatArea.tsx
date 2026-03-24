import { useEffect, useRef, useState } from 'react'
import { useAppStore } from '../../store/app'
import { sendMessage } from '../../api/client'

export default function ChatArea() {
  const { messages, activeServerId, activeChannelId, channels, addMessage } = useAppStore()
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)
  const channel = channels.find((c) => c.id === activeChannelId)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const send = async () => {
    if (!input.trim() || !activeServerId || !activeChannelId) return
    const res = await sendMessage(activeServerId, activeChannelId, input.trim())
    addMessage(res.data)
    setInput('')
  }

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', background: '#111' }}>
      <div style={{ padding: '12px 16px', borderBottom: '1px solid #222', fontWeight: 600 }}>
        # {channel?.name}
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
        {messages.map((m) => (
          <div key={m.id} style={{ display: 'flex', gap: 10 }}>
            <div style={{ width: 36, height: 36, borderRadius: '50%', background: '#7c3aed', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, flexShrink: 0 }}>
              {m.username[0].toUpperCase()}
            </div>
            <div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                <span style={{ fontWeight: 600, fontSize: 14 }}>{m.username}</span>
                <span style={{ fontSize: 11, color: '#555' }}>{new Date(m.created_at).toLocaleTimeString()}</span>
              </div>
              <p style={{ margin: 0, fontSize: 14, color: '#ccc', lineHeight: 1.5 }}>{m.content}</p>
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <div style={{ padding: '12px 16px', borderTop: '1px solid #222' }}>
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && send()}
          placeholder={`Message #${channel?.name}`}
          style={{ width: '100%', padding: '12px 16px', borderRadius: 8, border: 'none', background: '#1e1e2e', color: '#fff', fontSize: 14, boxSizing: 'border-box', outline: 'none' }} />
      </div>
    </div>
  )
}
