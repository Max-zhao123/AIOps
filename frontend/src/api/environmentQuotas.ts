import client from './client'
import type { EnvironmentQuota } from '@/types'

/** 获取环境配额 */
export async function getEnvironmentQuota(environmentSlug: string): Promise<EnvironmentQuota> {
  const res = await client.get(`/environments/${environmentSlug}/quotas`)
  return res.data?.data ?? res.data
}

/** 更新环境配额 */
export async function updateEnvironmentQuota(
  environmentSlug: string,
  data: Record<string, { daily_limit: number }>,
): Promise<EnvironmentQuota> {
  const res = await client.put(`/environments/${environmentSlug}/quotas`, data)
  return res.data?.data ?? res.data
}
