import client from './client'
import type { AuditLog } from '@/types'

export async function listAudit(params?: { environment?: string; page?: number; pageSize?: number }): Promise<AuditLog[]> {
  const res = await client.get('/audit', { params })
  return res.data?.data ?? res.data ?? []
}
