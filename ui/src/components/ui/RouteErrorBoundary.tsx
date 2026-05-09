import { type ReactNode, Suspense } from 'react'
import { ErrorBoundary } from './ErrorBoundary'
import { LoadingSpinner } from './LoadingSpinner'

interface RouteErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
}

export function RouteErrorBoundary({ children, fallback }: RouteErrorBoundaryProps) {
  return (
    <ErrorBoundary fallback={fallback}>
      <Suspense fallback={<PageLoader />}>{children}</Suspense>
    </ErrorBoundary>
  )
}

function PageLoader() {
  return (
    <div className='flex items-center justify-center h-64'>
      <LoadingSpinner size='lg' label='Loading page...' />
    </div>
  )
}

export default RouteErrorBoundary
