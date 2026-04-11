import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as api from '@/api/client'
import type { CaseSpec } from '@/types'
import toast from 'react-hot-toast'

// Query keys
export const queryKeys = {
  cases: ['cases'] as const,
  case: (id: string) => ['cases', id] as const,
  caseEvents: (id: string) => ['cases', id, 'events'] as const,
  approvals: (caseId?: string) => ['approvals', caseId] as const,
  reports: ['reports'] as const,
}

// Cases hooks
export function useCases() {
  return useQuery({
    queryKey: queryKeys.cases,
    queryFn: async () => {
      const result = await api.getCases()
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data ?? []
    },
  })
}

export function useCase(id: string) {
  return useQuery({
    queryKey: queryKeys.case(id),
    queryFn: async () => {
      const result = await api.getCase(id)
      if (result.error) {
        throw new Error(result.error)
      }
      if (!result.data) {
        throw new Error('Case not found')
      }
      return result.data
    },
    enabled: !!id,
  })
}

export function useCaseEvents(id: string) {
  return useQuery({
    queryKey: queryKeys.caseEvents(id),
    queryFn: async () => {
      const result = await api.getCaseEvents(id)
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data ?? []
    },
    enabled: !!id,
    refetchInterval: 5000, // Poll every 5 seconds for real-time updates
  })
}

export function useCreateCase() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (spec: CaseSpec) => {
      const result = await api.createCase(spec)
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cases })
      toast.success('Case created successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to create case: ${error.message}`)
    },
  })
}

export function useRunCase() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (caseId: string) => {
      const result = await api.runCase(caseId)
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.case(data!.case.id) })
      queryClient.invalidateQueries({ queryKey: queryKeys.caseEvents(data!.case.id) })
      toast.success(`Case ${data!.status}`)
    },
    onError: (error: Error) => {
      toast.error(`Failed to run case: ${error.message}`)
    },
  })
}

// Approvals hooks
export function useApprovals(caseId?: string) {
  return useQuery({
    queryKey: queryKeys.approvals(caseId),
    queryFn: async () => {
      const result = await api.getApprovals(caseId)
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data ?? []
    },
  })
}

export function useApproveApproval() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (approvalId: string) => {
      const result = await api.approveApproval(approvalId)
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['approvals'] })
      if (data?.case_id) {
        queryClient.invalidateQueries({ queryKey: queryKeys.case(data.case_id) })
      }
      toast.success('Approval granted')
    },
    onError: (error: Error) => {
      toast.error(`Failed to approve: ${error.message}`)
    },
  })
}

export function useRejectApproval() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (approvalId: string) => {
      const result = await api.rejectApproval(approvalId)
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['approvals'] })
      if (data?.case_id) {
        queryClient.invalidateQueries({ queryKey: queryKeys.case(data.case_id) })
      }
      toast.success('Approval rejected')
    },
    onError: (error: Error) => {
      toast.error(`Failed to reject: ${error.message}`)
    },
  })
}

// Reports hooks
export function useBuildReport() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (caseId: string) => {
      const result = await api.buildReport(caseId)
      if (result.error) {
        throw new Error(result.error)
      }
      return result.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.reports })
      toast.success('Report built successfully')
    },
    onError: (error: Error) => {
      toast.error(`Failed to build report: ${error.message}`)
    },
  })
}
