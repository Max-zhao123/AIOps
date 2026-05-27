import client from './client'
import type { KBDocument } from '@/types'

export async function listDocuments(): Promise<KBDocument[]> {
  const res = await client.get('/kb/documents')
  return res.data?.data ?? res.data ?? []
}

export async function getDocument(id: number): Promise<KBDocument> {
  const res = await client.get(`/kb/documents/${id}`)
  return res.data?.data ?? res.data
}

export async function createDocument(data: Partial<KBDocument>): Promise<KBDocument> {
  const res = await client.post('/kb/documents', data)
  return res.data?.data ?? res.data
}

export async function deleteDocument(id: number): Promise<void> {
  await client.delete(`/kb/documents/${id}`)
}
