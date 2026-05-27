import { Result, Button } from 'antd'
import { useAuthStore } from '@/stores/authStore'

interface Props {
  requiredRole?: string
  children: React.ReactNode
}

export default function RoleGuard({ requiredRole, children }: Props) {
  const { hasRole } = useAuthStore()

  if (requiredRole && !hasRole(requiredRole)) {
    return (
      <Result
        status="403"
        title="403"
        subTitle="抱歉，您没有权限访问此页面。"
        extra={
          <Button type="primary" onClick={() => window.location.href = '/dashboard'}>
            返回首页
          </Button>
        }
      />
    )
  }

  return <>{children}</>
}
