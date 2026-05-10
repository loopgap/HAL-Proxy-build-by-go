import { type ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store'

interface GuardProps {
  children: ReactNode
  requireAuth?: boolean
  requiredRoles?: string[]
  fallbackPath?: string
}

export function AuthGuard({
  children,
  requireAuth = true,
  requiredRoles = [],
  fallbackPath = '/login',
}: GuardProps) {
  const { user, token } = useAuthStore()
  const location = useLocation()

  if (requireAuth && !token) {
    return <Navigate to={fallbackPath} state={{ from: location }} replace />
  }

  if (requiredRoles.length > 0 && user) {
    const hasRequiredRole = requiredRoles.includes(user.role) || user.role === 'admin'
    if (!hasRequiredRole) {
      return <Navigate to='/403' replace />
    }
  }

  return <>{children}</>
}

export default AuthGuard