import { type HTMLAttributes, type ReactNode } from 'react'
import clsx from 'clsx'

type CardVariant = 'default' | 'bordered' | 'elevated' | 'outlined'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: CardVariant
  padding?: 'none' | 'sm' | 'md' | 'lg'
  hoverable?: boolean
  children: ReactNode
}

const variantClasses: Record<CardVariant, string> = {
  default: 'bg-white shadow-sm',
  bordered: 'bg-white border border-gray-200',
  elevated: 'bg-white shadow-lg',
  outlined: 'bg-transparent border-2 border-gray-300',
}

const paddingClasses = {
  none: '',
  sm: 'p-3',
  md: 'p-4',
  lg: 'p-6',
}

export function Card({
  variant = 'default',
  padding = 'md',
  hoverable = false,
  className,
  children,
  ...props
}: CardProps) {
  return (
    <div
      className={clsx(
        'rounded-xl transition-all duration-200',
        variantClasses[variant],
        paddingClasses[padding],
        hoverable && 'hover:shadow-md cursor-pointer',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}

interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {
  title?: string
  subtitle?: string
  action?: ReactNode
  children?: ReactNode
}

export function CardHeader({
  title,
  subtitle,
  action,
  children,
  className,
  ...props
}: CardHeaderProps) {
  return (
    <div className={clsx('flex items-start justify-between mb-4', className)} {...props}>
      <div>
        {title && <h3 className='text-lg font-semibold text-gray-900'>{title}</h3>}
        {subtitle && <p className='text-sm text-gray-500 mt-1'>{subtitle}</p>}
      </div>
      {action && <div>{action}</div>}
      {children}
    </div>
  )
}

interface CardContentProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
}

export function CardContent({ children, className, ...props }: CardContentProps) {
  return (
    <div className={clsx(className)} {...props}>{children}</div>
  )
}

interface CardFooterProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
}

export function CardFooter({ children, className, ...props }: CardFooterProps) {
  return (
    <div className={clsx('flex items-center justify-end gap-2 mt-4 pt-4 border-t border-gray-100', className)} {...props}>
      {children}
    </div>
  )
}

export default Card