import client from './client'
import type { AuditLog, ComplianceReport } from '@/types'

export async function listAudit(params?: { environment?: string; page?: number; pageSize?: number }): Promise<AuditLog[]> {
  const res = await client.get('/audit', { params })
  return res.data?.data ?? res.data ?? []
}

/** 导出审计日志（返回 blob） */
export async function exportAudit(params: {
  format: 'csv' | 'pdf'
  environment?: string
  start_date?: string
  end_date?: string
}): Promise<Blob> {
  const res = await client.get('/audit/export', {
    params,
    responseType: 'blob',
  })
  return res.data
}

/** 获取合规报告 */
export async function getComplianceReport(params?: {
  period?: string
  environment?: string
}): Promise<ComplianceReport> {
  const res = await client.get('/audit/report', { params })
  return res.data?.data ?? res.data
}
