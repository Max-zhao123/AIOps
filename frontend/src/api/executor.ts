import client from './client'
import type { ExecutionRecord, ConfirmRequest } from '@/types'

export async function listPending(): Promise<ExecutionRecord[]> {
  const res = await client.get('/actions/pending')
  return res.data?.items ?? res.data?.data ?? []
}

export async function confirmAction(data: ConfirmRequest): Promise<void> {
  await client.post('/actions/confirm', data)
}
