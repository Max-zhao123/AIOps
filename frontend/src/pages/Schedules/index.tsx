import { useState, useEffect, useCallback } from 'react'
import {
  Table, Button, Modal, Form, Input, Select, Switch, App, Space, Typography, Tag, Popconfirm, message as antMsg,
} from 'antd'
import {
  PlusOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined,
  ClockCircleOutlined, CheckCircleOutlined, CloseCircleOutlined,
} from '@ant-design/icons'
import {
  listSchedules, createSchedule, updateSchedule, deleteSchedule,
  toggleSchedule, triggerSchedule, listScheduleExecutions,
} from '@/api/schedules'
import type { Schedule, ScheduleExecution, CreateScheduleRequest } from '@/types'

const { Title } = Typography

/** F-040: 定时调度管理页 */
export default function Schedules() {
  const [schedules, setSchedules] = useState<Schedule[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Schedule | null>(null)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listSchedules()
      setSchedules(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载调度列表失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ timezone: 'Asia/Shanghai', enabled: true })
    setModalOpen(true)
  }

  const openEdit = (record: Schedule) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      cron: record.cron,
      timezone: record.timezone,
      target_type: record.target_type,
      target_id: record.target_id,
      enabled: record.enabled,
    })
    setModalOpen(true)
  }

  const handleSubmit = async (values: CreateScheduleRequest) => {
    try {
      if (editing) {
        await updateSchedule(editing.id, values)
        message.success('更新成功')
      } else {
        await createSchedule(values)
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
      await deleteSchedule(id)
      message.success('删除成功')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  const handleToggle = async (id: number, enabled: boolean) => {
    try {
      await toggleSchedule(id, enabled)
      message.success(enabled ? '已启用' : '已禁用')
      load()
    } catch {
      message.error('操作失败')
    }
  }

  const handleTrigger = async (id: number) => {
    try {
      await triggerSchedule(id)
      message.success('已触发执行')
      load()
    } catch {
      message.error('触发失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>定时调度管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          新建调度
        </Button>
      </div>

      <Table
        dataSource={schedules}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 20, showTotal: (t) => `共 ${t} 条` }}
        expandable={{
          expandedRowRender: (record) => <ExecutionHistoryTable scheduleId={record.id} />,
        }}
        columns={[
          { title: '名称', dataIndex: 'name', key: 'name', width: 180 },
          { title: 'Cron 表达式', dataIndex: 'cron', key: 'cron', width: 140 },
          { title: '时区', dataIndex: 'timezone', key: 'timezone', width: 140 },
          {
            title: '目标类型', dataIndex: 'target_type', key: 'target_type', width: 120,
            render: (v: string) => <Tag>{v}</Tag>,
          },
          { title: '目标ID', dataIndex: 'target_id', key: 'target_id', width: 80 },
          {
            title: '状态', dataIndex: 'enabled', key: 'enabled', width: 80,
            render: (v: boolean, r: Schedule) => <Switch checked={v} size="small" onChange={(checked) => handleToggle(r.id, checked)} />,
          },
          {
            title: '上次执行', dataIndex: 'last_run_at', key: 'last_run_at', width: 170,
            render: (v: string) => v ? new Date(v).toLocaleString() : '-',
          },
          {
            title: '下次执行', dataIndex: 'next_run_at', key: 'next_run_at', width: 170,
            render: (v: string) => v ? new Date(v).toLocaleString() : '-',
          },
          {
            title: '操作', key: 'actions', width: 200,
            render: (_: unknown, r: Schedule) => (
              <Space>
                <Button size="small" icon={<PlayCircleOutlined />} onClick={() => handleTrigger(r.id)}>
                  触发
                </Button>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>
                  编辑
                </Button>
                <Popconfirm title="确定删除此调度？" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editing ? '编辑调度' : '新建调度'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields() }}
        onOk={() => form.submit()}
        width={520}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="调度名称" rules={[{ required: true, message: '请输入调度名称' }]}>
            <Input placeholder="如：每日巡检" />
          </Form.Item>
          <Form.Item name="cron" label="Cron 表达式" rules={[{ required: true, message: '请输入 Cron 表达式' }]}
            extra="如：0 8 * * * 表示每天 8:00">
            <Input placeholder="0 8 * * *" />
          </Form.Item>
          <Form.Item name="timezone" label="时区" rules={[{ required: true, message: '请选择时区' }]}>
            <Select options={[
              { label: 'Asia/Shanghai', value: 'Asia/Shanghai' },
              { label: 'UTC', value: 'UTC' },
              { label: 'America/New_York', value: 'America/New_York' },
              { label: 'Europe/London', value: 'Europe/London' },
            ]} />
          </Form.Item>
          <Form.Item name="target_type" label="目标类型" rules={[{ required: true, message: '请选择目标类型' }]}>
            <Select options={[
              { label: '巡检任务', value: 'inspection' },
              { label: 'Runbook', value: 'runbook' },
            ]} />
          </Form.Item>
          <Form.Item name="target_id" label="目标 ID" rules={[{ required: true, message: '请输入目标 ID' }]}>
            <Input type="number" placeholder="1" />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

/** 展开行：执行历史列表 */
function ExecutionHistoryTable({ scheduleId }: { scheduleId: number }) {
  const [executions, setExecutions] = useState<ScheduleExecution[]>([])
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listScheduleExecutions(scheduleId)
      setExecutions(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载执行历史失败')
    } finally {
      setLoading(false)
    }
  }, [scheduleId, message])

  useEffect(() => { load() }, [load])

  return (
    <Table
      dataSource={executions}
      rowKey="id"
      loading={loading}
      size="small"
      pagination={{ pageSize: 5 }}
      columns={[
        {
          title: '状态', dataIndex: 'status', key: 'status', width: 100,
          render: (v: string) => {
            const config: Record<string, { color: string; icon: React.ReactNode }> = {
              success: { color: 'green', icon: <CheckCircleOutlined /> },
              failed: { color: 'red', icon: <CloseCircleOutlined /> },
              running: { color: 'blue', icon: <ClockCircleOutlined /> },
            }
            const c = config[v] || { color: 'default', icon: null }
            return <Tag color={c.color} icon={c.icon}>{v}</Tag>
          },
        },
        { title: '开始时间', dataIndex: 'started_at', key: 'started_at', width: 170,
          render: (v: string) => v ? new Date(v).toLocaleString() : '-' },
        { title: '结束时间', dataIndex: 'completed_at', key: 'completed_at', width: 170,
          render: (v: string) => v ? new Date(v).toLocaleString() : '-' },
        { title: '结果', dataIndex: 'result_summary', key: 'result_summary', ellipsis: true },
        { title: '错误', dataIndex: 'error_message', key: 'error_message', ellipsis: true,
          render: (v: string) => v ? <span style={{ color: '#ff4d4f' }}>{v}</span> : '-' },
      ]}
    />
  )
}
