import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Play, FileText, AlertCircle } from 'lucide-react'
import { useCase, useRunCase, useBuildReport } from '@/hooks/useApi'

export default function CaseDetail() {
  const { id } = useParams<{ id: string }>()
  const { data: caseRecord, isLoading, error } = useCase(id || '')
  const runCase = useRunCase()
  const buildReport = useBuildReport()

  const handleRun = () => {
    if (id && window.confirm('Are you sure you want to run this case?')) {
      runCase.mutate(id)
    }
  }

  const handleBuildReport = () => {
    if (id) {
      buildReport.mutate(id)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (error || !caseRecord) {
    return (
      <div>
        <div className="flex items-center gap-4 mb-6">
          <Link to="/cases" className="text-gray-500 hover:text-gray-700">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <h1 className="text-2xl font-bold">Case Details</h1>
        </div>
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center gap-2 text-red-800">
            <AlertCircle className="w-5 h-5" />
            <p>Case not found or error loading case</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <Link to="/cases" className="text-gray-500 hover:text-gray-700">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-bold">{caseRecord.title || caseRecord.id}</h1>
      </div>

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-gray-500">Status</p>
            <span className={`px-2 py-1 rounded-full text-sm ${
              caseRecord.status === 'completed' ? 'bg-green-100 text-green-800' :
              caseRecord.status === 'running' ? 'bg-blue-100 text-blue-800' :
              caseRecord.status === 'paused' ? 'bg-yellow-100 text-yellow-800' :
              'bg-gray-100 text-gray-800'
            }`}>
              {caseRecord.status}
            </span>
          </div>
          <div>
            <p className="text-sm text-gray-500">Commands</p>
            <p className="font-medium">{caseRecord.spec.commands?.length || 0}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Progress</p>
            <p className="font-medium">{caseRecord.next_command} / {caseRecord.spec.commands?.length || 0}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Created</p>
            <p className="font-medium">{new Date(caseRecord.created_at).toLocaleString()}</p>
          </div>
        </div>
      </div>

      <div className="flex gap-4">
        <button
          onClick={handleRun}
          disabled={runCase.isPending}
          className="flex items-center gap-2 bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 disabled:opacity-50"
        >
          <Play className="w-5 h-5" />
          {runCase.isPending ? 'Running...' : 'Run Case'}
        </button>
        <button
          onClick={handleBuildReport}
          disabled={buildReport.isPending}
          className="flex items-center gap-2 bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 disabled:opacity-50"
        >
          <FileText className="w-5 h-5" />
          {buildReport.isPending ? 'Building...' : 'Build Report'}
        </button>
      </div>
    </div>
  )
}
