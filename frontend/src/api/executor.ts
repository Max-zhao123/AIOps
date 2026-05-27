import client from './client'
import type { ExecutionRecord, ConfirmRequest, RollbackRequest, RollbackResponse } from '@/types'

export async function listPending(): Promise<ExecutionRecord[]> {
  const res = await client.get('/actions/pending')
  return res.data?.items ?? res.data?.data ?? []
}

export async function confirmAction(data: ConfirmRequest): Promise<{ execution_id: number; status: string }> {
  const res = await client.post('/actions/confirm', data)
  return res.data?.data ?? res.data
}

/** 回滚操作 */
export async function rollbackAction(executionId: number, data?: RollbackRequest): Promise<RollbackResponse> {
  const res = await client.post(`/executions/${executionId}/rollback`, data)
  return res.data?.data ?? res.data
}

/** 获取执行详情 */
export async function getExecution(id: number): Promise<ExecutionRecord> {
  const res = await client.get(`/actions/${id}`)
  return res.data?.data ?? res.data
}
