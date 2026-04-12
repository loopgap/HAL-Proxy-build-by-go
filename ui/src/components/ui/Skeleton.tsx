import { type HTMLAttributes } from 'react'
import clsx from 'clsx'

interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'text' | 'circular' | 'rectangular'
  width?: string | number
  height?: string | number
  lines?: number
}

export function Skeleton({
  variant = 'rectangular',
  width,
  height,
  className,
  ...props
}: SkeletonProps) {
  const style: React.CSSProperties = {}
  if (width) style.width = typeof width === 'number' ? `${width}px` : width
  if (height) style.height = typeof height === 'number' ? `${height}px` : height
  
  return (
    <div
      className={clsx(
        'animate-pulse bg-gray-200 dark:bg-gray-700',
        {
          'rounded-full': variant === 'circular',
          'rounded-md': variant === 'rectangular',
          'rounded': variant === 'text',
          'h-4': variant === 'text',
          'w-full': variant === 'text',
        },
        className
      )}
      style={style}
      {...props}
    />
  )
}

export function SkeletonText({ lines = 3, ...props }: SkeletonProps & { lines?: number }) {
  return (
    <div className='space-y-2'>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton key={i} variant='text' {...props} />
      ))}
    </div>
  )
}

export default Skeleton
