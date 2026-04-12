import { useState, useEffect } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Activity, FileText, CheckCircle, LayoutDashboard, Menu, LogOut } from 'lucide-react'
import { useAuthStore } from '@/store'

interface LayoutProps {
  children: React.ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()

  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 768) {
        setMobileMenuOpen(false)
      }
    }
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const navItems = [
    { path: '/', label: 'Dashboard', icon: LayoutDashboard },
    { path: '/cases', label: 'Cases', icon: Activity },
    { path: '/approvals', label: 'Approvals', icon: CheckCircle },
    { path: '/reports', label: 'Reports', icon: FileText },
  ]

  return (
    <div className="min-h-screen flex">
      {/* Overlay */}
      {mobileMenuOpen && (
        <div 
          className='fixed inset-0 bg-black/50 z-40 md:hidden'
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed inset-y-0 left-0 z-50 w-64 bg-gray-800 text-white flex flex-col
        transform transition-transform duration-200 ease-in-out
        md:relative md:translate-x-0
        ${mobileMenuOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        <div className="p-4 border-b border-gray-700">
          <div className='flex items-center gap-2'>
            <button 
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className='md:hidden p-2'
            >
              <Menu className="w-6 h-6" />
            </button>
            <h1 className="text-xl font-bold flex items-center gap-2">
              BridgeOS
            </h1>
          </div>
        </div>
        
        <nav className="flex-1 p-4">
          <ul className="space-y-2">
            {navItems.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.path
              return (
                <li key={item.path}>
                  <Link
                    to={item.path}
                    className={`flex items-center gap-3 px-4 py-2 rounded-lg transition-colors ${
                      isActive
                        ? 'bg-primary-600 text-white'
                        : 'text-gray-300 hover:bg-gray-700'
                    }`}
                  >
                    <Icon className="w-5 h-5" />
                    {item.label}
                  </Link>
                </li>
              )
            })}
          </ul>
        </nav>

        <div className='p-4 border-t border-gray-700 space-y-2'>
          {user && (
            <div className='text-sm text-gray-300 px-4'>
              {user.name}
            </div>
          )}
          <button
            onClick={handleLogout}
            className='flex items-center gap-3 w-full px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 rounded-lg transition-colors'
          >
            <LogOut className='w-4 h-4' />
            Logout
          </button>
          <div className='text-xs text-gray-500 px-4 pt-2 border-t border-gray-700'>
            BridgeOS v1.0.0
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 bg-gray-100 p-6">
        {children}
      </main>
    </div>
  )
}
