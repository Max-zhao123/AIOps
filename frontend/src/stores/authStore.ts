import { create } from 'zustand'
import type { User } from '@/types'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  login: (token: string, user: User) => void
  logout: () => void
  hasRole: (role: string) => boolean
}

function getSafeItem(key: string): string | null {
  try {
    const v = localStorage.getItem(key)
    if (!v || v === 'undefined' || v === 'null') return null
    return v
  } catch {
    return null
  }
}

function getSafeUser(): User | null {
  try {
    const raw = getSafeItem('user')
    if (!raw) return null
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: getSafeItem('token'),
  user: getSafeUser(),
  isAuthenticated: !!(getSafeItem('token')),

  login: (token: string, user: User) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    set({ token, user, isAuthenticated: true })
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ token: null, user: null, isAuthenticated: false })
  },

  hasRole: (role: string) => {
    const user = get().user
    if (!user) return false
    if (user.role === 'admin') return true
    return user.role === role
  },
}))

// 清理上次登录 bug 残留的坏数据
if (localStorage.getItem('token') === 'undefined') localStorage.removeItem('token')
if (localStorage.getItem('user') === 'undefined') localStorage.removeItem('user')
