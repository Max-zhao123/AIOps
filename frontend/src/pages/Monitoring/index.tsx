import { useState } from 'react'
import { Card, Select, Typography, App, Input, Space, Button, Tooltip } from 'antd'
import { LinkOutlined, ReloadOutlined, ExpandOutlined } from '@ant-design/icons'
import EnvironmentSelector from '@/components/EnvironmentSelector'

const { Title, Text } = Typography

/** F-050: 系统监控页 — iframe 嵌入 Grafana 或 ECharts 自建面板 */
export default function Monitoring() {
  const [mode, setMode] = useState<'grafana' | 'builtin'>('builtin')
  const [environment, setEnvironment] = useState<string>()
  const [grafanaUrl, setGrafanaUrl] = useState('http://localhost:3000')
  const { message } = App.useApp()

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>系统监控</Title>
        <Space>
          <EnvironmentSelector
            value={environment}
            onChange={setEnvironment}
            placeholder="全部环境"
            style={{ width: 160 }}
          />
          <Select
            value={mode}
            onChange={setMode}
            style={{ width: 140 }}
            options={[
              { label: '内置面板', value: 'builtin' },
              { label: 'Grafana', value: 'grafana' },
            ]}
          />
        </Space>
      </div>

      {mode === 'grafana' ? (
        <Card>
          <Space style={{ marginBottom: 12 }}>
            <Text>Grafana 地址：</Text>
            <Input
              value={grafanaUrl}
              onChange={(e) => setGrafanaUrl(e.target.value)}
              style={{ width: 360 }}
              placeholder="http://localhost:3000"
            />
            <Tooltip title="在新标签页中打开">
              <Button
                icon={<ExpandOutlined />}
                onClick={() => window.open(grafanaUrl, '_blank')}
              />
            </Tooltip>
          </Space>
          <div style={{ border: '1px solid #e8e8e8', borderRadius: 8, overflow: 'hidden' }}>
            <iframe
              src={grafanaUrl}
              title="Grafana Dashboard"
              style={{ width: '100%', height: 'calc(100vh - 280px)', minHeight: 500, border: 'none' }}
              sandbox="allow-scripts allow-same-origin allow-popups"
            />
          </div>
        </Card>
      ) : (
        <BuiltInDashboard environment={environment} />
      )}
    </div>
  )
}

/** 内置简易监控面板 */
function BuiltInDashboard({ environment }: { environment?: string }) {
  // 使用简单占位卡片，真实场景中对接 /api/v1/monitoring/metrics
  const metrics = [
    { title: 'API 请求量', value: '-', unit: 'req/min', color: '#1890ff' },
    { title: '平均响应时间', value: '-', unit: 'ms', color: '#52c41a' },
    { title: '错误率', value: '-', unit: '%', color: '#ff4d4f' },
    { title: '活跃会话', value: '-', unit: '', color: '#722ed1' },
    { title: 'CPU 使用率', value: '-', unit: '%', color: '#fa8c16' },
    { title: '内存使用率', value: '-', unit: '%', color: '#13c2c2' },
  ]

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
      {metrics.map((m) => (
        <Card key={m.title} size="small">
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 13, color: '#666', marginBottom: 8 }}>{m.title}</div>
            <div style={{ fontSize: 28, fontWeight: 700, color: m.color }}>
              {m.value}
              <span style={{ fontSize: 13, fontWeight: 400, color: '#999' }}> {m.unit}</span>
            </div>
            <div style={{ fontSize: 11, color: '#bbb', marginTop: 8 }}>
              连接后端监控接口后显示实时数据
              {environment && ` · 环境: ${environment}`}
            </div>
          </div>
        </Card>
      ))}

      <Card
        style={{ gridColumn: '1 / -1' }}
        title="系统指标面板"
        extra={<Button size="small" icon={<ReloadOutlined />}>刷新</Button>}
      >
        <div style={{
          textAlign: 'center',
          padding: '60px 0',
          color: '#999',
          background: '#fafafa',
          borderRadius: 8,
        }}>
          <LinkOutlined style={{ fontSize: 36, marginBottom: 12, color: '#d9d9d9' }} />
          <div>接入 Prometheus/Grafana 后将在此显示实时指标图表</div>
          <div style={{ fontSize: 12, marginTop: 4 }}>
            切换到 Grafana 模式可直接嵌入现有仪表盘
          </div>
        </div>
      </Card>
    </div>
  )
}
