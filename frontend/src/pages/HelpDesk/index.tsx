import { useState, useRef } from 'react'
import { Typography, Input, Button, App, Avatar, Space, Empty, Spin, Card } from 'antd'
import { SendOutlined, CustomerServiceOutlined, UserOutlined } from '@ant-design/icons'
import { helpdeskChat } from '@/api/helpdesk'
import StreamingMessage from '@/components/StreamingMessage'

const { Title, Text } = Typography
const { TextArea } = Input

interface Message {
  role: 'user' | 'assistant'
  content: string
}

export default function HelpDesk() {
  const [messages, setMessages] = useState<Message[]>([])
  const [inputValue, setInputValue] = useState('')
  const [loading, setLoading] = useState(false)
  const { message: antMsg } = App.useApp()
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const send = async () => {
    const text = inputValue.trim()
    if (!text || loading) return

    setMessages((prev) => [...prev, { role: 'user', content: text }])
    setInputValue('')
    setLoading(true)

    try {
      const res = await helpdeskChat(text)
      setMessages((prev) => [...prev, { role: 'assistant', content: res.reply || '抱歉，我无法回答这个问题。' }])
    } catch {
      setMessages((prev) => [...prev, { role: 'assistant', content: '请求失败，请稍后重试。' }])
    } finally {
      setLoading(false)
      setTimeout(() => messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }), 100)
    }
  }

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <CustomerServiceOutlined style={{ fontSize: 36, color: '#1677ff' }} />
        <Title level={3} style={{ marginTop: 8 }}>IT HelpDesk</Title>
        <Text type="secondary">智能服务台，解答常见IT问题</Text>
      </div>

      <Card
        style={{
          minHeight: 420,
          maxHeight: 'calc(100vh - 280px)',
          overflow: 'auto',
          marginBottom: 16,
        }}
      >
        {messages.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 60, color: '#999' }}>
            <Text type="secondary">输入问题，如：电脑如何加域？/ 如何申请邮箱？/ 上不了网怎么办？</Text>
          </div>
        ) : (
          messages.map((msg, i) => (
            <div
              key={i}
              style={{
                display: 'flex',
                marginBottom: 16,
                justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
              }}
            >
              {msg.role === 'assistant' && (
                <Avatar
                  icon={<CustomerServiceOutlined />}
                  style={{ background: '#1677ff', marginRight: 10, flexShrink: 0 }}
                />
              )}
              <div
                style={{
                  maxWidth: '80%',
                  padding: '10px 16px',
                  borderRadius: 12,
                  background: msg.role === 'user' ? '#1677ff' : '#f5f5f5',
                  color: msg.role === 'user' ? '#fff' : '#333',
                  borderBottomLeftRadius: msg.role === 'assistant' ? 4 : 12,
                  borderBottomRightRadius: msg.role === 'user' ? 4 : 12,
                }}
              >
                {msg.role === 'user' ? (
                  <Text style={{ color: '#fff' }}>{msg.content}</Text>
                ) : (
                  <StreamingMessage content={msg.content} />
                )}
              </div>
              {msg.role === 'user' && (
                <Avatar
                  icon={<UserOutlined />}
                  style={{ background: '#52c41a', marginLeft: 10, flexShrink: 0 }}
                />
              )}
            </div>
          ))
        )}
        {loading && (
          <div style={{ textAlign: 'center', padding: 8 }}>
            <Text type="secondary"><Spin size="small" /> 思考中....</Text>
          </div>
        )}
        <div ref={messagesEndRef} />
      </Card>

      <Space.Compact style={{ width: '100%' }}>
        <TextArea
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault()
              send()
            }
          }}
          placeholder="输入问题，回车发送..."
          autoSize={{ minRows: 1, maxRows: 4 }}
          disabled={loading}
        />
        <Button
          type="primary"
          icon={<SendOutlined />}
          onClick={send}
          loading={loading}
        >
          发送
        </Button>
      </Space.Compact>
    </div>
  )
}
