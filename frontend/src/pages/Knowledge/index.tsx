import { useState, useEffect, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, App, Popconfirm, Space, Typography, Empty } from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { listDocuments, createDocument, deleteDocument } from '@/api/knowledge'
import type { KBDocument } from '@/types'

const { Title } = Typography
const { TextArea } = Input

export default function Knowledge() {
  const [docs, setDocs] = useState<KBDocument[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listDocuments()
      setDocs(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  const handleCreate = async (values: { title: string; content: string }) => {
    try {
      await createDocument(values)
      message.success('创建成功')
      setModalOpen(false)
      form.resetFields()
      load()
    } catch {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteDocument(id)
      message.success('已删除')
      load()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>知识库</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          新建文档
        </Button>
      </div>

      {docs.length === 0 && !loading ? (
        <Empty description="暂无文档" />
      ) : (
        <Table
          dataSource={docs}
          rowKey="id"
          loading={loading}
          columns={[
            { title: '标题', dataIndex: 'title', key: 'title' },
            { title: '内容预览', dataIndex: 'content', key: 'content', ellipsis: true, width: 400 },
            { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
            { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 180 },
            {
              title: '操作',
              key: 'actions',
              width: 100,
              render: (_: unknown, record: KBDocument) => (
                <Popconfirm title="确认删除？" onConfirm={() => handleDelete(record.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />} />
                </Popconfirm>
              ),
            },
          ]}
        />
      )}

      <Modal
        title="新建知识库文档"
        open={modalOpen}
        onCancel={() => { setModalOpen(false); form.resetFields() }}
        onOk={() => form.submit()}
        width={700}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input placeholder="文档标题" />
          </Form.Item>
          <Form.Item name="content" label="内容（Markdown）" rules={[{ required: true }]}>
            <TextArea rows={12} placeholder="使用 Markdown 格式编写内容" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
