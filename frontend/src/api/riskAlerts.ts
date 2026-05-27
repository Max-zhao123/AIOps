import client from './client'
import type { RiskAlert } from '@/types'

export async function listRiskAlerts(): Promise<RiskAlert[]> {
  const res = await client.get('/risk-alerts')
  return res.data?.data ?? res.data ?? []
}

export async function scanRiskAlerts(): Promise<void> {
  await client.post('/risk-alerts/scan')
}
