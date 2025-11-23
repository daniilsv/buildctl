import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { setToken } from '../utils/auth'

export default function Callback() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()

  useEffect(() => {
    const code = searchParams.get('code')
    const state = searchParams.get('state')

    if (!code || !state) {
      navigate('/auth/login')
      return
    }

    // Exchange code for token via backend callback
    // Use fetch directly to avoid interceptor issues
    fetch(`/auth/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`, {
      credentials: 'include',
    })
      .then((response) => response.json())
      .then((data) => {
        const { access_token } = data
        if (access_token) {
          setToken(access_token)
          navigate('/')
        } else {
          navigate('/auth/login')
        }
      })
      .catch(() => {
        navigate('/auth/login')
      })
  }, [searchParams, navigate])

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
      <div>Authenticating...</div>
    </div>
  )
}

