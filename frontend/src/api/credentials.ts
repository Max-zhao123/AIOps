import client from './client'
import type { Credential, CreateCredentialRequest, RotateCredentialRequest } from '@/types'

/** 获取凭证列表 */
export async function listCredentials(): Promise<Credential[]> {
  const res = await client.get('/credentials')
  return res.data?.data ?? res.data ?? []
}

/** 创建凭证 */
export async function createCredential(data: CreateCredentialRequest): Promise<Credential> {
  const res = await client.post('/credentials', data)
  return res.data?.data ?? res.data
}

/** 轮转凭证 */
export async function rotateCredential(id: number, data: RotateCredentialRequest): Promise<Credential> {
  const res = await client.post(`/credentials/${id}/rotate`, data)
  return res.data?.data ?? res.data
}

/** 删除凭证 */
export async function deleteCredential(id: number): Promise<void> {
  await client.delete(`/credentials/${id}`)
}
