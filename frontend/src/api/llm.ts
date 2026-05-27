import client from './client'
import type { LlmConfig } from '@/types'

export async function listLlmConfigs(): Promise<LlmConfig[]> {
  const res = await client.get('/llm/config')
  return res.data?.data ?? []
}

export async function createLlmConfig(data: Partial<LlmConfig>): Promise<LlmConfig> {
  const res = await client.post('/llm/config', data)
  return res.data?.data ?? res.data
}

export async function updateLlmConfig(id: number, data: Partial<LlmConfig>): Promise<LlmConfig> {
  const res = await client.put(`/llm/config/${id}`, data)
  return res.data?.data ?? res.data
}

export async function deleteLlmConfig(id: number): Promise<void> {
  await client.delete(`/llm/config/${id}`)
}
