import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Play, FileText } from 'lucide-react'

export default function CaseDetail() {
  const { id } = useParams<{ id: string }>()

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <Link
          to="/cases"
          className="text-gray-500 hover:text-gray-700"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-bold">Case Details</h1>
      </div>

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <p className="text-gray-500">Case ID: {id}</p>
        <p className="text-gray-500 mt-2">Loading case details...</p>
      </div>

      <div className="flex gap-4">
        <button className="flex items-center gap-2 bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700">
          <Play className="w-5 h-5" />
          Run Case
        </button>
        <button className="flex items-center gap-2 bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700">
          <FileText className="w-5 h-5" />
          Build Report
        </button>
      </div>
    </div>
  )
}
