import { useEffect } from 'react'
import clsx from 'clsx'
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-react'

type ToastType = 'success' | 'error' | 'info' | 'warning'

interface ToastProps {
  id: string
  message: string
  type?: ToastType
  duration?: number
  onClose: (id: string) => void
}

const icons = {
  success: CheckCircle,
  error: AlertCircle,
  info: Info,
  warning: AlertTriangle,
}

const styles = {
  success: 'bg-green-50 border-green-200 text-green-800',
  error: 'bg-red-50 border-red-200 text-red-800',
  info: 'bg-blue-50 border-blue-200 text-blue-800',
  warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
}

const iconStyles = {
  success: 'text-green-500',
  error: 'text-red-500',
  info: 'text-blue-500',
  warning: 'text-yellow-500',
}

export function Toast({
  id,
  message,
  type = 'info',
  duration = 4000,
  onClose,
}: ToastProps) {
  const Icon = icons[type]

  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onClose(id)
      }, duration)
      return () => clearTimeout(timer)
    }
  }, [id, duration, onClose])

  return (
    <div
      className={clsx(
        'flex items-center gap-3 px-4 py-3 rounded-lg border shadow-lg',
        'transform transition-all duration-300 ease-in-out',
        'animate-slide-in-right',
        styles[type]
      )}
      role='alert'
    >
      <Icon className={clsx('w-5 h-5 flex-shrink-0', iconStyles[type])} />
      <p className='flex-1 text-sm font-medium'>{message}</p>
      <button
        onClick={() => onClose(id)}
        className='flex-shrink-0 p-1 rounded hover:bg-black/5 transition-colors'
        aria-label='Close'
      >
        <X className='w-4 h-4' />
      </button>
    </div>
  )
}

interface ToastContainerProps {
  toasts: Array<{ id: string; message: string; type: ToastType }>
  onClose: (id: string) => void
}

export function ToastContainer({ toasts, onClose }: ToastContainerProps) {
  return (
    <div className='fixed top-4 right-4 z-[100] flex flex-col gap-2 max-w-sm w-full'>
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          id={toast.id}
          message={toast.message}
          type={toast.type}
          onClose={onClose}
        />
      ))}
    </div>
  )
}

export default Toast