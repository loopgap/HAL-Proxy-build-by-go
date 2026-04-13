import { CheckCircle, XCircle, Clock, AlertTriangle } from 'lucide-react'

type CaseStatus = 'completed' | 'running' | 'paused' | 'rejected' | 'ready' | 'draft'
type ApprovalStatus = 'pending' | 'approved' | 'rejected'
type RiskClass = 'observe' | 'mutate' | 'destructive' | 'exclusive'

interface StatusBadgeProps {
  status: string
  variant?: 'case' | 'approval'
}

const caseStatusStyles: Record<CaseStatus, string> = {
  completed: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  running: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  paused: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  rejected: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  ready: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
  draft: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
}

const approvalStatusStyles: Record<ApprovalStatus, string> = {
  pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  approved: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  rejected: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
}

const approvalStatusIcons: Record<ApprovalStatus, typeof Clock> = {
  pending: Clock,
  approved: CheckCircle,
  rejected: XCircle,
}

const riskClassStyles: Record<RiskClass, string> = {
  observe: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  mutate: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
  destructive: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  exclusive: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
}

export function StatusBadge({ status, variant = 'case' }: StatusBadgeProps) {
  if (variant === 'approval') {
    const styles = approvalStatusStyles as Record<string, string>
    const icons = approvalStatusIcons as Record<string, typeof Clock>
    const Icon = icons[status] || Clock
    const style = styles[status] || styles.pending

    return (
      <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${style}`}>
        <Icon className='w-3 h-3' />
        {status}
      </span>
    )
  }

  const styles = caseStatusStyles as Record<string, string>
  const style = styles[status] || styles.draft

  return (
    <span className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${style}`}>
      {status}
    </span>
  )
}

interface RiskBadgeProps {
  riskClass: string
}

export function RiskBadge({ riskClass }: RiskBadgeProps) {
  const styles = riskClassStyles as Record<string, string>
  const style = styles[riskClass] || styles.observe

  return (
    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${style}`}>
      <AlertTriangle className='w-3 h-3' />
      {riskClass}
    </span>
  )
}

export default StatusBadge