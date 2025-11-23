import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { getToken } from '../utils/auth'

export default function Login() {
  const navigate = useNavigate()

  useEffect(() => {
    // If already authenticated, redirect to home
    if (getToken()) {
      navigate('/')
      return
    }
  }, [navigate])

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
      <div>
        <h1>Build Assistant</h1>
        <a href="/auth/login" style={{ display: 'inline-block', marginTop: '1rem', padding: '0.5rem 1rem', backgroundColor: '#007bff', color: 'white', textDecoration: 'none', borderRadius: '4px' }}>
          Login with OIDC
        </a>
      </div>
    </div>
  )
}

