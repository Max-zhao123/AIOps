import { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Tag, App, Spin, Empty, Typography, Space } from 'antd'
import { ApiOutlined } from '@ant-design/icons'
import { listPlugins } from '@/api/plugins'
import type { Plugin } from '@/types'

const { Title, Text } = Typography

export default function Plugins() {
  const [plugins, setPlugins] = useState<Plugin[]>([])
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listPlugins()
      setPlugins(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => { load() }, [load])

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <Title level={4}>插件列表</Title>

      <Row gutter={[16, 16]}>
        {plugins.length === 0 ? (
          <Col span={24}><Empty description="暂无插件" /></Col>
        ) : (
          plugins.map((p) => (
            <Col xs={24} sm={12} lg={8} key={p.name}>
              <Card
                hoverable
                title={
                  <Space>
                    <ApiOutlined />
                    <span>{p.name}</span>
                  </Space>
                }
                extra={
                  <Tag color={p.online ? 'green' : 'red'}>
                    {p.online ? '在线' : '离线'}
                  </Tag>
                }
              >
                <Space direction="vertical" size={4}>
                  <Text type="secondary">类型：</Text>
                  <Tag>{p.type || '通用'}</Tag>
                  <Text type="secondary">描述：</Text>
                  <Text>{p.description || '暂无描述'}</Text>
                </Space>
              </Card>
            </Col>
          ))
        )}
      </Row>
    </div>
  )
}
