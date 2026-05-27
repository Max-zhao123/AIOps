import client from './client'
import type { SecurityPolicy, SimulateRequest, PolicyEvaluateResult } from '@/types'

export async function listPolicies(): Promise<SecurityPolicy[]> {
  const res = await client.get('/security-boundaries')
  return res.data?.data ?? res.data ?? []
}

export async function getPolicy(id: number): Promise<SecurityPolicy> {
  const res = await client.get(`/security-boundaries/${id}`)
  return res.data?.data ?? res.data
}

export async function createPolicy(data: Partial<SecurityPolicy>): Promise<SecurityPolicy> {
  const res = await client.post('/security-boundaries', data)
  return res.data?.data ?? res.data
}

export async function updatePolicy(id: number, data: Partial<SecurityPolicy>): Promise<SecurityPolicy> {
  const res = await client.put(`/security-boundaries/${id}`, data)
  return res.data?.data ?? res.data
}

export async function enablePolicy(id: number): Promise<void> {
  await client.post(`/security-boundaries/${id}/enable`)
}

export async function disablePolicy(id: number): Promise<void> {
  await client.post(`/security-boundaries/${id}/disable`)
}

export async function simulatePolicy(id: number, data: SimulateRequest): Promise<PolicyEvaluateResult> {
  const res = await client.post(`/security-boundaries/${id}/simulate`, data)
  return res.data?.data ?? res.data
}

export async function importTemplate(): Promise<string> {
  const res = await client.get('/security-boundaries/import-template')
  return res.data?.data ?? res.data
}
