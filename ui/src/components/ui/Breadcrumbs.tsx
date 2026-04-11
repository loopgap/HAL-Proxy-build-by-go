import { Link, useLocation } from 'react-router-dom'
import { ChevronRight, Home } from 'lucide-react'
import clsx from 'clsx'

export interface BreadcrumbItem {
  label: string
  path?: string
  icon?: React.ComponentType<{ className?: string }>
}

interface BreadcrumbsProps {
  items: BreadcrumbItem[]
  separator?: React.ReactNode
  className?: string
  showHome?: boolean
}

export function Breadcrumbs({
  items,
  separator = <ChevronRight className='w-4 h-4 text-gray-400' />,
  className,
  showHome = true,
}: BreadcrumbsProps) {
  const allItems = showHome
    ? [{ label: 'Home', path: '/', icon: Home }, ...items]
    : items

  return (
    <nav aria-label='Breadcrumb' className={clsx('flex items-center', className)}>
      <ol className='flex items-center space-x-2'>
        {allItems.map((item, index) => {
          const isLast = index === allItems.length - 1
          const Icon = item.icon

          return (
            <li key={index} className='flex items-center'>
              {!isLast ? (
                <>
                  {item.path ? (
                    <Link
                      to={item.path}
                      className='flex items-center gap-1 text-sm text-gray-600 hover:text-primary-600 transition-colors'
                    >
                      {Icon && <Icon className='w-4 h-4' />}
                      <span>{item.label}</span>
                    </Link>
                  ) : (
                    <span className='flex items-center gap-1 text-sm text-gray-500'>
                      {Icon && <Icon className='w-4 h-4' />}
                      <span>{item.label}</span>
                    </span>
                  )}
                  <span className='mx-2'>{separator}</span>
                </>
              ) : (
                <span className='text-sm font-medium text-gray-900'>
                  {Icon && <Icon className='w-4 h-4 inline mr-1' />}
                  {item.label}
                </span>
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}

export function useBreadcrumbs(): BreadcrumbItem[] {
  const location = useLocation()
  const pathnames = location.pathname.split('/').filter((x) => x)

  return pathnames.map((pathname, index) => {
    const routePath = '/' + pathnames.slice(0, index + 1).join('/')
    const label = pathname.charAt(0).toUpperCase() + pathname.slice(1).replace(/-/g, ' ')

    return {
      label,
      path: routePath,
    }
  })
}

export default Breadcrumbs