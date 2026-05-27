import { Select } from 'antd'
import { useEffect, useState } from 'react'
import { listEnvironments } from '@/api/environments'
import type { Environment } from '@/types'

interface Props {
  value?: string
  onChange?: (slug: string) => void
  style?: React.CSSProperties
  placeholder?: string
}

export default function EnvironmentSelector({ value, onChange, style, placeholder }: Props) {
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
      value={value}
      onChange={onChange}
      loading={loading}
      placeholder={placeholder || '选择环境'}
      style={{ minWidth: 160, ...style }}
      options={envs.map((e) => ({ label: e.name, value: e.slug }))}
    />
  )
}
