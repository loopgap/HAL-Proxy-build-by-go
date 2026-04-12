import { Activity, CheckCircle, AlertTriangle, FileText } from 'lucide-react'
import { useCases, useApprovals, useReports } from '@/hooks/useApi'

export default function Dashboard() {
  const { data: cases = [] } = useCases()
  const { data: approvals = [] } = useApprovals()
  const { data: reports = [] } = useReports()

  const pendingApprovals = approvals.filter(a => a.status === 'pending')
  const completedCases = cases.filter(c => c.status === 'completed')

  const stats = [
    { label: 'Total Cases', value: cases.length, icon: Activity, color: 'bg-blue-500' },
    { label: 'Pending Approvals', value: pendingApprovals.length, icon: AlertTriangle, color: 'bg-yellow-500' },
    { label: 'Completed', value: completedCases.length, icon: CheckCircle, color: 'bg-green-500' },
    { label: 'Reports', value: reports.length, icon: FileText, color: 'bg-purple-500' },
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
        <h2 className="text-lg font-semibold mb-4">Recent Cases</h2>
        {cases.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>No cases yet</p>
            <p className="text-sm mt-2">Create a case to get started</p>
          </div>
        ) : (
          <div className="space-y-4">
            {cases.slice(0, 5).map((c) => (
              <div key={c.id} className="flex items-center justify-between p-4 border rounded-lg">
                <div>
                  <p className="font-medium">{c.title || c.id}</p>
                  <p className="text-sm text-gray-500">Status: {c.status}</p>
                </div>
                <span className={`px-3 py-1 rounded-full text-sm ${
                  c.status === 'completed' ? 'bg-green-100 text-green-800' :
                  c.status === 'running' ? 'bg-blue-100 text-blue-800' :
                  'bg-gray-100 text-gray-800'
                }`}>
                  {c.status}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
