import { Activity, CheckCircle, AlertTriangle, FileText } from 'lucide-react'

export default function Dashboard() {
  const stats = [
    { label: 'Total Cases', value: '0', icon: Activity, color: 'bg-blue-500' },
    { label: 'Pending Approvals', value: '0', icon: AlertTriangle, color: 'bg-yellow-500' },
    { label: 'Completed', value: '0', icon: CheckCircle, color: 'bg-green-500' },
    { label: 'Reports', value: '0', icon: FileText, color: 'bg-purple-500' },
  ]

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
      
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {stats.map((stat) => {
          const Icon = stat.icon
          return (
            <div key={stat.label} className="bg-white rounded-lg shadow p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-sm">{stat.label}</p>
                  <p className="text-3xl font-bold mt-1">{stat.value}</p>
                </div>
                <div className={`${stat.color} p-3 rounded-lg`}>
                  <Icon className="w-6 h-6 text-white" />
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Recent Activity */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">Recent Activity</h2>
        <div className="text-center py-12 text-gray-500">
          <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <p>No recent activity</p>
          <p className="text-sm mt-2">Create a case to get started</p>
        </div>
      </div>
    </div>
  )
}
