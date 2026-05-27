import client from './client'

export async function helpdeskChat(query: string): Promise<{ reply: string }> {
  const res = await client.post('/helpdesk/chat', { query })
  return res.data?.data ?? res.data
}

export async function imWebhook(text: string): Promise<void> {
  await client.post('/im/webhook', { text })
}
