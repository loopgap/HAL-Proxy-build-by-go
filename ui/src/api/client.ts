import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig, AxiosError } from 'axios'
import type {
  ApiResponse,
  CaseRecord,
  Approval,
  ReportSummary,
  RunResult,
  CaseSpec,
  CasesListResponse,
  CaseEventsResponse,
} from '@/types'

export interface RequestConfig extends AxiosRequestConfig {
  retries?: number
  retryDelay?: number
}

const DEFAULT_TIMEOUT = 30000
const DEFAULT_RETRIES = 3
const DEFAULT_RETRY_DELAY = 1000

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/v1'

const api: AxiosInstance = axios.create({
  baseURL: API_BASE,
  timeout: DEFAULT_TIMEOUT,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('auth_token')
    if (token && config.headers) config.headers.Authorization = 'Bearer ' + token
    config.headers['X-Request-Time'] = new Date().toISOString()
    config.headers['X-Request-ID'] = crypto.randomUUID()
    return config
  },
  (error: AxiosError) => Promise.reject(error)
)

api.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & RequestConfig
    if (!error.response) {
      if (originalRequest && shouldRetry(error)) return retryRequest(originalRequest)
      return Promise.reject({ message: 'Network error.', code: 'NETWORK_ERROR' })
    }
    const status = error.response.status
    if (status === 401) { localStorage.removeItem('auth_token'); window.location.href = '/login' }
    if (status === 429 && originalRequest && shouldRetry(error)) { await delay(5000); return retryRequest(originalRequest) }
    if ((status === 500 || status === 502 || status === 503) && originalRequest && shouldRetry(error)) return retryRequest(originalRequest)
    return Promise.reject({ message: (error.response.data as { error?: string })?.error || error.message, code: status.toString() })
  }
)

function shouldRetry(error: AxiosError): boolean {
  const config = error.config as InternalAxiosRequestConfig & RequestConfig
  if (!config || config.signal?.aborted) return false
  const retries = config.retries ?? DEFAULT_RETRIES
  const retryCount = (config as unknown as { _retryCount?: number })._retryCount ?? 0
  return retryCount < retries
}

async function retryRequest(config: InternalAxiosRequestConfig & RequestConfig): Promise<AxiosResponse> {
  const retryCount = ((config as unknown as { _retryCount?: number })._retryCount ?? 0) + 1
  ;(config as unknown as { _retryCount?: number })._retryCount = retryCount
  const delayMs = (config.retryDelay ?? DEFAULT_RETRY_DELAY) * Math.pow(2, retryCount - 1)
  await delay(delayMs)
  return api(config)
}

function delay(ms: number): Promise<void> { return new Promise((resolve) => setTimeout(resolve, ms)) }
function handleResponse<T>(response: AxiosResponse): ApiResponse<T> { return { data: response.data as T, error: null, status: response.status } }
function handleError(error: unknown): ApiResponse<never> {
  if (error instanceof Error) {
    return { data: null, error: error.message }
  }
  if (typeof error === 'string') {
    return { data: null, error }
  }
  return { data: null, error: 'An unexpected error occurred' }
}

// Case APIs
export async function getCases(): Promise<ApiResponse<CasesListResponse>> {
  try { const response = await api.get('/cases'); return handleResponse<CasesListResponse>(response) }
  catch (error) { return handleError(error) }
}

export async function getCase(id: string): Promise<ApiResponse<CaseRecord>> {
  try { const response = await api.get('/cases/' + id); return handleResponse<CaseRecord>(response) }
  catch (error) { return handleError(error) }
}

export async function createCase(spec: CaseSpec): Promise<ApiResponse<CaseRecord>> {
  try { const response = await api.post('/cases', spec); return handleResponse<CaseRecord>(response) }
  catch (error) { return handleError(error) }
}

export async function runCase(id: string): Promise<ApiResponse<RunResult>> {
  try { const response = await api.post('/cases/' + id + '/run'); return handleResponse<RunResult>(response) }
  catch (error) { return handleError(error) }
}

export async function getCaseEvents(id: string): Promise<ApiResponse<CaseEventsResponse>> {
  try { const response = await api.get('/cases/' + id + '/events'); return handleResponse<CaseEventsResponse>(response) }
  catch (error) { return handleError(error) }
}

// Approval APIs
export async function getApprovals(caseId?: string): Promise<ApiResponse<Approval[]>> {
  try { const params = caseId ? { case_id: caseId } : {}; const response = await api.get('/approvals', { params }); return handleResponse<Approval[]>(response) }
  catch (error) { return handleError(error) }
}

export async function approveApproval(id: string): Promise<ApiResponse<Approval>> {
  try { const response = await api.post('/approvals/' + id + '/approve'); return handleResponse<Approval>(response) }
  catch (error) { return handleError(error) }
}

export async function rejectApproval(id: string): Promise<ApiResponse<Approval>> {
  try { const response = await api.post('/approvals/' + id + '/reject'); return handleResponse<Approval>(response) }
  catch (error) { return handleError(error) }
}

export async function getReports(): Promise<ApiResponse<ReportSummary[]>> {
  try { const response = await api.get('/reports'); return handleResponse<ReportSummary[]>(response) }
  catch (error) { return handleError(error) }
}

export async function getReport(id: string): Promise<ApiResponse<ReportSummary>> {
  try { const response = await api.get('/reports/' + id); return handleResponse<ReportSummary>(response) }
  catch (error) { return handleError(error) }
}

export async function getReportContent(id: string): Promise<ApiResponse<string>> {
  try {
    const response = await api.get('/reports/' + id + '/content', { responseType: 'text' })
    return handleResponse<string>(response)
  } catch (error) { return handleError(error) }
}

export async function buildReport(caseId: string): Promise<ApiResponse<ReportSummary>> {
  try { const response = await api.post('/reports/' + caseId + '/build'); return handleResponse<ReportSummary>(response) }
  catch (error) { return handleError(error) }
}

export { api }
export default api
