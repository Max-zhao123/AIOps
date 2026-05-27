import { Progress, Tag } from 'antd'

interface PasswordStrengthProps {
  password: string
}

/** 密码强度指示器 */
export default function PasswordStrength({ password }: PasswordStrengthProps) {
  const { score, label, color } = calcStrength(password)

  if (!password) return null

  return (
    <div style={{ marginTop: 4 }}>
      <Progress
        percent={score * 20}
        size="small"
        strokeColor={color}
        showInfo={false}
        style={{ marginBottom: 2 }}
      />
      <Tag color={color} style={{ fontSize: 11 }}>
        {label}
      </Tag>
    </div>
  )
}

function calcStrength(password: string): { score: number; label: string; color: string } {
  if (!password) return { score: 0, label: '', color: '#d9d9d9' }

  let score = 0
  if (password.length >= 8) score++
  if (password.length >= 12) score++
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++
  if (/\d/.test(password)) score++
  if (/[^a-zA-Z0-9]/.test(password)) score++

  if (score <= 1) return { score, label: '弱', color: '#ff4d4f' }
  if (score <= 2) return { score, label: '较弱', color: '#faad14' }
  if (score <= 3) return { score, label: '中等', color: '#fa8c16' }
  if (score <= 4) return { score, label: '强', color: '#52c41a' }
  return { score, label: '非常强', color: '#1890ff' }
}
