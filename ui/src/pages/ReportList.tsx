import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useReports } from '@/hooks/useApi'
import { getReportContent } from '@/api/client'
import { Table, type Column } from '@/components/ui/Table'
import { Button } from '@/components/ui/Button'
import { Download, Eye, RefreshCw } from 'lucide-react'
import toast from 'react-hot-toast'
import type { ReportSummary } from '@/types'

function formatDate(dateStr: string): string {
  try {
    return new Date(dateStr).toLocaleString()
  } catch {
    return dateStr
  }
}

export default function ReportList() {
  const { data: reports, isLoading, error, refetch } = useReports()
  const [activeReportId, setActiveReportId] = useState<string | null>(null)

  async function withReportContent(report: ReportSummary, action: (body: string) => void) {
    setActiveReportId(report.id)
    try {
      const result = await getReportContent(report.id)
      if (result.error || !result.data) {
        throw new Error(result.error || 'Failed to load report content')
      }
      action(result.data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to load report content')
    } finally {
      setActiveReportId(null)
    }
  }

  async function handleViewReport(report: ReportSummary) {
    await withReportContent(report, (body) => {
      const blob = new Blob([body], { type: 'text/markdown;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      window.open(url, '_blank')
      setTimeout(() => URL.revokeObjectURL(url), 60_000)
    })
  }

  async function handleDownloadReport(report: ReportSummary) {
    await withReportContent(report, (body) => {
      const blob = new Blob([body], { type: 'text/markdown;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = report.path.split(/[\\/]/).pop() || `report-${report.id.slice(0, 8)}.md`
      link.click()
      setTimeout(() => URL.revokeObjectURL(url), 60_000)
    })
  }

  const columns: Column<ReportSummary>[] = [
    {
      key: 'id',
      title: 'ID',
      render: (r) => <span className='font-mono text-xs'>{r.id.slice(0, 8)}</span>,
    },
    {
      key: 'case_id',
      title: 'Case',
      render: (r) => (
        <Link to={`/cases/${r.case_id}`} className='text-primary-600 hover:underline'>
          {r.case_id.slice(0, 8)}...
        </Link>
      ),
    },
    {
      key: 'created_at',
      title: 'Generated At',
      render: (r) => formatDate(r.created_at),
    },
    {
      key: 'command_count',
      title: 'Commands',
      align: 'center',
    },
    {
      key: 'event_count',
      title: 'Events',
      align: 'center',
    },
    {
      key: 'actions',
      title: 'Actions',
      align: 'right',
      render: (r) => (
        <div className='flex items-center justify-end gap-2'>
          <Button
            size='sm'
            variant='ghost'
            onClick={() => { void handleViewReport(r) }}
            title='View Report'
            disabled={activeReportId === r.id}
          >
            <Eye className='w-4 h-4' />
          </Button>
          <Button
            size='sm'
            variant='ghost'
            onClick={() => { void handleDownloadReport(r) }}
            title='Download Report'
            disabled={activeReportId === r.id}
          >
            <Download className='w-4 h-4' />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div>
      <div className='flex items-center justify-between mb-6'>
        <h1 className='text-2xl font-bold'>Reports</h1>
        <Button variant='outline' size='sm' onClick={() => refetch()}>
          <RefreshCw className='w-4 h-4 mr-2' />
          Refresh
        </Button>
      </div>

      {error && (
        <div className='bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-4 rounded-lg mb-4'>
          Failed to load reports: {error.message}
        </div>
      )}

      <Table
        columns={columns}
        data={reports || []}
        keyExtractor={(r) => r.id}
        isLoading={isLoading}
        emptyMessage='No reports yet. Run a case to generate a report.'
        hoverable
      />

      {!isLoading && (reports || []).length > 0 && (
        <div className='mt-4 text-sm text-gray-500 dark:text-gray-400'>
          <p>Reports are read from the BridgeOS store. {reports?.length} report(s) found.</p>
          <p className='mt-1'>
            <Link to='/cases' className='text-primary-600 hover:underline'>Go to Cases</Link>
            {' '}to run a case and generate a new report.
          </p>
        </div>
      )}
    </div>
  )
}
