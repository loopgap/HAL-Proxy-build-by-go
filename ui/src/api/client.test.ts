import { beforeEach, describe, expect, it, vi } from 'vitest'

const mockApi = {
  get: vi.fn(),
  post: vi.fn(),
  interceptors: {
    request: { use: vi.fn() },
    response: { use: vi.fn() },
  },
}

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => mockApi),
  },
}))

describe('api client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('uses canonical resource routes', async () => {
    mockApi.post.mockResolvedValue({ data: { ok: true }, status: 200 })
    mockApi.get.mockResolvedValue({ data: 'report body', status: 200 })
    const client = await import('./client')

    await client.runCase('case-1')
    await client.approveApproval('approval-1')
    await client.rejectApproval('approval-2')
    await client.buildReport('case-2')
    await client.getReport('report-1')
    await client.getReportContent('report-1')

    expect(mockApi.post).toHaveBeenNthCalledWith(1, '/cases/case-1/run')
    expect(mockApi.post).toHaveBeenNthCalledWith(2, '/approvals/approval-1/approve')
    expect(mockApi.post).toHaveBeenNthCalledWith(3, '/approvals/approval-2/reject')
    expect(mockApi.post).toHaveBeenNthCalledWith(4, '/reports/case-2/build')
    expect(mockApi.get).toHaveBeenNthCalledWith(1, '/reports/report-1')
    expect(mockApi.get).toHaveBeenNthCalledWith(2, '/reports/report-1/content', { responseType: 'text' })
  })

  it('returns case and event envelopes without flattening in the client layer', async () => {
    mockApi.get
      .mockResolvedValueOnce({
        data: { items: [{ id: 'case-1' }], next_cursor: '', has_more: false },
        status: 200,
      })
      .mockResolvedValueOnce({
        data: { items: [{ sequence: 1, type: 'bridge.case.created' }], total: 1, limit: 100, offset: 0 },
        status: 200,
      })
      .mockResolvedValueOnce({
        data: [{ id: 'report-1', path: '/tmp/report.md' }],
        status: 200,
      })

    const client = await import('./client')

    const cases = await client.getCases()
    const events = await client.getCaseEvents('case-1')
    const reports = await client.getReports()

    expect(cases.data?.items).toHaveLength(1)
    expect(events.data?.items).toHaveLength(1)
    expect(reports.data).toEqual([{ id: 'report-1', path: '/tmp/report.md' }])
  })
})
