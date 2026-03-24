import { useState } from 'react'
import { useAppStore } from '../../store/app'
import { createServer, joinServer, getServers } from '../../api/client'

export default function ServerList() {
  const { servers, activeServerId, setActiveServer, setServers } = useAppStore()
  const [showModal, setShowModal] = useState(false)
  const [mode, setMode] = useState<'create' | 'join'>('create')
  const [input, setInput] = useState('')

  const submit = async () => {
    if (mode === 'create') await createServer(input)
    else await joinServer(input)
    const res = await getServers()
    setServers(res.data)
    setInput('')
    setShowModal(false)
  }

  return (
    <div style={{ width: 72, background: '#0d0d0d', display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '12px 0', gap: 8 }}>
      {servers.map((s) => (
        <div key={s.id} onClick={() => setActiveServer(s.id)}
          title={s.name}
          style={{
            width: 48, height: 48, borderRadius: activeServerId === s.id ? 16 : '50%',
            background: activeServerId === s.id ? '#7c3aed' : '#1e1e2e',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer', fontWeight: 700, fontSize: 18, transition: 'all 0.15s',
          }}>
          {s.name[0].toUpperCase()}
        </div>
      ))}
      <div onClick={() => setShowModal(true)}
        style={{ width: 48, height: 48, borderRadius: '50%', background: '#1e1e2e', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', fontSize: 24, color: '#7c3aed' }}>
        +
      </div>

      {showModal && (
        <div style={{ position: 'fixed', inset: 0, background: '#0008', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: '#1e1e2e', padding: 24, borderRadius: 12, width: 300 }}>
            <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
              {(['create', 'join'] as const).map((m) => (
                <button key={m} onClick={() => setMode(m)}
                  style={{ flex: 1, padding: 8, borderRadius: 8, border: 'none', background: mode === m ? '#7c3aed' : '#333', color: '#fff', cursor: 'pointer' }}>
                  {m === 'create' ? 'Create' : 'Join'}
                </button>
              ))}
            </div>
            <input
              placeholder={mode === 'create' ? 'Server name' : 'Invite code'}
              value={input} onChange={(e) => setInput(e.target.value)}
              style={{ width: '100%', padding: '10px 12px', borderRadius: 8, border: '1px solid #333', background: '#181825', color: '#fff', fontSize: 14, boxSizing: 'border-box', marginBottom: 12 }} />
            <div style={{ display: 'flex', gap: 8 }}>
              <button onClick={() => setShowModal(false)} style={{ flex: 1, padding: 8, borderRadius: 8, border: 'none', background: '#333', color: '#fff', cursor: 'pointer' }}>Cancel</button>
              <button onClick={submit} style={{ flex: 1, padding: 8, borderRadius: 8, border: 'none', background: '#7c3aed', color: '#fff', cursor: 'pointer' }}>OK</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
