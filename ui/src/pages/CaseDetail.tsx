import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Play, FileText, AlertCircle, Clock, CheckCircle, XCircle, Activity } from 'lucide-react'
import { useCase, useCaseEvents, useRunCase, useBuildReport } from '@/hooks/useApi'
import type { EventEnvelope } from '@/types'

function EventIcon({ type }: { type: string }) {
  if (type.includes('completed')) return <CheckCircle className='w-4 h-4 text-green-500' />
  if (type.includes('failed') || type.includes('error')) return <XCircle className='w-4 h-4 text-red-500' />
  if (type.includes('approval')) return <Clock className='w-4 h-4 text-yellow-500' />
  return <Activity className='w-4 h-4 text-blue-500' />
}

function formatTime(timestamp: string): string {
  try {
    return new Date(timestamp).toLocaleTimeString()
  } catch {
    return timestamp
  }
}

function formatPayload(payload: unknown): React.ReactNode {
  if (!payload) return null
  if (typeof payload !== 'object') return <span>{String(payload)}</span>
  try {
    return (
      <pre className='mt-1 text-xs text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-900 rounded p-2 overflow-x-auto'>
        {JSON.stringify(payload, null, 2)}
      </pre>
    )
  } catch {
    return <span>{String(payload)}</span>
  }
}

function EventTimeline({ events }: { events: EventEnvelope[] }) {
  if (events.length === 0) {
    return (
      <div className='text-center py-8 text-gray-500 dark:text-gray-400'>
        <Activity className='w-8 h-8 mx-auto mb-2 opacity-50' />
        <p>No events recorded yet</p>
      </div>
    )
  }

  return (
    <div className='space-y-3'>
      {events.map((event, index) => (
        <div key={`${event.sequence}-${index}`} className='flex items-start gap-3'>
          <div className='mt-1'>
            <EventIcon type={event.type} />
          </div>
          <div className='flex-1 min-w-0'>
            <div className='flex items-center justify-between'>
              <span className='font-medium text-sm truncate'>{event.type}</span>
              <span className='text-xs text-gray-500 ml-2'>{formatTime(event.created_at)}</span>
            </div>
            {event.payload ? formatPayload(event.payload) : null}
          </div>
        </div>
      ))}
    </div>
  )
}

export default function CaseDetail() {
  const { id } = useParams<{ id: string }>()
  const { data: caseRecord, isLoading, error } = useCase(id || '')
  const { data: events = [] } = useCaseEvents(id || '')
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
      <div className='flex items-center justify-center h-64'>
        <div className='animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600'></div>
      </div>
    )
  }

  if (error || !caseRecord) {
    return (
      <div>
        <div className='flex items-center gap-4 mb-6'>
          <Link to='/cases' className='text-gray-500 hover:text-gray-700'>
            <ArrowLeft className='w-5 h-5' />
          </Link>
          <h1 className='text-2xl font-bold'>Case Details</h1>
        </div>
        <div className='bg-red-50 border border-red-200 rounded-lg p-4'>
          <div className='flex items-center gap-2 text-red-800'>
            <AlertCircle className='w-5 h-5' />
            <p>Case not found or error loading case</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className='flex items-center gap-4 mb-6'>
        <Link to='/cases' className='text-gray-500 hover:text-gray-700'>
          <ArrowLeft className='w-5 h-5' />
        </Link>
        <h1 className='text-2xl font-bold'>{caseRecord.spec.title || caseRecord.id}</h1>
      </div>

      <div className='bg-white rounded-lg shadow p-6 mb-6'>
        <div className='grid grid-cols-2 gap-4'>
          <div>
            <p className='text-sm text-gray-500'>Status</p>
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
            <p className='text-sm text-gray-500'>Commands</p>
            <p className='font-medium'>{caseRecord.spec.commands?.length || 0}</p>
          </div>
          <div>
            <p className='text-sm text-gray-500'>Progress</p>
            <p className='font-medium'>{caseRecord.next_command} / {caseRecord.spec.commands?.length || 0}</p>
          </div>
          <div>
            <p className='text-sm text-gray-500'>Created</p>
            <p className='font-medium'>{new Date(caseRecord.created_at).toLocaleString()}</p>
          </div>
        </div>
      </div>

      <div className='flex gap-4 mb-6'>
        <button
          onClick={handleRun}
          disabled={runCase.isPending}
          className='flex items-center gap-2 bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 disabled:opacity-50'
        >
          <Play className='w-5 h-5' />
          {runCase.isPending ? 'Running...' : 'Run Case'}
        </button>
        <button
          onClick={handleBuildReport}
          disabled={buildReport.isPending}
          className='flex items-center gap-2 bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 disabled:opacity-50'
        >
          <FileText className='w-5 h-5' />
          {buildReport.isPending ? 'Building...' : 'Build Report'}
        </button>
      </div>

      <div className='bg-white rounded-lg shadow p-6'>
        <h2 className='text-lg font-semibold mb-4'>Event Timeline</h2>
        <EventTimeline events={events} />
      </div>
    </div>
  )
}