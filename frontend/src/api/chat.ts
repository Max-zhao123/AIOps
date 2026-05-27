import client from './client'
import type { ChatRequest, ChatSession, ChatMessage } from '@/types'

export async function postChat(data: ChatRequest): Promise<{ reply: string; actionPlans?: unknown[] }> {
  const res = await client.post('/chat', { ...data, stream: false })
  return res.data?.data ?? res.data
}

export async function getStreamChatUrl(data: ChatRequest): Promise<string> {
  const params = new URLSearchParams({
    environment: data.environment,
    message: data.message,
  })
  return `${client.defaults.baseURL}/chat/stream?${params}`
}

/** 获取会话列表，支持 status 过滤 */
export async function listSessions(params?: { status?: string }): Promise<ChatSession[]> {
  const res = await client.get('/chat/sessions', { params })
  return res.data?.data ?? res.data ?? []
}

export async function listMessages(sessionId: number): Promise<ChatMessage[]> {
  const res = await client.get(`/chat/sessions/${sessionId}/messages`)
  return res.data?.data ?? res.data ?? []
}

export async function postAssist(data: { error_message: string; context?: string }): Promise<{ reply: string }> {
  const res = await client.post('/chat/assist', data)
  return res.data?.data ?? res.data
}

export async function postRCA(data: { incident: string; context?: string }): Promise<{ reply: string }> {
  const res = await client.post('/chat/rca', data)
  return res.data?.data ?? res.data
}

/** 归档会话 */
export async function archiveSession(sessionId: number): Promise<void> {
  await client.post(`/chat/sessions/${sessionId}/archive`)
}

/** 删除会话 */
export async function deleteSession(sessionId: number): Promise<void> {
  await client.delete(`/chat/sessions/${sessionId}`)
}
