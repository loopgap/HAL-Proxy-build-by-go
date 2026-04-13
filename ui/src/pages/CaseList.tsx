import { Link } from 'react-router-dom'
import { Plus, Activity } from 'lucide-react'
import { useCases } from '@/hooks/useApi'
import { Table, type Column } from '@/components/ui/Table'
import { StatusBadge } from '@/components/ui/StatusBadge'
import type { CaseRecord } from '@/types'

export default function CaseList() {
  const { data: cases = [], isLoading, error } = useCases()

  const columns: Column<CaseRecord>[] = [
    {
      key: 'title',
      title: 'Title',
      render: (c) => (
        <Link to={`/cases/${c.id}`} className='text-primary-600 hover:underline font-medium'>
          {c.spec.title || c.id.slice(0, 8)}
        </Link>
      ),
    },
    {
      key: 'status',
      title: 'Status',
      render: (c) => <StatusBadge status={c.status} />,
    },
    {
      key: 'commands',
      title: 'Commands',
      render: (c) => c.spec.commands?.length || 0,
    },
    {
      key: 'created_at',
      title: 'Created',
      render: (c) => new Date(c.created_at).toLocaleDateString(),
    },
    {
      key: 'actions',
      title: 'Actions',
      align: 'right',
      render: (c) => (
        <Link to={`/cases/${c.id}`} className='text-primary-600 hover:underline text-sm'>
          View Details
        </Link>
      ),
    },
  ]

  return (
    <div>
      <div className='flex justify-between items-center mb-6'>
        <h1 className='text-2xl font-bold'>Cases</h1>
        <Link
          to='/cases/new'
          className='flex items-center gap-2 bg-primary-600 text-white px-4 py-2 rounded-lg hover:bg-primary-700'
        >
          <Plus className='w-4 h-4' />
          New Case
        </Link>
      </div>

      {error && (
        <div className='bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-4 rounded-lg mb-4'>
          Error loading cases: {error.message}
        </div>
      )}

      <Table
        columns={columns}
        data={cases}
        keyExtractor={(c) => c.id}
        isLoading={isLoading}
        emptyMessage='No cases yet'
        hoverable
      />

      {!isLoading && cases.length === 0 && (
        <div className='mt-4 text-center text-sm text-gray-500 dark:text-gray-400'>
          <Activity className='w-8 h-8 mx-auto mb-2 opacity-50' />
          <p>Create your first case using the CLI or API</p>
          <p className='mt-1'>
            <code className='bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded'>bridge create case --spec '{"{...}"}'</code>
          </p>
        </div>
      )}
    </div>
  )
}