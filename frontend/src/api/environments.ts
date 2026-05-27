import client from './client'
import type { Environment } from '@/types'

export async function listEnvironments(): Promise<Environment[]> {
  const res = await client.get('/environments')
  return res.data?.data ?? res.data ?? []
}

export async function createEnvironment(data: Partial<Environment>): Promise<Environment> {
  const res = await client.post('/environments', data)
  return res.data?.data ?? res.data
}

export async function updateEnvironment(id: number, data: Partial<Environment>): Promise<Environment> {
  const res = await client.put(`/environments/${id}`, data)
  return res.data?.data ?? res.data
}

export async function deleteEnvironment(id: number): Promise<void> {
  await client.delete(`/environments/${id}`)
}
