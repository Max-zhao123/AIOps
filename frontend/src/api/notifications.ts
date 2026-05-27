import client from './client'
import type {
  NotificationChannel,
  NotificationPolicy,
  NotificationTemplate,
  TestNotificationRequest,
} from '@/types'

// ============ 通知通道 ============

/** 获取通知通道列表 */
export async function listChannels(): Promise<NotificationChannel[]> {
  const res = await client.get('/notification-channels')
  return res.data?.data ?? res.data ?? []
}

/** 创建通知通道 */
export async function createChannel(data: Partial<NotificationChannel>): Promise<NotificationChannel> {
  const res = await client.post('/notification-channels', data)
  return res.data?.data ?? res.data
}

/** 更新通知通道 */
export async function updateChannel(id: number, data: Partial<NotificationChannel>): Promise<NotificationChannel> {
  const res = await client.put(`/notification-channels/${id}`, data)
  return res.data?.data ?? res.data
}

/** 删除通知通道 */
export async function deleteChannel(id: number): Promise<void> {
  await client.delete(`/notification-channels/${id}`)
}

/** 测试发送通知 */
export async function testNotification(id: number, data?: TestNotificationRequest): Promise<{ success: boolean; message: string }> {
  const res = await client.post(`/notification-channels/${id}/test`, data)
  return res.data?.data ?? res.data
}

// ============ 通知策略 ============

/** 获取通知策略列表 */
export async function listPolicies(): Promise<NotificationPolicy[]> {
  const res = await client.get('/notification-policies')
  return res.data?.data ?? res.data ?? []
}

/** 创建通知策略 */
export async function createPolicy(data: Partial<NotificationPolicy>): Promise<NotificationPolicy> {
  const res = await client.post('/notification-policies', data)
  return res.data?.data ?? res.data
}

/** 更新通知策略 */
export async function updatePolicy(id: number, data: Partial<NotificationPolicy>): Promise<NotificationPolicy> {
  const res = await client.put(`/notification-policies/${id}`, data)
  return res.data?.data ?? res.data
}

/** 删除通知策略 */
export async function deletePolicy(id: number): Promise<void> {
  await client.delete(`/notification-policies/${id}`)
}

// ============ 通知模板 ============

/** 获取通知模板列表 */
export async function listTemplates(): Promise<NotificationTemplate[]> {
  const res = await client.get('/notification-templates')
  return res.data?.data ?? res.data ?? []
}

/** 创建通知模板 */
export async function createTemplate(data: Partial<NotificationTemplate>): Promise<NotificationTemplate> {
  const res = await client.post('/notification-templates', data)
  return res.data?.data ?? res.data
}

/** 更新通知模板 */
export async function updateTemplate(id: number, data: Partial<NotificationTemplate>): Promise<NotificationTemplate> {
  const res = await client.put(`/notification-templates/${id}`, data)
  return res.data?.data ?? res.data
}

/** 删除通知模板 */
export async function deleteTemplate(id: number): Promise<void> {
  await client.delete(`/notification-templates/${id}`)
}
