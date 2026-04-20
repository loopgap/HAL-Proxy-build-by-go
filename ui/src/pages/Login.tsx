import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login, logout } = useAuthStore()
  const [token, setToken] = useState('')
  const [error, setError] = useState('')

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/'

  const handleUseToken = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const trimmed = token.trim()
    if (!trimmed) {
      setError('Enter a Bearer token or continue in local trusted mode.')
      return
    }

    login(
      { id: 'api-token', name: 'API Token', email: 'token@bridgeos.local' },
      trimmed
    )
    navigate(from, { replace: true })
  }

  const handleContinueTrusted = () => {
    setError('')
    logout()
    navigate(from, { replace: true })
  }

  return (
    <div className='min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4'>
      <Card className='w-full max-w-md'>
        <CardHeader title='BridgeOS Access' subtitle='Use local trusted mode on loopback, or paste an existing Bearer token.' />
        <CardContent>
          <div className='space-y-4'>
            {error && (
              <div className='bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm p-3 rounded-md'>
                {error}
              </div>
            )}
            <div className='rounded-md bg-blue-50 dark:bg-blue-900/20 p-4 text-sm text-blue-900 dark:text-blue-100'>
              <p className='font-medium'>Recommended for local development</p>
              <p className='mt-1'>
                Start the daemon with <code>BRIDGEOS_LOCAL_TRUSTED=true</code> and access BridgeOS from loopback.
                No username/password login API is implemented in this build.
              </p>
            </div>

            <Button type='button' className='w-full' onClick={handleContinueTrusted}>
              Continue In Local Trusted Mode
            </Button>

            <form onSubmit={handleUseToken} className='space-y-3 border-t border-gray-200 dark:border-gray-700 pt-4'>
              <div className='space-y-2'>
                <label htmlFor='token' className='text-sm font-medium text-gray-700 dark:text-gray-300'>
                  Existing Bearer Token
                </label>
                <textarea
                  id='token'
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  className='w-full min-h-28 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:text-white'
                  placeholder='Paste a JWT or API bearer token'
                />
              </div>
              <Button type='submit' variant='outline' className='w-full'>
                Use Bearer Token
              </Button>
            </form>

            <div className='text-sm text-gray-500 dark:text-gray-400 space-y-1'>
              <p>Supported access modes:</p>
              <p>1. Loopback + local trusted mode</p>
              <p>2. Existing Bearer token or API key</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
