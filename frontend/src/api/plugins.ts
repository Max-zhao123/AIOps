import client from './client'
import type { Plugin, AuditLog } from '@/types'

export async function listPlugins(): Promise<Plugin[]> {
  const res = await client.get('/plugins')
  return res.data?.data ?? res.data ?? []
}

export async function listAudit(params?: { environment?: string; page?: number; pageSize?: number }): Promise<AuditLog[]> {
  const res = await client.get('/audit', { params })
  return res.data?.data ?? res.data ?? []
}
