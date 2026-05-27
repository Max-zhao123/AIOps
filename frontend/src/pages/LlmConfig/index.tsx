import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, Switch, App, Popconfirm, Space, Typography, Tag } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { listLlmConfigs, createLlmConfig, updateLlmConfig, deleteLlmConfig } from '@/api/llm'
import type { LlmConfig } from '@/types'

const { Title } = Typography

export default function LlmConfigPage() {
  const [configs, setConfigs] = useState<LlmConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<LlmConfig | null>(null)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listLlmConfigs()
      setConfigs(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载配置失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (r: LlmConfig) => {
    setEditing(r)
    form.setFieldsValue(r)
    setModalOpen(true)
  }

  const handleSubmit = async () => {
    const values = await form.validateFields()
    try {
      if (editing) {
        await updateLlmConfig(editing.id, values)
        message.success('更新成功')
      } else {
        await createLlmConfig(values)
        message.success('创建成功')
      }
      setModalOpen(false)
      setEditing(null)
      form.resetFields()
      load()
    } catch {
      message.error('操作失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteLlmConfig(id)
      message.success('已删除')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>LLM 模型配置</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增模型</Button>
      </div>

      <Table
        dataSource={configs}
        rowKey="id"
        loading={loading}
        columns={[
          { title: '名称', dataIndex: 'name', key: 'name', width: 140 },
          {
            title: 'API 地址',
            dataIndex: 'baseUrl',
            key: 'baseUrl',
            width: 280,
            ellipsis: true,
          },
          { title: '模型', dataIndex: 'model', key: 'model', width: 160 },
          {
            title: '状态',
            dataIndex: 'active',
            key: 'active',
            width: 80,
            render: (v: boolean) => (v ? <Tag color="green">启用</Tag> : <Tag color="default">禁用</Tag>),
          },
          {
            title: '操作',
            key: 'actions',
            width: 160,
            render: (_: unknown, r: LlmConfig) => (
              <Space>
                <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>编辑</Button>
                <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
        locale={{ emptyText: '暂无 LLM 配置，请新增模型' }}
      />

      <Modal
        title={editing ? '编辑模型' : '新增模型'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields() }}
        onOk={handleSubmit}
        width={560}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="配置名称" rules={[{ required: true }]} extra="如 openai / deepseek / local">
            <Input placeholder="openai" />
          </Form.Item>
          <Form.Item
            name="baseUrl"
            label="API 地址"
            rules={[{ required: true, message: '请输入 API 地址' }]}
            extra="OpenAI 兼容接口地址，如 https://api.openai.com/v1"
          >
            <Input placeholder="https://api.openai.com/v1" />
          </Form.Item>
          <Form.Item
            name="apiKey"
            label="API Key"
            rules={[{ required: true, message: '请输入 API Key' }]}
          >
            <Input.Password placeholder="sk-..." />
          </Form.Item>
          <Form.Item
            name="model"
            label="模型名称"
            rules={[{ required: true, message: '请输入模型名称' }]}
            extra="如 gpt-4o-mini / deepseek-chat / qwen-plus"
          >
            <Input placeholder="gpt-4o-mini" />
          </Form.Item>
          {editing && (
            <Form.Item name="active" label="启用" valuePropName="checked">
              <Switch />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
