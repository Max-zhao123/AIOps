import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { isTokenExpired } from '@/utils/jwt'

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const { token, isAuthenticated, logout } = useAuthStore()

  if (!isAuthenticated || !token) {
    return <Navigate to="/login" replace />
  }

  if (isTokenExpired(token)) {
    logout()
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}
