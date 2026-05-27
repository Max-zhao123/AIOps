import { useState, useEffect, useCallback } from 'react'
import {
  Table, Button, Modal, Form, Input, Select, Switch, App, Space, Typography, Popconfirm, Tag,
} from 'antd'
import { PlusOutlined, ExperimentOutlined, ImportOutlined } from '@ant-design/icons'
import {
  listPolicies, createPolicy, updatePolicy, enablePolicy, disablePolicy, importTemplate,
} from '@/api/policies'
import type { SecurityPolicy } from '@/types'
import { useNavigate } from 'react-router-dom'
import Editor from '@monaco-editor/react'

const { Title } = Typography
const { TextArea } = Input

const DECISION_COLOR: Record<string, string> = { ALLOW: 'green', ASK: 'orange', DENY: 'red' }

export default function Policies() {
  const [policies, setPolicies] = useState<SecurityPolicy[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [yamlValue, setYamlValue] = useState('')
  const [form] = Form.useForm()
  const { message } = App.useApp()
  const navigate = useNavigate()

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

  const handleEnable = async (id: number, enabled: boolean) => {
    try {
      if (enabled) {
        await disablePolicy(id)
      } else {
        await enablePolicy(id)
      }
      message.success(enabled ? '已禁用' : '已启用')
      load()
    } catch {
      message.error('操作失败')
    }
  }

  const openEditor = (policy?: SecurityPolicy) => {
    if (policy) {
      setEditingId(policy.id)
      form.setFieldsValue({ name: policy.name })
      setYamlValue(policy.spec_yaml || '')
    } else {
      setEditingId(null)
      form.resetFields()
      setYamlValue('defaultDecision: ASK\nrules: []\n')
    }
    setModalOpen(true)
  }

  const handleSave = async () => {
    const name = form.getFieldValue('name')
    if (!name) {
      message.error('请输入策略名称')
      return
    }
    try {
      if (editingId) {
        await updatePolicy(editingId, { name, spec_yaml: yamlValue })
        message.success('更新成功')
      } else {
        await createPolicy({ name, spec_yaml: yamlValue })
        message.success('创建成功')
      }
      setModalOpen(false)
      load()
    } catch {
      message.error('保存失败')
    }
  }

  const handleImport = async () => {
    try {
      const template = await importTemplate()
      if (template) {
        setYamlValue(template)
        message.success('模板已导入')
      }
    } catch {
      message.error('导入失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>安全策略</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>
          新建策略
        </Button>
      </div>

      <Table
        dataSource={policies}
        rowKey="id"
        loading={loading}
        columns={[
          { title: '名称', dataIndex: 'name', key: 'name' },
          {
            title: '默认决策',
            dataIndex: 'spec_yaml',
            key: 'defaultDecision',
            width: 120,
            render: (yaml: string) => {
              const match = yaml?.match(/defaultDecision:\s*(\w+)/)
              const d = match?.[1] || '-'
              return <Tag color={DECISION_COLOR[d] || 'default'}>{d}</Tag>
            },
          },
          {
            title: '状态',
            dataIndex: 'enabled',
            key: 'enabled',
            width: 100,
            render: (_: boolean, record: SecurityPolicy) => (
              <Popconfirm
                title={`确认${record.enabled ? '禁用' : '启用'}？`}
                onConfirm={() => handleEnable(record.id, record.enabled)}
              >
                <Switch
                  checked={record.enabled}
                  checkedChildren="启用"
                  unCheckedChildren="禁用"
                />
              </Popconfirm>
            ),
          },
          { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 180 },
          {
            title: '操作',
            key: 'actions',
            width: 220,
            render: (_: unknown, record: SecurityPolicy) => (
              <Space>
                <Button size="small" onClick={() => openEditor(record)}>编辑</Button>
                <Button
                  size="small"
                  icon={<ExperimentOutlined />}
                  onClick={() => navigate(`/policies/${record.id}/simulate`)}
                >
                  模拟
                </Button>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editingId ? '编辑策略' : '新建策略'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={handleSave}
        width={800}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="策略名称" rules={[{ required: true }]}>
            <Input placeholder="如：K8s 只读策略" />
          </Form.Item>
        </Form>

        <div style={{ marginBottom: 8 }}>
          <Space>
            <Button size="small" icon={<ImportOutlined />} onClick={handleImport}>导入模板</Button>
            <span style={{ color: '#999', fontSize: 12 }}>YAML 编辑器</span>
          </Space>
        </div>

        <div style={{ border: '1px solid #d9d9d9', borderRadius: 6, overflow: 'hidden' }}>
          <Editor
            height="360px"
            language="yaml"
            value={yamlValue}
            onChange={(val) => setYamlValue(val || '')}
            theme="vs-dark"
            options={{
              minimap: { enabled: false },
              fontSize: 13,
              lineNumbers: 'on',
              scrollBeyondLastLine: false,
              automaticLayout: true,
            }}
          />
        </div>
      </Modal>
    </div>
  )
}
