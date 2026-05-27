import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, App, Popconfirm, Space, Typography, Tag } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { listEnvironments, createEnvironment, updateEnvironment, deleteEnvironment } from '@/api/environments'
import type { Environment } from '@/types'

const { Title } = Typography

export default function Environments() {
  const [envs, setEnvs] = useState<Environment[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Environment | null>(null)
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

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (record: Environment) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      description: record.description,
    })
    setModalOpen(true)
  }

  const handleSubmit = async (values: { name: string; slug?: string; description?: string }) => {
    if (!editing) {
      if (!values.slug || !/^[a-z][a-z0-9-]*$/.test(values.slug)) {
        message.error('Slug 必须为小写英文+数字+连字符，如 development')
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
    } else {
      try {
        await updateEnvironment(editing.id, { name: values.name, description: values.description })
        message.success('更新成功')
        setModalOpen(false)
        setEditing(null)
        load()
      } catch {
        message.error('更新失败')
      }
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteEnvironment(id)
      message.success('删除成功')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>环境管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          新建环境
        </Button>
      </div>

      <Table
        dataSource={envs}
        rowKey="id"
        loading={loading}
        columns={[
          { title: '名称', dataIndex: 'name', key: 'name', width: 200 },
          {
            title: 'Slug',
            dataIndex: 'slug',
            key: 'slug',
            width: 160,
            render: (v: string) => <Tag color="blue">{v}</Tag>,
          },
          { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
          {
            title: '创建时间',
            dataIndex: 'created_at',
            key: 'created_at',
            width: 180,
          },
          {
            title: '操作',
            key: 'actions',
            width: 160,
            render: (_: unknown, r: Environment) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>
                  编辑
                </Button>
                <Popconfirm title="确定删除此环境？" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>
                    删除
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editing ? '编辑环境' : '新建环境'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields() }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入环境名称' }]}>
            <Input placeholder="如：开发环境" />
          </Form.Item>
          {!editing && (
            <Form.Item
              name="slug"
              label="Slug"
              rules={[{ required: true, message: '请输入 Slug' }]}
              extra="小写英文+数字+连字符，如 development"
            >
              <Input placeholder="development" />
            </Form.Item>
          )}
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="环境描述（可选）" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
