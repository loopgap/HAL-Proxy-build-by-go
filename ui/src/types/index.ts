// CaseRecord represents a BridgeOS case record
export interface CaseRecord {
  id: string
  title: string
  status: CaseStatus
  spec: CaseSpec
  next_command: number
  created_at: string
  updated_at: string
}

// CaseStatus represents the possible states of a case
export type CaseStatus = 'draft' | 'ready' | 'running' | 'paused' | 'completed' | 'rejected'

// CaseSpec defines the specification for a case
export interface CaseSpec {
  title: string
  commands: CaseCommandSpec[]
}

// CaseCommandSpec defines a single command in a case
export interface CaseCommandSpec {
  name: string
  action: string
  risk_class: RiskClass
  parameters?: Record<string, unknown>
}

// RiskClass represents the risk level of a command
export type RiskClass = 'observe' | 'mutate' | 'destructive' | 'exclusive'

// EventEnvelope represents an event in the system
export interface EventEnvelope {
  sequence: number
  case_id: string
  type: string
  payload: unknown
  created_at: string
}

// Approval represents an approval request for a case command
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

// ApprovalStatus represents the possible states of an approval
export type ApprovalStatus = 'pending' | 'approved' | 'rejected'

// ReportSummary represents a generated report
export interface ReportSummary {
  id: string
  case_id: string
  path: string
  command_count: number
  event_count: number
  created_at: string
}

// RunResult represents the result of running a case
export interface RunResult {
  case: CaseRecord
  status: RunStatus
  pending_approval?: Approval
}

// RunStatus represents the possible outcomes of running a case
export type RunStatus = 
  | 'already_completed' 
  | 'awaiting_approval' 
  | 'rejected' 
  | 'completed'

// API response wrapper
export interface ApiResponse<T> {
  data: T | null
  error: string | null
  status?: number
}

export interface CasesListResponse {
  items: CaseRecord[]
  next_cursor: string
  has_more: boolean
}

export interface CaseEventsResponse {
  items: EventEnvelope[]
  total: number
  limit: number
  offset: number
}
