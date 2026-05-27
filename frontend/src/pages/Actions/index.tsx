import { useState, useEffect, useCallback } from 'react'
import { Table, Button, App, Typography, Tag, Space, Popconfirm } from 'antd'
import { PlayCircleOutlined, EyeOutlined } from '@ant-design/icons'
import { listPending, confirmAction } from '@/api/executor'
import type { ExecutionRecord } from '@/types'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

const { Title } = Typography

export default function Actions() {
  const [records, setRecords] = useState<ExecutionRecord[]>([])
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { message } = App.useApp()
  const hasRole = useAuthStore((s) => s.hasRole)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listPending()
      setRecords(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => {
    load()
    const timer = setInterval(load, 10000)
    return () => clearInterval(timer)
  }, [load])

  const handleConfirm = async (id: number, approved: boolean) => {
    try {
      await confirmAction({ executionId: id, approved })
      message.success(approved ? '已确认' : '已拒绝')
      load()
    } catch {
      message.error('操作失败')
    }
  }

  const readonly = !hasRole('admin') && !hasRole('operator')

  return (
    <div>
      <Title level={4}>待确认操作</Title>

      <Table
        dataSource={records}
        rowKey="id"
        loading={loading}
        columns={[
          { title: '时间', dataIndex: 'created_at', key: 'time', width: 180 },
          { title: '用户', dataIndex: 'username', key: 'user', width: 100 },
          {
            title: '操作摘要',
            key: 'summary',
            ellipsis: true,
            render: (_: unknown, r: ExecutionRecord) => {
              const plan = r.action_plan
              if (!plan) return '-'
              return `${plan.plugin} / ${plan.action}${plan.summary ? ` - ${plan.summary}` : ''}`
            },
          },
          {
            title: '命令预览',
            key: 'cmd',
            width: 300,
            render: (_: unknown, r: ExecutionRecord) => (
              <code style={{ fontSize: 12, background: '#f5f5f5', padding: '2px 6px', borderRadius: 4 }}>
                {r.action_plan?.commandPreview || '-'}
              </code>
            ),
          },
          {
            title: '环境',
            dataIndex: 'environment',
            key: 'env',
            width: 120,
            render: (v: string) => <Tag>{v}</Tag>,
          },
          {
            title: '决策',
            dataIndex: 'decision',
            key: 'decision',
            width: 80,
            render: (v: string) => (
              <Tag color={v === 'ALLOW' ? 'green' : v === 'ASK' ? 'orange' : 'red'}>{v}</Tag>
            ),
          },
          {
            title: '操作',
            key: 'actions',
            width: 200,
            render: (_: unknown, r: ExecutionRecord) => {
              if (readonly) return <Tag>无权限</Tag>
              return (
                <Space>
                  <Popconfirm title="确认执行？" onConfirm={() => handleConfirm(r.id, true)}>
                    <Button type="primary" size="small">确认执行</Button>
                  </Popconfirm>
                  <Popconfirm title="确认拒绝？" onConfirm={() => handleConfirm(r.id, false)}>
                    <Button danger size="small">拒绝</Button>
                  </Popconfirm>
                </Space>
              )
            },
          },
        ]}
        locale={{ emptyText: '暂无待确认操作' }}
      />
    </div>
  )
}
