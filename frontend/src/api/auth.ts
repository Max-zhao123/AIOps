import client from './client'
import type { LoginRequest, LoginResponse, ChangePasswordRequest } from '@/types'

interface ApiLoginResponse {
  code: number
  data: {
    token: string
    role: string
    id?: number
    username?: string
    password_expired?: boolean
  }
}

export async function login(req: LoginRequest): Promise<LoginResponse> {
  const res = await client.post<ApiLoginResponse>('/auth/login', req)
  const { token, role, id, username, password_expired } = res.data.data
  return {
    token,
    user: {
      id: id ?? 0,
      username: username ?? req.username,
      role: role as LoginResponse['user']['role'],
      password_expired: password_expired ?? false,
    },
  }
}

/** 修改密码 */
export async function changePassword(userId: number, data: ChangePasswordRequest): Promise<void> {
  await client.put(`/users/${userId}/password`, data)
}
