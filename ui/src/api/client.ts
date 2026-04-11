import axios from 'axios'
import type { CaseRecord, Approval, ReportSummary, RunResult, CaseSpec, EventEnvelope } from '@/types'

const API_BASE = '/v1'

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
})

// API response types
interface ApiResponse<T> {
  data: T | null
  error: string | null
}

function handleResponse<T>(response: any): ApiResponse<T> {
  if (response.data.error) {
    return { data: null, error: response.data.error }
  }
  return { data: response.data as T, error: null }
}

function handleError(error: any): ApiResponse<never> {
  const message = error.response?.data?.error || error.message || 'An error occurred'
  return { data: null, error: message }
}

// Case APIs
export async function getCases(): Promise<ApiResponse<CaseRecord[]>> {
  try {
    // For now, return empty list since ListCases is not implemented in backend
    return { data: [], error: null }
  } catch (error) {
    return handleError(error)
  }
}

export async function getCase(id: string): Promise<ApiResponse<CaseRecord>> {
  try {
    const response = await api.get(`/cases/${id}`)
    return handleResponse<CaseRecord>(response)
  } catch (error) {
    return handleError(error)
  }
}

export async function createCase(spec: CaseSpec): Promise<ApiResponse<CaseRecord>> {
  try {
    const response = await api.post('/cases', spec)
    return handleResponse<CaseRecord>(response)
  } catch (error) {
    return handleError(error)
  }
}

export async function runCase(id: string): Promise<ApiResponse<RunResult>> {
  try {
    const response = await api.post(`/cases/${id}:run`)
    return handleResponse<RunResult>(response)
  } catch (error) {
    return handleError(error)
  }
}

export async function getCaseEvents(id: string): Promise<ApiResponse<EventEnvelope[]>> {
  try {
    const response = await api.get(`/cases/${id}/events`)
    return handleResponse<EventEnvelope[]>(response)
  } catch (error) {
    return handleError(error)
  }
}

// Approval APIs
export async function getApprovals(caseId?: string): Promise<ApiResponse<Approval[]>> {
  try {
    const params = caseId ? { case_id: caseId } : {}
    const response = await api.get('/approvals', { params })
    return handleResponse<Approval[]>(response)
  } catch (error) {
    return handleError(error)
  }
}

export async function approveApproval(id: string): Promise<ApiResponse<Approval>> {
  try {
    const response = await api.post(`/approvals/${id}:approve`)
    return handleResponse<Approval>(response)
  } catch (error) {
    return handleError(error)
  }
}

export async function rejectApproval(id: string): Promise<ApiResponse<Approval>> {
  try {
    const response = await api.post(`/approvals/${id}:reject`)
    return handleResponse<Approval>(response)
  } catch (error) {
    return handleError(error)
  }
}

// Report APIs
export async function buildReport(caseId: string): Promise<ApiResponse<ReportSummary>> {
  try {
    const response = await api.post(`/reports/${caseId}:build`)
    return handleResponse<ReportSummary>(response)
  } catch (error) {
    return handleError(error)
  }
}

export { api }
