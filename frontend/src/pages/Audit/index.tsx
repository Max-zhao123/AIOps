import { useState, useEffect, useCallback } from 'react'
import { Table, Select, Input, App, Typography, Tag, Space } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { listAudit } from '@/api/audit'
import EnvironmentSelector from '@/components/EnvironmentSelector'
import type { AuditLog } from '@/types'

const { Title } = Typography

export default function Audit() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(false)
  const [environment, setEnvironment] = useState<string>()
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listAudit({ environment, page, pageSize })
      setLogs(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [environment, page, pageSize, message])

  useEffect(() => { load() }, [load])

  return (
    <div>
      <Title level={4}>审计日志</Title>

      <Space style={{ marginBottom: 16 }}>
        <EnvironmentSelector
          value={environment}
          onChange={(v) => { setEnvironment(v); setPage(1) }}
          placeholder="全部环境"
          style={{ width: 160 }}
        />
        <Input
          prefix={<SearchOutlined />}
          placeholder="搜索关键词"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          style={{ width: 200 }}
          allowClear
        />
      </Space>

      <Table
        dataSource={logs.filter((l) => !keyword || JSON.stringify(l).includes(keyword))}
        rowKey="id"
        loading={loading}
        pagination={{ current: page, pageSize, onChange: setPage, showTotal: (t) => `共 ${t} 条` }}
        columns={[
          { title: '时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
          { title: '用户', dataIndex: 'username', key: 'username', width: 100 },
          { title: '操作', dataIndex: 'action', key: 'action', width: 120 },
          { title: '资源', dataIndex: 'resource_type', key: 'resource_type', width: 100 },
          { title: '资源ID', dataIndex: 'resource_id', key: 'resource_id', width: 80 },
          {
            title: '环境',
            dataIndex: 'environment',
            key: 'environment',
            width: 120,
            render: (v: string) => v ? <Tag>{v}</Tag> : '-',
          },
          {
            title: '结果',
            dataIndex: 'result',
            key: 'result',
            width: 80,
            render: (v: string) => (
              <Tag color={v === 'success' ? 'green' : v === 'denied' ? 'red' : 'default'}>
                {v}
              </Tag>
            ),
          },
          { title: '详情', dataIndex: 'detail', key: 'detail', ellipsis: true },
        ]}
      />
    </div>
  )
}
