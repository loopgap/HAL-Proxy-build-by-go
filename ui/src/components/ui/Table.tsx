import { type ReactNode, type HTMLAttributes } from 'react'
import clsx from 'clsx'
import { LoadingSpinner } from './LoadingSpinner'

export interface Column<T> {
  key: string
  title: string
  render?: (item: T, index: number) => ReactNode
  className?: string
  width?: string
  align?: 'left' | 'center' | 'right'
}

interface TableProps<T> extends HTMLAttributes<HTMLDivElement> {
  columns: Column<T>[]
  data: T[]
  keyExtractor: (item: T) => string
  emptyMessage?: string
  isLoading?: boolean
  onRowClick?: (item: T) => void
  striped?: boolean
  hoverable?: boolean
}

const alignClasses = {
  left: 'text-left',
  center: 'text-center',
  right: 'text-right',
}

export function Table<T>({
  columns,
  data,
  keyExtractor,
  emptyMessage = 'No data available',
  isLoading = false,
  onRowClick,
  striped = false,
  hoverable = false,
  className,
  ...props
}: TableProps<T>) {
  if (isLoading) {
    return (
      <div className={clsx('bg-white rounded-xl shadow-sm overflow-hidden', className)} {...props}>
        <div className='flex items-center justify-center h-32'>
          <LoadingSpinner size='md' />
        </div>
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className={clsx('bg-white rounded-xl shadow-sm overflow-hidden', className)} {...props}>
        <div className='flex items-center justify-center h-32 text-gray-500'>
          {emptyMessage}
        </div>
      </div>
    )
  }

  return (
    <div className={clsx('bg-white rounded-xl shadow-sm overflow-hidden', className)} {...props}>
      <div className='overflow-x-auto'>
        <table className='w-full'>
          <thead>
            <tr className='bg-gray-50 border-b border-gray-200'>
              {columns.map((column) => (
                <th
                  key={column.key}
                  className={clsx(
                    'px-4 py-3 text-xs font-semibold text-gray-600 uppercase tracking-wider',
                    column.align ? alignClasses[column.align] : 'text-left',
                    column.className
                  )}
                  style={{ width: column.width }}
                >
                  {column.title}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.map((item, index) => (
              <tr
                key={keyExtractor(item)}
                className={clsx(
                  'border-b border-gray-100',
                  striped && index % 2 === 1 && 'bg-gray-50/50',
                  hoverable && 'hover:bg-gray-50 transition-colors',
                  onRowClick && 'cursor-pointer'
                )}
                onClick={() => onRowClick?.(item)}
              >
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className={clsx(
                      'px-4 py-3 text-sm text-gray-700',
                      column.align ? alignClasses[column.align] : 'text-left',
                      column.className
                    )}
                  >
                    {column.render
                      ? column.render(item, index)
                      : String((item as Record<string, unknown>)[column.key] ?? '')}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default Table