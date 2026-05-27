import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, App, Space, Typography, Tag, Popconfirm, Empty } from 'antd'
import { PlusOutlined, PlayCircleOutlined, FileTextOutlined, DeleteOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import EnvironmentSelector from '@/components/EnvironmentSelector'
import { listInspections, createInspection, runInspection } from '@/api/inspections'
import type { Inspection, InspectionStep } from '@/types'

const { Title, Text } = Typography

const DEFAULT_STEPS: InspectionStep[] = [{ plugin: 'mock', action: 'echo', parameters: { message: 'step1' } }]

export default function Inspections() {
  const [inspections, setInspections] = useState<Inspection[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [stepsText, setStepsText] = useState(JSON.stringify(DEFAULT_STEPS, null, 2))
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listInspections()
      setInspections(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const handleCreate = async () => {
    const name = form.getFieldValue('name')
    const env = form.getFieldValue('environment')
    if (!name || !env) { message.warning('请填写名称和环境'); return }

    try {
      const steps = JSON.parse(stepsText)
      await createInspection({ name, environment: env, steps })
      message.success('创建成功')
      setModalOpen(false)
      form.resetFields()
      setStepsText(JSON.stringify(DEFAULT_STEPS, null, 2))
      load()
    } catch (e) {
      if (e instanceof SyntaxError) message.error('步骤 JSON 格式错误')
      else message.error('创建失败')
    }
  }

  const handleRun = async (id: number) => {
    try {
      await runInspection(id)
      message.success('巡检已启动')
      load()
    } catch {
      message.error('执行失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>巡检任务</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          新建巡检
        </Button>
      </div>

      {inspections.length === 0 && !loading ? (
        <Empty description="暂无巡检任务" />
      ) : (
        <Table
          dataSource={inspections}
          rowKey="id"
          loading={loading}
          columns={[
            { title: '名称', dataIndex: 'name', key: 'name' },
            {
              title: '环境',
              dataIndex: 'environment',
              key: 'env',
              width: 120,
              render: (v: string) => <Tag>{v}</Tag>,
            },
            { title: '步骤数', dataIndex: 'steps', key: 'steps', width: 80, render: (s: InspectionStep[]) => s?.length || 0 },
            { title: '最近执行', dataIndex: 'last_run_at', key: 'lastRun', width: 180 },
            {
              title: '状态',
              dataIndex: 'status',
              key: 'status',
              width: 80,
              render: (v: string) => <Tag color={v === 'completed' ? 'green' : 'default'}>{v || '-'}</Tag>,
            },
            {
              title: '操作',
              key: 'actions',
              width: 240,
              render: (_: unknown, r: Inspection) => (
                <Space>
                  <Button size="small" icon={<PlayCircleOutlined />} onClick={() => handleRun(r.id)}>
                    执行
                  </Button>
                  <Button
                    size="small"
                    icon={<FileTextOutlined />}
                    onClick={() => navigate(`/inspections/${r.id}/reports`)}
                  >
                    报告
                  </Button>
                </Space>
              ),
            },
          ]}
        />
      )}

      <Modal
        title="新建巡检任务"
        open={modalOpen}
        onCancel={() => { setModalOpen(false); form.resetFields() }}
        onOk={handleCreate}
        width={700}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="任务名称" rules={[{ required: true }]}>
            <Input placeholder="如：核心应用巡检" />
          </Form.Item>
          <Form.Item name="environment" label="环境" rules={[{ required: true }]}>
            <EnvironmentSelector style={{ width: '100%' }} />
          </Form.Item>
        </Form>
        <div style={{ marginTop: 8 }}>
          <Text strong style={{ display: 'block', marginBottom: 4 }}>步骤（JSON）</Text>
          <Input.TextArea
            rows={8}
            value={stepsText}
            onChange={(e) => setStepsText(e.target.value)}
            placeholder='[{"plugin":"mock","action":"echo","parameters":{"message":"step1"}}]'
            style={{ fontFamily: 'monospace', fontSize: 13 }}
          />
        </div>
      </Modal>
    </div>
  )
}
