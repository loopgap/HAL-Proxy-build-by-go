import { Link } from 'react-router-dom'
import { Home, ArrowLeft, Search } from 'lucide-react'
import Button from '@/components/ui/Button'

export function NotFoundPage() {
  return (
    <div className='min-h-[60vh] flex items-center justify-center p-6'>
      <div className='text-center max-w-md'>
        <div className='mb-8'>
          <div className='text-[150px] font-bold text-gray-100 leading-none select-none'>
            404
          </div>
          <div className='-mt-16 relative'>
            <div className='w-24 h-24 mx-auto bg-primary-100 rounded-full flex items-center justify-center'>
              <Search className='w-12 h-12 text-primary-600' />
            </div>
          </div>
        </div>

        <h1 className='text-3xl font-bold text-gray-900 mb-4'>
          Page Not Found
        </h1>
        <p className='text-gray-600 mb-8'>
          The page you are looking for might have been removed, had its name changed,
          or is temporarily unavailable.
        </p>

        <div className='flex flex-col sm:flex-row items-center justify-center gap-4'>
          <Button
            variant='outline'
            leftIcon={<ArrowLeft className='w-4 h-4' />}
            onClick={() => window.history.back()}
          >
            Go Back
          </Button>
          <Link to='/'>
            <Button
              variant='primary'
              leftIcon={<Home className='w-4 h-4' />}
            >
              Back to Home
            </Button>
          </Link>
        </div>
      </div>
    </div>
  )
}

export default NotFoundPage