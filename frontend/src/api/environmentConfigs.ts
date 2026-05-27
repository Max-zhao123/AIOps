import client from './client'
import type { EnvironmentConfig, UpsertEnvironmentConfigRequest } from '@/types'

/** 获取环境配置列表 */
export async function listEnvironmentConfigs(environmentSlug: string): Promise<EnvironmentConfig[]> {
  const res = await client.get(`/environments/${environmentSlug}/configs`)
  return res.data?.data ?? res.data ?? []
}

/** 新增/更新环境配置 */
export async function upsertEnvironmentConfig(
  environmentSlug: string,
  data: UpsertEnvironmentConfigRequest,
): Promise<EnvironmentConfig> {
  const res = await client.post(`/environments/${environmentSlug}/config`, data)
  return res.data?.data ?? res.data
}

/** 删除环境配置 */
export async function deleteEnvironmentConfig(
  environmentSlug: string,
  key: string,
): Promise<void> {
  await client.delete(`/environments/${environmentSlug}/config/${encodeURIComponent(key)}`)
}
