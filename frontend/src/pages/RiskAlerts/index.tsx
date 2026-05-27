import { useState, useEffect, useCallback } from 'react'
import { List, Card, Button, App, Typography, Tag, Space, Empty, Spin } from 'antd'
import { ScanOutlined, WarningOutlined } from '@ant-design/icons'
import { listRiskAlerts, scanRiskAlerts } from '@/api/riskAlerts'
import type { RiskAlert } from '@/types'

const { Title, Text } = Typography

const LEVEL_CONFIG: Record<string, { color: string; label: string }> = {
  critical: { color: '#ff4d4f', label: '严重' },
  high: { color: '#fa8c16', label: '高危' },
  medium: { color: '#fadb14', label: '中等' },
  low: { color: '#1677ff', label: '低危' },
}

export default function RiskAlerts() {
  const [alerts, setAlerts] = useState<RiskAlert[]>([])
  const [loading, setLoading] = useState(false)
  const [scanning, setScanning] = useState(false)
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listRiskAlerts()
      setAlerts(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const handleScan = async () => {
    setScanning(true)
    try {
      await scanRiskAlerts()
      message.success('扫描完成')
      load()
    } catch {
      message.error('扫描失败')
    } finally {
      setScanning(false)
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>高危预警</Title>
        <Button
          type="primary"
          icon={<ScanOutlined />}
          onClick={handleScan}
          loading={scanning}
          danger
        >
          立即扫描
        </Button>
      </div>

      {loading ? (
        <Spin style={{ display: 'block', margin: '60px auto' }} />
      ) : alerts.length === 0 ? (
        <Empty description="暂无预警" />
      ) : (
        <List
          dataSource={alerts}
          renderItem={(alert) => {
            const cfg = LEVEL_CONFIG[alert.level] || LEVEL_CONFIG.low
            return (
              <Card
                size="small"
                style={{ marginBottom: 12, borderLeft: `4px solid ${cfg.color}` }}
              >
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  <Space>
                    <WarningOutlined style={{ color: cfg.color }} />
                    <Text strong>{alert.title}</Text>
                    <Tag color={cfg.color}>{cfg.label}</Tag>
                    <Tag>{alert.source}</Tag>
                  </Space>
                  <Text type="secondary">{alert.description}</Text>
                  <Text type="secondary" style={{ fontSize: 11 }}>
                    {alert.created_at}
                  </Text>
                </Space>
              </Card>
            )
          }}
        />
      )}
    </div>
  )
}
