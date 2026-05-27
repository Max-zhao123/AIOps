import client from './client'
import type { SLADefinition, SLAStats, CreateSLARequest } from '@/types'

/** 获取 SLA 定义列表 */
export async function listSLADefinitions(): Promise<SLADefinition[]> {
  const res = await client.get('/sla-definitions')
  return res.data?.data ?? res.data ?? []
}

/** 创建 SLA 定义 */
export async function createSLADefinition(data: CreateSLARequest): Promise<SLADefinition> {
  const res = await client.post('/sla-definitions', data)
  return res.data?.data ?? res.data
}

/** 更新 SLA 定义 */
export async function updateSLADefinition(id: number, data: Partial<CreateSLARequest>): Promise<SLADefinition> {
  const res = await client.put(`/sla-definitions/${id}`, data)
  return res.data?.data ?? res.data
}

/** 删除 SLA 定义 */
export async function deleteSLADefinition(id: number): Promise<void> {
  await client.delete(`/sla-definitions/${id}`)
}

/** 获取 SLA 统计数据 */
export async function getSLAStats(params?: {
  period?: string
  environment?: string
}): Promise<SLAStats> {
  const res = await client.get('/sla/stats', { params })
  return res.data?.data ?? res.data
}
