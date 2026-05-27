import { useState, useEffect, useCallback } from 'react'
import { Table, Button, App, Typography, Tag, Space, Popconfirm } from 'antd'
import { PlayCircleOutlined } from '@ant-design/icons'
import { listPending, confirmAction } from '@/api/executor'
import type { ExecutionRecord } from '@/types'
import { useAuthStore } from '@/stores/authStore'
import ApprovalStatus from '@/components/ApprovalStatus'
import RollbackButton from '@/components/RollbackButton'

const { Title } = Typography

/** F-043 + F-044: 审批流改造 + 回滚操作 */
export default function Actions() {
  const [records, setRecords] = useState<ExecutionRecord[]>([])
  const [loading, setLoading] = useState(false)
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
          { title: '时间', dataIndex: 'createdAt', key: 'time', width: 180 },
          { title: '用户ID', dataIndex: 'userId', key: 'user', width: 80 },
          {
            title: '操作摘要',
            key: 'summary',
            ellipsis: true,
            render: (_: unknown, r: ExecutionRecord) => `${r.plugin} / ${r.action}`,
          },
          {
            title: '命令预览',
            key: 'cmd',
            width: 300,
            render: (_: unknown, r: ExecutionRecord) => (
              <code style={{ fontSize: 12, background: '#f5f5f5', padding: '2px 6px', borderRadius: 4 }}>
                {JSON.stringify(JSON.parse(r.planJson || '{}'), null, 2).substring(0, 200)}
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
            title: '审批状态',
            key: 'approval_status',
            width: 180,
            render: (_: unknown, r: ExecutionRecord) => {
              if (!r.approvers || r.approvers.length === 0) return '-'
              return (
                <ApprovalStatus
                  approvers={r.approvers}
                  timeoutAt={r.approval_timeout_at}
                />
              )
            },
          },
          {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            width: 100,
            render: (v: string) => {
              const statusMap: Record<string, { color: string; label: string }> = {
                pending: { color: 'default', label: '待审批' },
                running: { color: 'processing', label: '执行中' },
                succeeded: { color: 'success', label: '成功' },
                failed: { color: 'error', label: '失败' },
                rejected: { color: 'warning', label: '已拒绝' },
                rolled_back: { color: 'purple', label: '已回滚' },
                rolling_back: { color: 'purple', label: '回滚中' },
              }
              const cfg = statusMap[v] || { color: 'default', label: v }
              return <Tag color={cfg.color}>{cfg.label}</Tag>
            },
          },
          {
            title: '操作',
            key: 'actions',
            width: 260,
            render: (_: unknown, r: ExecutionRecord) => {
              if (readonly) return <Tag>无权限</Tag>
              return (
                <Space wrap>
                  {(r.status === 'pending' || !r.status) && (
                    <>
                      <Popconfirm title="确认执行？" onConfirm={() => handleConfirm(r.id, true)}>
                        <Button type="primary" size="small">确认执行</Button>
                      </Popconfirm>
                      <Popconfirm title="确认拒绝？" onConfirm={() => handleConfirm(r.id, false)}>
                        <Button danger size="small">拒绝</Button>
                      </Popconfirm>
                    </>
                  )}
                  <RollbackButton record={r} onSuccess={load} />
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
