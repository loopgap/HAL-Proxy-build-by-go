import { useEffect, useCallback, type ReactNode } from 'react'
import clsx from 'clsx'
import { X } from 'lucide-react'
import { createPortal } from 'react-dom'

interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: ReactNode
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  showCloseButton?: boolean
  closeOnOverlayClick?: boolean
  closeOnEscape?: boolean
  className?: string
}

const sizeClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-4xl',
}

export function Modal({
  isOpen,
  onClose,
  title,
  children,
  size = 'md',
  showCloseButton = true,
  closeOnOverlayClick = true,
  closeOnEscape = true,
  className,
}: ModalProps) {
  const handleEscape = useCallback(
    (event: KeyboardEvent) => {
      if (closeOnEscape && event.key === 'Escape') {
        onClose()
      }
    },
    [closeOnEscape, onClose]
  )

  useEffect(() => {
    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
      document.body.style.overflow = 'hidden'
    }

    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = ''
    }
  }, [isOpen, handleEscape])

  if (!isOpen) return null

  const modalContent = (
    <div
      className='fixed inset-0 z-50 flex items-center justify-center p-4'
      role='dialog'
      aria-modal='true'
      aria-labelledby={title ? 'modal-title' : undefined}
    >
      <div
        className='absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity'
        onClick={closeOnOverlayClick ? onClose : undefined}
        aria-hidden='true'
      />

      <div
        className={clsx(
          'relative bg-white rounded-xl shadow-2xl w-full',
          'transform transition-all duration-300 ease-out',
          'animate-modal-in',
          sizeClasses[size],
          className
        )}
      >
        {(title || showCloseButton) && (
          <div className='flex items-center justify-between px-6 py-4 border-b border-gray-200'>
            {title && (
              <h2 id='modal-title' className='text-lg font-semibold text-gray-900'>
                {title}
              </h2>
            )}
            {showCloseButton && (
              <button
                onClick={onClose}
                className='p-2 rounded-lg hover:bg-gray-100 transition-colors'
                aria-label='Close modal'
              >
                <X className='w-5 h-5 text-gray-500' />
              </button>
            )}
          </div>
        )}

        <div className='px-6 py-4 max-h-[calc(100vh-200px)] overflow-y-auto'>
          {children}
        </div>
      </div>
    </div>
  )

  return createPortal(modalContent, document.body)
}

export default Modal