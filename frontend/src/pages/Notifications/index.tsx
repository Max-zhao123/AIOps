import { useState, useEffect, useCallback } from 'react'
import {
  Table, Button, Modal, Form, Input, Select, Switch, App, Space, Typography, Tag, Popconfirm, Tabs, Card, message as antMsg,
} from 'antd'
import {
  PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined,
} from '@ant-design/icons'
import {
  listChannels, createChannel, updateChannel, deleteChannel, testNotification,
  listPolicies, createPolicy, updatePolicy, deletePolicy,
  listTemplates, createTemplate, updateTemplate, deleteTemplate,
} from '@/api/notifications'
import type {
  NotificationChannel, NotificationPolicy, NotificationTemplate, TestNotificationRequest,
} from '@/types'

const { Title } = Typography
const { TextArea } = Input

/** F-041: 通知通道配置页 */
export default function Notifications() {
  return (
    <div>
      <Title level={4}>通知管理</Title>
      <Tabs
        items={[
          { key: 'channels', label: '通知通道', children: <ChannelTab /> },
          { key: 'policies', label: '通知策略', children: <PolicyTab /> },
          { key: 'templates', label: '通知模板', children: <TemplateTab /> },
        ]}
      />
    </div>
  )
}

// ============ 通道 Tab ============
function ChannelTab() {
  const [channels, setChannels] = useState<NotificationChannel[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationChannel | null>(null)
  const [testOpen, setTestOpen] = useState(false)
  const [testChannelId, setTestChannelId] = useState<number>(0)
  const [testRecipient, setTestRecipient] = useState('')
  const [testMessage, setTestMessage] = useState('')
  const [testLoading, setTestLoading] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listChannels()
      setChannels(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载通道列表失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ type: 'email', enabled: true })
    setModalOpen(true)
  }

  const openEdit = (record: NotificationChannel) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      type: record.type,
      enabled: record.enabled,
    })
    setModalOpen(true)
  }

  const handleSubmit = async (values: Partial<NotificationChannel>) => {
    try {
      const payload = { ...values, config: values.config || {} }
      if (editing) {
        await updateChannel(editing.id, payload)
        message.success('更新成功')
      } else {
        await createChannel(payload)
        message.success('创建成功')
      }
      setModalOpen(false)
      setEditing(null)
      form.resetFields()
      load()
    } catch {
      message.error(editing ? '更新失败' : '创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteChannel(id)
      message.success('删除成功')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  const handleTestSend = async () => {
    setTestLoading(true)
    try {
      const result = await testNotification(testChannelId, {
        recipient: testRecipient || undefined,
        message: testMessage || undefined,
      })
      if (result.success) {
        message.success('测试发送成功')
      } else {
        message.error(`发送失败: ${result.message}`)
      }
      setTestOpen(false)
      setTestRecipient('')
      setTestMessage('')
    } catch {
      message.error('测试发送失败')
    } finally {
      setTestLoading(false)
    }
  }

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建通道</Button>
      </div>

      <Table
        dataSource={channels}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 20, showTotal: (t) => `共 ${t} 条` }}
        columns={[
          { title: '名称', dataIndex: 'name', key: 'name', width: 180 },
          {
            title: '类型', dataIndex: 'type', key: 'type', width: 100,
            render: (v: string) => {
              const colorMap: Record<string, string> = { email: 'blue', webhook: 'green', sms: 'orange' }
              return <Tag color={colorMap[v] || 'default'}>{v}</Tag>
            },
          },
          {
            title: '状态', dataIndex: 'enabled', key: 'enabled', width: 80,
            render: (v: boolean) => <Tag color={v ? 'green' : 'red'}>{v ? '启用' : '禁用'}</Tag>,
          },
          {
            title: '配置', dataIndex: 'config', key: 'config', ellipsis: true,
            render: (v: Record<string, unknown>) => JSON.stringify(v),
          },
          {
            title: '操作', key: 'actions', width: 260,
            render: (_: unknown, r: NotificationChannel) => (
              <Space>
                <Button size="small" icon={<SendOutlined />}
                  onClick={() => { setTestChannelId(r.id); setTestOpen(true) }}>
                  测试
                </Button>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>编辑</Button>
                <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      {/* 通道编辑 Modal */}
      <Modal
        title={editing ? '编辑通道' : '新建通道'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields() }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="通道名称" rules={[{ required: true, message: '请输入通道名称' }]}>
            <Input placeholder="如：邮件通知" />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[
              { label: 'Email', value: 'email' },
              { label: 'Webhook', value: 'webhook' },
              { label: 'SMS', value: 'sms' },
            ]} />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* 测试发送 Modal */}
      <Modal
        title="测试发送通知"
        open={testOpen}
        onCancel={() => { setTestOpen(false); setTestRecipient(''); setTestMessage('') }}
        onOk={handleTestSend}
        confirmLoading={testLoading}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Input
            placeholder="收件人（可选）"
            value={testRecipient}
            onChange={(e) => setTestRecipient(e.target.value)}
          />
          <TextArea
            rows={3}
            placeholder="测试消息内容（可选）"
            value={testMessage}
            onChange={(e) => setTestMessage(e.target.value)}
          />
        </Space>
      </Modal>
    </>
  )
}

// ============ 策略 Tab ============
function PolicyTab() {
  const [policies, setPolicies] = useState<NotificationPolicy[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationPolicy | null>(null)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listPolicies()
      setPolicies(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载策略列表失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ enabled: true, severity: 'critical', channel_ids: '', silence_window_min: 0, suppress_lower_severity: false })
    setModalOpen(true)
  }

  const openEdit = (record: NotificationPolicy) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      channel_ids: record.channel_ids,
      severity: record.severity,
      silence_window_min: record.silence_window_min,
      suppress_lower_severity: record.suppress_lower_severity,
      enabled: record.enabled,
    })
    setModalOpen(true)
  }

  const handleSubmit = async (values: Partial<NotificationPolicy>) => {
    try {
      if (editing) {
        await updatePolicy(editing.id, values)
        message.success('更新成功')
      } else {
        await createPolicy(values)
        message.success('创建成功')
      }
      setModalOpen(false)
      setEditing(null)
      form.resetFields()
      load()
    } catch {
      message.error(editing ? '更新失败' : '创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deletePolicy(id)
      message.success('删除成功')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建策略</Button>
      </div>
      <Table
        dataSource={policies}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 20 }}
        columns={[
          { title: '策略名称', dataIndex: 'name', key: 'name', width: 180 },
          { title: '通道 IDs', dataIndex: 'channel_ids', key: 'channel_ids', width: 140 },
          {
            title: '严重级别', dataIndex: 'severity', key: 'severity', width: 100,
            render: (v: string) => <Tag color="orange">{v}</Tag>,
          },
          {
            title: '静默窗口 (min)', dataIndex: 'silence_window_min', key: 'silence_window_min', width: 140,
          },
          {
            title: '抑制低级别', dataIndex: 'suppress_lower_severity', key: 'suppress_lower_severity', width: 120,
            render: (v: boolean) => <Tag color={v ? 'blue' : 'default'}>{v ? '是' : '否'}</Tag>,
          },
          {
            title: '状态', dataIndex: 'enabled', key: 'enabled', width: 80,
            render: (v: boolean) => <Tag color={v ? 'green' : 'red'}>{v ? '启用' : '禁用'}</Tag>,
          },
          {
            title: '操作', key: 'actions', width: 160,
            render: (_: unknown, r: NotificationPolicy) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>编辑</Button>
                <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editing ? '编辑策略' : '新建策略'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields() }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="策略名称" rules={[{ required: true, message: '请输入策略名称' }]}>
            <Input placeholder="如：高严重级别通知" />
          </Form.Item>
          <Form.Item name="channel_ids" label="通知通道 IDs" rules={[{ required: true, message: '请输入通道 IDs' }]}
            extra="多个通道 ID 用英文逗号分隔，如：1,2,3">
            <Input placeholder="1,2,3" />
          </Form.Item>
          <Form.Item name="severity" label="严重级别过滤" rules={[{ required: true, message: '请选择严重级别' }]}>
            <Select options={[
              { label: 'Critical', value: 'critical' },
              { label: 'High', value: 'high' },
              { label: 'Medium', value: 'medium' },
              { label: 'Low', value: 'low' },
            ]} />
          </Form.Item>
          <Form.Item name="silence_window_min" label="静默窗口 (分钟)" rules={[{ required: true, message: '请输入静默窗口' }]}>
            <Input type="number" placeholder="0" />
          </Form.Item>
          <Form.Item name="suppress_lower_severity" label="抑制低级别通知" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

// ============ 模板 Tab ============
function TemplateTab() {
  const [templates, setTemplates] = useState<NotificationTemplate[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<NotificationTemplate | null>(null)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listTemplates()
      setTemplates(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载模板列表失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ channel_type: 'email' })
    setModalOpen(true)
  }

  const openEdit = (record: NotificationTemplate) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      channel_type: record.channel_type,
      subject: record.subject,
      body: record.body,
    })
    setModalOpen(true)
  }

  const handleSubmit = async (values: Partial<NotificationTemplate>) => {
    try {
      if (editing) {
        await updateTemplate(editing.id, values)
        message.success('更新成功')
      } else {
        await createTemplate(values)
        message.success('创建成功')
      }
      setModalOpen(false)
      setEditing(null)
      form.resetFields()
      load()
    } catch {
      message.error(editing ? '更新失败' : '创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteTemplate(id)
      message.success('删除成功')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建模板</Button>
      </div>
      <Table
        dataSource={templates}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 20 }}
        columns={[
          { title: '模板名称', dataIndex: 'name', key: 'name', width: 180 },
          {
            title: '通道类型', dataIndex: 'channel_type', key: 'channel_type', width: 100,
            render: (v: string) => <Tag>{v}</Tag>,
          },
          { title: '主题', dataIndex: 'subject', key: 'subject', width: 200, ellipsis: true },
          { title: '内容', dataIndex: 'body', key: 'body', ellipsis: true },
          {
            title: '操作', key: 'actions', width: 160,
            render: (_: unknown, r: NotificationTemplate) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>编辑</Button>
                <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editing ? '编辑模板' : '新建模板'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields() }}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="模板名称" rules={[{ required: true, message: '请输入模板名称' }]}>
            <Input placeholder="如：告警通知模板" />
          </Form.Item>
          <Form.Item name="channel_type" label="通道类型" rules={[{ required: true }]}>
            <Select options={[
              { label: 'Email', value: 'email' },
              { label: 'Webhook', value: 'webhook' },
              { label: 'SMS', value: 'sms' },
            ]} />
          </Form.Item>
          <Form.Item name="subject" label="主题">
            <Input placeholder="邮件主题（仅 Email 类型需要）" />
          </Form.Item>
          <Form.Item name="body" label="内容模板" rules={[{ required: true, message: '请输入模板内容' }]}
            extra="支持变量占位符：{{.EventName}}, {{.Severity}}, {{.Message}} 等">
            <TextArea rows={6} placeholder="通知内容模板..." />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
