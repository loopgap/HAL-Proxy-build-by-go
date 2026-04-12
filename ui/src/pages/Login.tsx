import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login } = useAuthStore()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/'

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      const response = await fetch('/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Login failed')
      }

      const data = await response.json()
      login({ id: data.user_id, name: username, email: `${username}@example.com` }, data.token)
      navigate(from, { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className='min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4'>
      <Card className='w-full max-w-md'>
        <CardHeader title='BridgeOS' subtitle='Enter your credentials to access the system' />
        <CardContent>
          <form onSubmit={handleSubmit} className='space-y-4'>
            {error && (
              <div className='bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm p-3 rounded-md'>
                {error}
              </div>
            )}
            <div className='space-y-2'>
              <label htmlFor='username' className='text-sm font-medium text-gray-700 dark:text-gray-300'>
                Username
              </label>
              <input
                id='username'
                type='text'
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className='w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:text-white'
                placeholder='Enter username'
                required
              />
            </div>
            <div className='space-y-2'>
              <label htmlFor='password' className='text-sm font-medium text-gray-700 dark:text-gray-300'>
                Password
              </label>
              <input
                id='password'
                type='password'
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className='w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:text-white'
                placeholder='Enter password'
                required
              />
            </div>
            <Button type='submit' className='w-full' disabled={isLoading}>
              {isLoading ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>
          <div className='mt-4 text-center text-sm text-gray-500 dark:text-gray-400'>
            <p>Development mode: Use any credentials</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}