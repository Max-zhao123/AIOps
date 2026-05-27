import client from './client'
import type { Schedule, ScheduleExecution, CreateScheduleRequest } from '@/types'

/** 获取调度列表 */
export async function listSchedules(params?: { enabled?: boolean }): Promise<Schedule[]> {
  const res = await client.get('/schedules', { params })
  return res.data?.data ?? res.data ?? []
}

/** 获取单个调度 */
export async function getSchedule(id: number): Promise<Schedule> {
  const res = await client.get(`/schedules/${id}`)
  return res.data?.data ?? res.data
}

/** 创建调度 */
export async function createSchedule(data: CreateScheduleRequest): Promise<Schedule> {
  const res = await client.post('/schedules', data)
  return res.data?.data ?? res.data
}

/** 更新调度 */
export async function updateSchedule(id: number, data: Partial<CreateScheduleRequest>): Promise<Schedule> {
  const res = await client.put(`/schedules/${id}`, data)
  return res.data?.data ?? res.data
}

/** 删除调度 */
export async function deleteSchedule(id: number): Promise<void> {
  await client.delete(`/schedules/${id}`)
}

/** 切换调度启停 */
export async function toggleSchedule(id: number, enabled: boolean): Promise<void> {
  await client.patch(`/schedules/${id}/toggle`, { enabled })
}

/** 手动触发调度 */
export async function triggerSchedule(id: number): Promise<void> {
  await client.post(`/schedules/${id}/trigger`)
}

/** 获取调度执行历史 */
export async function listScheduleExecutions(
  scheduleId: number,
  params?: { page?: number; pageSize?: number },
): Promise<ScheduleExecution[]> {
  const res = await client.get(`/schedules/${scheduleId}/executions`, { params })
  return res.data?.data ?? res.data ?? []
}
