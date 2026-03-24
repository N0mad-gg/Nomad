import { useState } from 'react'
import { login, register } from '../api/client'
import { useAuthStore } from '../store/auth'

export default function Login() {
  const [isRegister, setIsRegister] = useState(false)
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const setAuth = useAuthStore((s) => s.setAuth)

  const submit = async () => {
    try {
      const res = isRegister
        ? await register(username, email, password)
        : await login(email, password)
      setAuth(res.data.token, res.data.user_id)
    } catch {
      setError('Failed. Check your credentials.')
    }
  }

  return (
    <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', background: '#111' }}>
      <div style={{ background: '#1e1e2e', padding: 32, borderRadius: 12, width: 340 }}>
        <h1 style={{ color: '#fff', marginBottom: 24 }}>Nomad</h1>
        {isRegister && (
          <input placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)}
            style={inputStyle} />
        )}
        <input placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)}
          style={inputStyle} />
        <input placeholder="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)}
          style={inputStyle} />
        {error && <p style={{ color: '#f38ba8', fontSize: 13 }}>{error}</p>}
        <button onClick={submit} style={btnStyle}>
          {isRegister ? 'Register' : 'Login'}
        </button>
        <p style={{ color: '#888', fontSize: 13, marginTop: 12, cursor: 'pointer' }}
          onClick={() => setIsRegister(!isRegister)}>
          {isRegister ? 'Already have an account?' : 'Create an account'}
        </p>
      </div>
    </div>
  )
}

const inputStyle: React.CSSProperties = {
  width: '100%', padding: '10px 12px', marginBottom: 12, borderRadius: 8,
  border: '1px solid #333', background: '#181825', color: '#fff', fontSize: 14, boxSizing: 'border-box',
}
const btnStyle: React.CSSProperties = {
  width: '100%', padding: '10px', borderRadius: 8, border: 'none',
  background: '#7c3aed', color: '#fff', fontSize: 14, cursor: 'pointer',
}
