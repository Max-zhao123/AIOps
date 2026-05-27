import client from './client'
import type { Runbook } from '@/types'

export async function listRunbooks(): Promise<Runbook[]> {
  const res = await client.get('/runbooks')
  return res.data?.data ?? res.data ?? []
}

export async function createRunbook(data: Partial<Runbook>): Promise<Runbook> {
  const res = await client.post('/runbooks', data)
  return res.data?.data ?? res.data
}

export async function executeRunbook(id: number): Promise<unknown> {
  const res = await client.post(`/runbooks/${id}/execute`)
  return res.data?.data ?? res.data
}
