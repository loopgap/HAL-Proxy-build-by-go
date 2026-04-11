export interface CaseRecord {
  id: string
  title: string
  status: CaseStatus
  spec: CaseSpec
  next_command: number
  created_at: string
  updated_at: string
}

export type CaseStatus = 'draft' | 'ready' | 'running' | 'paused' | 'completed' | 'rejected'

export interface CaseSpec {
  title: string
  commands: CaseCommandSpec[]
}

export interface CaseCommandSpec {
  name: string
  action: string
  risk_class: RiskClass
  parameters?: Record<string, any>
}

export type RiskClass = 'observe' | 'mutate' | 'destructive' | 'exclusive'

export interface EventEnvelope {
  sequence: number
  case_id: string
  type: string
  payload: any
  created_at: string
}

export interface Approval {
  id: string
  case_id: string
  command_index: number
  command_name: string
  risk_class: RiskClass
  status: ApprovalStatus
  reason?: string
  decided_by?: string
  decided_at?: string
  created_at: string
}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected'

export interface ReportSummary {
  id: string
  case_id: string
  path: string
  command_count: number
  event_count: number
  created_at: string
}

export interface RunResult {
  case: CaseRecord
  status: string
  pending_approval?: Approval
}
