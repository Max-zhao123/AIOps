import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, App, Space, Typography, Tag, Statistic, Row, Col, Empty } from 'antd'
import { PlusOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { listRunbooks, createRunbook, executeRunbook } from '@/api/runbooks'
import type { Runbook, RunbookStep } from '@/types'

const { Title, Text } = Typography

const DEFAULT_STEPS: RunbookStep[] = [{ plugin: 'mock', action: 'echo', parameters: { message: 'step1' } }]

export default function Runbooks() {
  const [runbooks, setRunbooks] = useState<Runbook[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [stepsText, setStepsText] = useState(JSON.stringify(DEFAULT_STEPS, null, 2))
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listRunbooks()
      setRunbooks(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally { setLoading(false) }
  }, [message])

  useEffect(() => { load() }, [load])

  const handleCreate = async () => {
    const name = form.getFieldValue('name')
    if (!name) { message.warning('请输入名称'); return }
    try {
      const steps = JSON.parse(stepsText)
      await createRunbook({ name, steps })
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

  const handleExecute = async (id: number) => {
    try {
      const result = await executeRunbook(id)
      message.success('执行完成')
      load()
    } catch {
      message.error('执行失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>自愈手册</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          新建 Runbook
        </Button>
      </div>

      {runbooks.length === 0 && !loading ? (
        <Empty description="暂无 Runbook" />
      ) : (
        <Table
          dataSource={runbooks}
          rowKey="id"
          loading={loading}
          columns={[
            { title: '名称', dataIndex: 'name', key: 'name' },
            {
              title: '步骤数',
              dataIndex: 'steps',
              key: 'steps',
              width: 80,
              render: (s: RunbookStep[]) => s?.length || 0,
            },
            { title: '最近执行', dataIndex: 'last_run_at', key: 'lastRun', width: 180 },
            {
              title: '成功/失败',
              key: 'stats',
              width: 140,
              render: (_: unknown, r: Runbook) => (
                <Space>
                  <Tag color="green">✓ {r.success_count || 0}</Tag>
                  <Tag color="red">✗ {r.fail_count || 0}</Tag>
                </Space>
              ),
            },
            {
              title: '操作',
              key: 'actions',
              width: 120,
              render: (_: unknown, r: Runbook) => (
                <Button
                  size="small"
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={() => handleExecute(r.id)}
                >
                  执行
                </Button>
              ),
            },
          ]}
        />
      )}

      <Modal
        title="新建自愈手册"
        open={modalOpen}
        onCancel={() => { setModalOpen(false); form.resetFields() }}
        onOk={handleCreate}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="如：重启服务" />
          </Form.Item>
        </Form>
        <div>
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
