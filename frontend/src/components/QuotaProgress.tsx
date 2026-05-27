import { Progress, Tooltip, Space, Typography } from 'antd'

const { Text } = Typography

interface QuotaProgressProps {
  label: string
  current: number
  limit: number
  resetAt?: string
  showDetail?: boolean
}

/** 配额使用量进度条 */
export default function QuotaProgress({
  label,
  current,
  limit,
  resetAt,
  showDetail = true,
}: QuotaProgressProps) {
  const percent = limit > 0 ? Math.round((current / limit) * 100) : 0

  const getColor = (pct: number): string => {
    if (pct >= 90) return '#ff4d4f'
    if (pct >= 70) return '#faad14'
    return '#1890ff'
  }

  const format = () => `${current} / ${limit}`

  return (
    <div style={{ marginBottom: 8 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 2 }}>
        <Text style={{ fontSize: 13 }}>{label}</Text>
        {showDetail && (
          <Text type="secondary" style={{ fontSize: 12 }}>
            {current} / {limit}
          </Text>
        )}
      </div>
      <Tooltip title={`已用 ${current}，限额 ${limit}${resetAt ? `，重置于 ${new Date(resetAt).toLocaleString()}` : ''}`}>
        <Progress
          percent={Math.min(percent, 100)}
          size="small"
          strokeColor={getColor(percent)}
          format={format}
        />
      </Tooltip>
    </div>
  )
}
