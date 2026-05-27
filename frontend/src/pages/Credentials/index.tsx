import { useState, useEffect, useCallback } from 'react'
import {
  Table, Button, Modal, Form, Input, Select, App, Space, Typography, Tag, Popconfirm, Tooltip,
} from 'antd'
import {
  PlusOutlined, DeleteOutlined, SyncOutlined, KeyOutlined,
} from '@ant-design/icons'
import { listCredentials, createCredential, rotateCredential, deleteCredential } from '@/api/credentials'
import type { Credential, CreateCredentialRequest, RotateCredentialRequest } from '@/types'
import { useAuthStore } from '@/stores/authStore'

const { Title } = Typography

/** F-042: 凭证管理页 */
export default function Credentials() {
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [loading, setLoading] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [rotateOpen, setRotateOpen] = useState(false)
  const [rotatingId, setRotatingId] = useState<number>(0)
  const [createForm] = Form.useForm()
  const [rotateForm] = Form.useForm()
  const { message } = App.useApp()
  const hasRole = useAuthStore((s) => s.hasRole)
  const isAdmin = hasRole('admin')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listCredentials()
      setCredentials(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载凭证列表失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const handleCreate = async (values: CreateCredentialRequest) => {
    try {
      await createCredential(values)
      message.success('创建成功')
      setCreateOpen(false)
      createForm.resetFields()
      load()
    } catch {
      message.error('创建失败')
    }
  }

  const handleRotate = async (values: RotateCredentialRequest) => {
    try {
      await rotateCredential(rotatingId, values)
      message.success('轮转成功')
      setRotateOpen(false)
      rotateForm.resetFields()
      load()
    } catch {
      message.error('轮转失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteCredential(id)
      message.success('删除成功')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  /** 判断是否即将过期（7天内） */
  const isExpiringSoon = (expiresAt?: string) => {
    if (!expiresAt) return false
    const diff = new Date(expiresAt).getTime() - Date.now()
    return diff > 0 && diff < 7 * 24 * 60 * 60 * 1000
  }

  const isExpired = (expiresAt?: string) => {
    if (!expiresAt) return false
    return new Date(expiresAt).getTime() < Date.now()
  }

  if (!isAdmin) {
    return (
      <div style={{ textAlign: 'center', padding: 60 }}>
        <Title level={4} type="secondary">凭证管理仅管理员可见</Title>
      </div>
    )
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>凭证管理</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { createForm.resetFields(); setCreateOpen(true) }}>
          新建凭证
        </Button>
      </div>

      <Table
        dataSource={credentials}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 20, showTotal: (t) => `共 ${t} 条` }}
        rowClassName={(r) => {
          if (isExpired(r.expires_at)) return 'credential-expired'
          if (isExpiringSoon(r.expires_at)) return 'credential-expiring'
          return ''
        }}
        columns={[
          { title: '名称', dataIndex: 'name', key: 'name', width: 180 },
          {
            title: '类型', dataIndex: 'type', key: 'type', width: 120,
            render: (v: string) => {
              const colorMap: Record<string, string> = {
                smtp: 'blue', api_key: 'green', database: 'orange', kubeconfig: 'purple', other: 'default',
              }
              return <Tag color={colorMap[v] || 'default'}>{v}</Tag>
            },
          },
          {
            title: '凭证值', dataIndex: 'value', key: 'value', width: 200,
            render: (v: string) => (
              <Tooltip title="已脱敏">
                <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
                  {'*'.repeat(Math.min(v?.length || 8, 8))}
                </span>
              </Tooltip>
            ),
          },
          {
            title: '过期时间', dataIndex: 'expires_at', key: 'expires_at', width: 170,
            render: (v: string) => {
              if (!v) return <Tag>永不过期</Tag>
              const expired = isExpired(v)
              const expiring = isExpiringSoon(v)
              return (
                <Space>
                  <span>{new Date(v).toLocaleString()}</span>
                  {expired && <Tag color="error">已过期</Tag>}
                  {!expired && expiring && <Tag color="warning">即将过期</Tag>}
                </Space>
              )
            },
          },
          {
            title: '上次轮转', dataIndex: 'last_rotated_at', key: 'last_rotated_at', width: 170,
            render: (v: string) => v ? new Date(v).toLocaleString() : '-',
          },
          {
            title: '操作', key: 'actions', width: 180,
            render: (_: unknown, r: Credential) => (
              <Space>
                <Button size="small" icon={<SyncOutlined />}
                  onClick={() => { setRotatingId(r.id); rotateForm.resetFields(); setRotateOpen(true) }}>
                  轮转
                </Button>
                <Popconfirm title="确定删除此凭证？" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      {/* 新建凭证 Modal */}
      <Modal
        title="新建凭证"
        open={createOpen}
        onCancel={() => { setCreateOpen(false); createForm.resetFields() }}
        onOk={() => createForm.submit()}
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label="凭证名称" rules={[{ required: true, message: '请输入凭证名称' }]}>
            <Input placeholder="如：SMTP 凭证" prefix={<KeyOutlined />} />
          </Form.Item>
          <Form.Item name="type" label="凭证类型" rules={[{ required: true, message: '请选择凭证类型' }]}>
            <Select options={[
              { label: 'SMTP', value: 'smtp' },
              { label: 'API Key', value: 'api_key' },
              { label: 'Database', value: 'database' },
              { label: 'Kubeconfig', value: 'kubeconfig' },
              { label: '其他', value: 'other' },
            ]} />
          </Form.Item>
          <Form.Item name="value" label="凭证值" rules={[{ required: true, message: '请输入凭证值' }]}>
            <Input.TextArea rows={3} placeholder="凭证值（将加密存储）" />
          </Form.Item>
          <Form.Item name="expires_at" label="过期时间（可选）">
            <Input type="datetime-local" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 轮转凭证 Modal */}
      <Modal
        title="轮转凭证"
        open={rotateOpen}
        onCancel={() => { setRotateOpen(false); rotateForm.resetFields() }}
        onOk={() => rotateForm.submit()}
      >
        <Form form={rotateForm} layout="vertical" onFinish={handleRotate}>
          <Form.Item name="new_value" label="新凭证值" rules={[{ required: true, message: '请输入新凭证值' }]}>
            <Input.TextArea rows={3} placeholder="输入新的凭证值" />
          </Form.Item>
          <Form.Item name="expires_at" label="新过期时间（可选）">
            <Input type="datetime-local" />
          </Form.Item>
        </Form>
      </Modal>

      <style>{`
        .credential-expired { background: #fff1f0 !important; }
        .credential-expiring { background: #fffbe6 !important; }
      `}</style>
    </div>
  )
}
