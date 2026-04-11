import { Link, useLocation } from 'react-router-dom'
import { Activity, FileText, CheckCircle, LayoutDashboard, Menu } from 'lucide-react'

interface LayoutProps {
  children: React.ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const location = useLocation()

  const navItems = [
    { path: '/', label: 'Dashboard', icon: LayoutDashboard },
    { path: '/cases', label: 'Cases', icon: Activity },
    { path: '/approvals', label: 'Approvals', icon: CheckCircle },
    { path: '/reports', label: 'Reports', icon: FileText },
  ]

  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 bg-gray-800 text-white flex flex-col">
        <div className="p-4 border-b border-gray-700">
          <h1 className="text-xl font-bold flex items-center gap-2">
            <Menu className="w-6 h-6" />
            BridgeOS
          </h1>
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

        <div className="p-4 border-t border-gray-700 text-sm text-gray-400">
          BridgeOS v1.0.0
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 bg-gray-100 p-6">
        {children}
      </main>
    </div>
  )
}
