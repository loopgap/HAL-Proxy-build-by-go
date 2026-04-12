import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useApprovals, useApproveApproval, useRejectApproval } from '@/hooks/useApi'
import { Table, type Column } from '@/components/ui/Table'
import { Button } from '@/components/ui/Button'
import { CheckCircle, XCircle, Clock, AlertTriangle } from 'lucide-react'
import type { Approval } from '@/types'

function StatusBadge({ status }: { status: string }) {
  const styles = {
    pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    approved: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    rejected: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  }
  const icons = {
    pending: Clock,
    approved: CheckCircle,
    rejected: XCircle,
  }
  const Icon = icons[status as keyof typeof icons] || Clock
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${styles[status as keyof typeof styles] || styles.pending}`}>
      <Icon className='w-3 h-3' />
      {status}
    </span>
  )
}

function RiskBadge({ riskClass }: { riskClass: string }) {
  const styles = {
    observe: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
    mutate: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
    destructive: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
    exclusive: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  }
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${styles[riskClass as keyof typeof styles] || styles.observe}`}>
      <AlertTriangle className='w-3 h-3' />
      {riskClass}
    </span>
  )
}

export default function ApprovalList() {
  const { data: approvals, isLoading, error } = useApprovals()
  const approveMutation = useApproveApproval()
  const rejectMutation = useRejectApproval()
  const [filter, setFilter] = useState<string>('all')

  const filteredApprovals = (approvals || []).filter((a) => {
    if (filter === 'all') return true
    return a.status === filter
  })

  const handleApprove = async (id: string) => {
    if (window.confirm('Are you sure you want to approve this operation?')) {
      await approveMutation.mutateAsync(id)
    }
  }

  const handleReject = async (id: string) => {
    if (window.confirm('Are you sure you want to reject this operation?')) {
      await rejectMutation.mutateAsync(id)
    }
  }

  const columns: Column<Approval>[] = [
    {
      key: 'id',
      title: 'ID',
      render: (a) => <span className='font-mono text-xs'>{a.id.slice(0, 8)}</span>,
    },
    {
      key: 'case_id',
      title: 'Case',
      render: (a) => (
        <Link to={`/cases/${a.case_id}`} className='text-primary-600 hover:underline'>
          {a.case_id.slice(0, 8)}...
        </Link>
      ),
    },
    {
      key: 'risk_class',
      title: 'Risk',
      render: (a) => <RiskBadge riskClass={a.risk_class} />,
    },
    {
      key: 'status',
      title: 'Status',
      render: (a) => <StatusBadge status={a.status} />,
    },
    {
      key: 'created_at',
      title: 'Requested At',
      render: (a) => new Date(a.created_at).toLocaleString(),
    },
    {
      key: 'actions',
      title: 'Actions',
      align: 'right',
      render: (a) =>
        a.status === 'pending' && (
          <div className='flex items-center justify-end gap-2'>
            <Button
              size='sm'
              variant='ghost'
              onClick={() => handleApprove(a.id)}
              disabled={approveMutation.isPending || rejectMutation.isPending}
              className='text-green-600 hover:text-green-700 hover:bg-green-50'
            >
              <CheckCircle className='w-4 h-4 mr-1' />
              Approve
            </Button>
            <Button
              size='sm'
              variant='ghost'
              onClick={() => handleReject(a.id)}
              disabled={approveMutation.isPending || rejectMutation.isPending}
              className='text-red-600 hover:text-red-700 hover:bg-red-50'
            >
              <XCircle className='w-4 h-4 mr-1' />
              Reject
            </Button>
          </div>
        ),
    },
  ]

  return (
    <div>
      <div className='flex items-center justify-between mb-6'>
        <h1 className='text-2xl font-bold'>Approvals</h1>
        <div className='flex items-center gap-2'>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className='px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800'
          >
            <option value='all'>All</option>
            <option value='pending'>Pending</option>
            <option value='approved'>Approved</option>
            <option value='rejected'>Rejected</option>
          </select>
        </div>
      </div>

      {error && (
        <div className='bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-4 rounded-lg mb-4'>
          Failed to load approvals: {error.message}
        </div>
      )}

      <Table
        columns={columns}
        data={filteredApprovals}
        keyExtractor={(a) => a.id}
        isLoading={isLoading}
        emptyMessage='No approvals found'
        hoverable
      />
    </div>
  )
}