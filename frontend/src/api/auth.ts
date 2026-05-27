import client from './client'
import type { LoginRequest, LoginResponse } from '@/types'

interface ApiLoginResponse {
  code: number
  data: {
    token: string
    role: string
    id?: number
    username?: string
  }
}

export async function login(req: LoginRequest): Promise<LoginResponse> {
  const res = await client.post<ApiLoginResponse>('/auth/login', req)
  const { token, role, id, username } = res.data.data
  return {
    token,
    user: {
      id: id ?? 0,
      username: username ?? req.username,
      role: role as LoginResponse['user']['role'],
    },
  }
}
