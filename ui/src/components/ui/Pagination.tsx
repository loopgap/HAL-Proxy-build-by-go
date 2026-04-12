import { ChevronLeft, ChevronRight } from 'lucide-react'
import clsx from 'clsx'

interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
  className?: string
}

export function Pagination({ currentPage, totalPages, onPageChange, className }: PaginationProps) {
  if (totalPages <= 1) return null
  
  const hasPrev = currentPage > 1
  const hasNext = currentPage < totalPages
  
  const getVisiblePages = () => {
    const pages: (number | '...')[] = []
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) pages.push(i)
    } else {
      pages.push(1)
      if (currentPage > 3) pages.push('...')
      for (let i = Math.max(2, currentPage - 1); i <= Math.min(totalPages - 1, currentPage + 1); i++) {
        pages.push(i)
      }
      if (currentPage < totalPages - 2) pages.push('...')
      pages.push(totalPages)
    }
    return pages
  }
  
  return (
    <nav className={clsx('flex items-center justify-center gap-2', className)} aria-label='Pagination'>
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={!hasPrev}
        className={clsx(
          'p-2 rounded-lg border',
          hasPrev ? 'hover:bg-gray-100 dark:hover:bg-gray-700' : 'opacity-50 cursor-not-allowed'
        )}
        aria-label='Previous page'
      >
        <ChevronLeft className='w-5 h-5' />
      </button>
      
      <div className='flex items-center gap-1'>
        {getVisiblePages().map((page, idx) =>
          page === '...' ? (
            <span key={`ellipsis-${idx}`} className='px-2'>...</span>
          ) : (
            <button
              key={page}
              onClick={() => onPageChange(page)}
              className={clsx(
                'w-10 h-10 rounded-lg text-sm font-medium',
                page === currentPage
                  ? 'bg-primary-600 text-white'
                  : 'hover:bg-gray-100 dark:hover:bg-gray-700'
              )}
            >
              {page}
            </button>
          )
        )}
      </div>
      
      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={!hasNext}
        className={clsx(
          'p-2 rounded-lg border',
          hasNext ? 'hover:bg-gray-100 dark:hover:bg-gray-700' : 'opacity-50 cursor-not-allowed'
        )}
        aria-label='Next page'
      >
        <ChevronRight className='w-5 h-5' />
      </button>
      
      <span className='ml-2 text-sm text-gray-500'>
        Page {currentPage} of {totalPages}
      </span>
    </nav>
  )
}

export default Pagination