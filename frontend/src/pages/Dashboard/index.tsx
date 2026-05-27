import { useState, useEffect, useCallback } from 'react'
import { Row, Col, Card, Statistic, Timeline, Table, Tag, Typography, Spin, Empty } from 'antd'
import {
  EnvironmentOutlined,
  SafetyOutlined,
  ClockCircleOutlined,
  AuditOutlined,
} from '@ant-design/icons'
import { listEnvironments } from '@/api/environments'
import { listPolicies } from '@/api/policies'
import { listPending } from '@/api/executor'
import { listAudit } from '@/api/audit'
import { listPlugins } from '@/api/plugins'
import type { AuditLog, Plugin } from '@/types'
import { useNavigate } from 'react-router-dom'

const { Title } = Typography

export default function Dashboard() {
  const [envCount, setEnvCount] = useState(0)
  const [policyCount, setPolicyCount] = useState(0)
  const [pendingCount, setPendingCount] = useState(0)
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [plugins, setPlugins] = useState<Plugin[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [envs, policies, pending, audit, plugs] = await Promise.all([
        listEnvironments(),
        listPolicies(),
        listPending(),
        listAudit({ pageSize: 20 }),
        listPlugins(),
      ])
      setEnvCount(Array.isArray(envs) ? envs.length : 0)
      setPolicyCount(Array.isArray(policies) ? policies.length : 0)
      setPendingCount(Array.isArray(pending) ? pending.length : 0)
      setAuditLogs(Array.isArray(audit) ? audit.slice(0, 20) : [])
      setPlugins(Array.isArray(plugs) ? plugs : [])
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  if (loading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  return (
    <div>
      <Title level={4} style={{ marginBottom: 24 }}>概览仪表盘</Title>

      <Row gutter={[16, 16]}>
        <Col xs={12} sm={6}>
          <Card hoverable onClick={() => navigate('/environments')}>
            <Statistic
              title="环境数"
              value={envCount}
              prefix={<EnvironmentOutlined />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card hoverable onClick={() => navigate('/policies')}>
            <Statistic
              title="安全策略"
              value={policyCount}
              prefix={<SafetyOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card hoverable onClick={() => navigate('/actions')}>
            <Statistic
              title="待确认操作"
              value={pendingCount}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: pendingCount > 0 ? '#fa8c16' : '#999' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card hoverable onClick={() => navigate('/audit')}>
            <Statistic
              title="审计条目"
              value={auditLogs.length}
              prefix={<AuditOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={14}>
          <Card title="最近操作" size="small">
            {auditLogs.length === 0 ? (
              <Empty description="暂无审计记录" />
            ) : (
              <Timeline
                items={auditLogs.map((log) => ({
                  color: log.result === 'success' ? 'green' : log.result === 'denied' ? 'red' : 'blue',
                  children: (
                    <div>
                      <div>
                        <strong>{log.username}</strong>
                        <span style={{ margin: '0 8px', color: '#999' }}>{log.action}</span>
                        <Tag>{log.environment || '-'}</Tag>
                      </div>
                      <div style={{ color: '#999', fontSize: 12 }}>{log.created_at}</div>
                    </div>
                  ),
                }))}
              />
            )}
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title="插件状态" size="small">
            {plugins.length === 0 ? (
              <Empty description="暂无插件" />
            ) : (
              <Table
                size="small"
                pagination={false}
                dataSource={plugins}
                rowKey="name"
                columns={[
                  { title: '名称', dataIndex: 'name', key: 'name' },
                  { title: '类型', dataIndex: 'type', key: 'type', render: (v: string) => <Tag>{v}</Tag> },
                  {
                    title: '状态',
                    dataIndex: 'online',
                    key: 'online',
                    render: (v: boolean) => (
                      <Tag color={v ? 'green' : 'red'}>{v ? '在线' : '离线'}</Tag>
                    ),
                  },
                ]}
              />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  )
}
