import clsx from 'clsx'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
  fullScreen?: boolean
  label?: string
}

const sizeClasses = {
  sm: 'h-4 w-4 border-2',
  md: 'h-8 w-8 border-2',
  lg: 'h-12 w-12 border-3',
  xl: 'h-16 w-16 border-4',
}

export function LoadingSpinner({
  size = 'md',
  className,
  fullScreen = false,
  label = 'Loading...',
}: LoadingSpinnerProps) {
  const spinner = (
    <div className={clsx('flex flex-col items-center justify-center gap-3', className)}>
      <div
        className={clsx(
          'animate-spin rounded-full border-primary-200 border-t-primary-600',
          sizeClasses[size]
        )}
        role='status'
        aria-label={label}
      />
      {label && <span className='text-sm text-gray-500'>{label}</span>}
    </div>
  )

  if (fullScreen) {
    return (
      <div className='fixed inset-0 flex items-center justify-center bg-white/80 backdrop-blur-sm z-50'>
        {spinner}
      </div>
    )
  }

  return spinner
}

export default LoadingSpinner