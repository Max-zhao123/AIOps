import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, App, Popconfirm, Space, Typography } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { listEnvironments, createEnvironment } from '@/api/environments'
import type { Environment } from '@/types'

const { Title } = Typography

export default function Environments() {
  const [envs, setEnvs] = useState<Environment[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listEnvironments()
      setEnvs(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载环境列表失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => {
    load()
  }, [load])

  const handleCreate = async (values: { name: string; slug: string; description?: string }) => {
    if (!/^[a-z][a-z0-9-]*$/.test(values.slug)) {
      message.error('Slug 必须为小写英文+数字+连字符')
      return
    }
    try {
      await createEnvironment(values)
      message.success('创建成功')
      setModalOpen(false)
      form.resetFields()
      load()
    } catch {
      message.error('创建失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>环境管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          新建环境
        </Button>
      </div>

      <Table
        dataSource={envs}
        rowKey="id"
        loading={loading}
        columns={[
          { title: '名称', dataIndex: 'name', key: 'name' },
          { title: 'Slug', dataIndex: 'slug', key: 'slug' },
          { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
          { title: '创建时间', dataIndex: 'created_at', key: 'created_at' },
        ]}
      />

      <Modal
        title="新建环境"
        open={modalOpen}
        onCancel={() => { setModalOpen(false); form.resetFields() }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入环境名称' }]}>
            <Input placeholder="如：开发环境" />
          </Form.Item>
          <Form.Item
            name="slug"
            label="Slug"
            rules={[{ required: true, message: '请输入 Slug' }]}
            extra="小写英文+数字+连字符，如 development"
          >
            <Input placeholder="development" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="环境描述（可选）" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
