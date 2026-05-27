import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, App, Popconfirm, Space, Typography, Tag, Tabs, Card, message as antMsg } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import {
  listEnvironments, createEnvironment, updateEnvironment, deleteEnvironment,
} from '@/api/environments'
import {
  listEnvironmentConfigs, upsertEnvironmentConfig, deleteEnvironmentConfig,
} from '@/api/environmentConfigs'
import { getEnvironmentQuota, updateEnvironmentQuota } from '@/api/environmentQuotas'
import type { Environment, EnvironmentConfig, EnvironmentQuota, UpsertEnvironmentConfigRequest } from '@/types'
import QuotaProgress from '@/components/QuotaProgress'
import { useAuthStore } from '@/stores/authStore'

const { Title } = Typography
const { TextArea } = Input

/** F-046 + F-051: 环境管理 + 配置编辑 Tab + 配额展示 */
export default function Environments() {
  const [envs, setEnvs] = useState<Environment[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Environment | null>(null)
  const [detailSlug, setDetailSlug] = useState<string | null>(null)
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

  // 如果点击了详情，显示环境详情页
  if (detailSlug) {
    return (
      <EnvironmentDetail
        slug={detailSlug}
        onBack={() => setDetailSlug(null)}
      />
    )
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
            width: 220,
            render: (_: unknown, r: Environment) => (
              <Space>
                <Button size="small" onClick={() => setDetailSlug(r.slug)}>
                  详情
                </Button>
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

/** 环境详情页：基本信息 + 配置编辑 Tab + 配额展示 */
function EnvironmentDetail({ slug, onBack }: { slug: string; onBack: () => void }) {
  const { message } = App.useApp()
  const hasRole = useAuthStore((s) => s.hasRole)
  const isAdmin = hasRole('admin')

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
        <Button onClick={onBack} style={{ marginRight: 12 }}>← 返回</Button>
        <Title level={4} style={{ margin: 0 }}>环境详情：{slug}</Title>
      </div>

      <Tabs
        items={[
          { key: 'configs', label: '配置覆盖', children: <ConfigTab slug={slug} /> },
          { key: 'quota', label: '配额使用量', children: <QuotaTab slug={slug} /> },
        ]}
      />
    </div>
  )
}

/** F-046: 环境配置编辑 Tab */
function ConfigTab({ slug }: { slug: string }) {
  const [configs, setConfigs] = useState<EnvironmentConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listEnvironmentConfigs(slug)
      setConfigs(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载配置失败')
    } finally {
      setLoading(false)
    }
  }, [slug, message])

  useEffect(() => { load() }, [load])

  const handleUpsert = async (values: UpsertEnvironmentConfigRequest) => {
    try {
      await upsertEnvironmentConfig(slug, values)
      message.success('保存成功')
      setModalOpen(false)
      form.resetFields()
      load()
    } catch {
      message.error('保存失败')
    }
  }

  const handleDelete = async (key: string) => {
    try {
      await deleteEnvironmentConfig(slug, key)
      message.success('删除成功')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setModalOpen(true) }}>
          新增/更新配置
        </Button>
      </div>

      <Table
        dataSource={configs}
        rowKey="key"
        loading={loading}
        pagination={false}
        columns={[
          { title: '配置 Key', dataIndex: 'key', key: 'key', width: 200 },
          {
            title: '配置值', dataIndex: 'value', key: 'value', ellipsis: true,
            render: (v: string) => {
              try {
                const parsed = JSON.parse(v)
                return JSON.stringify(parsed, null, 2)
              } catch {
                return v
              }
            },
          },
          {
            title: '覆盖类型', dataIndex: 'override_type', key: 'override_type', width: 120,
            render: (v: string) => <Tag>{v || 'merge'}</Tag>,
          },
          {
            title: '操作', key: 'actions', width: 100,
            render: (_: unknown, r: EnvironmentConfig) => (
              <Popconfirm title="确定删除此配置？" onConfirm={() => handleDelete(r.key)}>
                <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
              </Popconfirm>
            ),
          },
        ]}
      />

      <Modal
        title="新增/更新配置"
        open={modalOpen}
        onCancel={() => { setModalOpen(false); form.resetFields() }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleUpsert}>
          <Form.Item name="key" label="配置 Key" rules={[{ required: true, message: '请输入配置 Key' }]}>
            <Input placeholder="如：max_concurrent_tasks" />
          </Form.Item>
          <Form.Item name="value" label="配置值 (JSON)" rules={[{ required: true, message: '请输入配置值' }]}
            extra="请输入有效的 JSON 格式">
            <TextArea rows={4} placeholder='{"limit": 10}' />
          </Form.Item>
          <Form.Item name="override_type" label="覆盖类型">
            <Input placeholder="merge（默认）" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

/** F-051: 环境配额展示 Tab */
function QuotaTab({ slug }: { slug: string }) {
  const [quota, setQuota] = useState<EnvironmentQuota | null>(null)
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()
  const hasRole = useAuthStore((s) => s.hasRole)
  const isAdmin = hasRole('admin')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getEnvironmentQuota(slug)
      setQuota(data)
    } catch {
      message.error('加载配额失败')
    } finally {
      setLoading(false)
    }
  }, [slug, message])

  useEffect(() => { load() }, [load])

  if (!quota) {
    return (
      <Card loading={loading}>
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          暂无配额数据
        </div>
      </Card>
    )
  }

  const quotaEntries = Object.entries(quota.quotas || {})

  return (
    <Card title={`配额使用量 - ${slug}`} extra={<Button size="small" onClick={load}>刷新</Button>}>
      {quotaEntries.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          暂无配额配置
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 16 }}>
          {quotaEntries.map(([key, item]) => (
            <Card key={key} size="small" style={{ background: '#fafafa' }}>
              <QuotaProgress
                label={key}
                current={item.current_count}
                limit={item.daily_limit}
                resetAt={item.reset_at}
              />
              {item.reset_at && (
                <div style={{ fontSize: 11, color: '#999', marginTop: 4 }}>
                  重置时间：{new Date(item.reset_at).toLocaleString()}
                </div>
              )}
            </Card>
          ))}
        </div>
      )}
    </Card>
  )
}
