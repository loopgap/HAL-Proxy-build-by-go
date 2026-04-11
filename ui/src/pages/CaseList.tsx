import { Link } from 'react-router-dom'
import { Plus, Activity } from 'lucide-react'

export default function CaseList() {
  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Cases</h1>
        <Link
          to="/cases/new"
          className="flex items-center gap-2 bg-primary-600 text-white px-4 py-2 rounded-lg hover:bg-primary-700"
        >
          <Plus className="w-5 h-5" />
          New Case
        </Link>
      </div>

      <div className="bg-white rounded-lg shadow">
        <div className="text-center py-12 text-gray-500">
          <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <p>No cases yet</p>
          <p className="text-sm mt-2">Create your first case using the CLI or API</p>
        </div>
      </div>
    </div>
  )
}
