import { Select } from 'antd'
import { useEffect, useState } from 'react'
import { listEnvironments } from '@/api/environments'
import type { Environment } from '@/types'

interface Props {
  value?: string
  onChange?: (slug: string) => void
  style?: React.CSSProperties
  placeholder?: string
  variant?: 'outlined' | 'borderless' | 'filled'
  size?: 'small' | 'middle' | 'large'
  id?: string
}

export default function EnvironmentSelector({
  value, onChange, style, placeholder, variant, size, id,
}: Props) {
  const [envs, setEnvs] = useState<Environment[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    listEnvironments()
      .then(setEnvs)
      .finally(() => setLoading(false))
  }, [])

  return (
    <Select
      id={id}
      value={value || undefined}
      onChange={onChange}
      loading={loading}
      placeholder={placeholder || '选择环境'}
      variant={variant}
      size={size}
      style={{ minWidth: 160, ...style }}
      options={envs.map((e) => ({ label: e.name, value: e.slug }))}
    />
  )
}
