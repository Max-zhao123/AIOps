import { useState, useEffect, useCallback } from 'react'
import { Table, Button, App, Typography, Tag, Space, Empty } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { listReports } from '@/api/inspections'
import type { InspectionReport } from '@/types'

const { Title } = Typography

export default function Reports() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [reports, setReports] = useState<InspectionReport[]>([])
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()

  const load = useCallback(async () => {
    if (!id) return
    setLoading(true)
    try {
      const data = await listReports(Number(id))
      setReports(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [id, message])

  useEffect(() => { load() }, [load])

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/inspections')}>返回</Button>
        <Title level={4} style={{ margin: 0 }}>巡检报告 #{id}</Title>
      </Space>

      {reports.length === 0 && !loading ? (
        <Empty description="暂无巡检报告" />
      ) : (
        <Table
          dataSource={reports}
          rowKey="id"
          loading={loading}
          columns={[
            { title: '执行时间', dataIndex: 'executed_at', key: 'time', width: 180 },
            { title: '步骤', dataIndex: 'step_index', key: 'step', width: 80, render: (v: number) => `步骤 ${v + 1}` },
            {
              title: '状态',
              dataIndex: 'status',
              key: 'status',
              width: 100,
              render: (v: string) => (
                <Tag color={v === 'success' ? 'green' : 'red'}>{v === 'success' ? '成功' : '失败'}</Tag>
              ),
            },
            { title: '耗时', dataIndex: 'duration_ms', key: 'duration', width: 100, render: (v: number) => `${v}ms` },
            {
              title: '结果/错误',
              key: 'detail',
              ellipsis: true,
              render: (_: unknown, r: InspectionReport) => (
                r.status === 'failed'
                  ? <span style={{ color: '#ff4d4f' }}>{r.error || r.result || '未知错误'}</span>
                  : <span>{r.result || '-'}</span>
              ),
            },
          ]}
        />
      )}
    </div>
  )
}
