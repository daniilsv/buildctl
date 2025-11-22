import { Outlet, Link, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useTheme } from './theme-provider'
import { Moon, Sun } from 'lucide-react'
import { Button } from './ui/button'
import apiClient from '../api/client'

export default function Layout() {
  const navigate = useNavigate()
  const { theme, setTheme } = useTheme()

  useEffect(() => {
    apiClient.get('/projects').catch(() => {
      navigate('/auth/login')
    })
  }, [navigate])

  return (
    <div className="min-h-screen flex flex-col">
      <nav className="border-b bg-card">
        <div className="container mx-auto px-4 py-3 flex items-center gap-4">
          <Link to="/" className="font-semibold text-lg">
            Build Assistant
          </Link>
          <div className="flex-1" />
          <Link to="/" className="text-sm hover:underline">
            Dashboard
          </Link>
          <Link to="/settings" className="text-sm hover:underline">
            Settings
          </Link>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          >
            <Sun className="h-5 w-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
            <Moon className="absolute h-5 w-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
            <span className="sr-only">Toggle theme</span>
          </Button>
          <a href="/auth/logout" className="text-sm hover:underline">
            Logout
          </a>
        </div>
      </nav>
      <main className="flex-1 container mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}

