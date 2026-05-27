import { useState, useEffect, useRef, useCallback } from 'react'
import {
  Layout, Input, Button, List, Typography, Space, Tag, Avatar, App, Spin, Empty, Divider, Select,
} from 'antd'
import {
  SendOutlined, PlusOutlined, UserOutlined, RobotOutlined,
  StopOutlined, ClearOutlined,
} from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { useChatStore } from '@/stores/chatStore'
import { useSSE } from '@/hooks/useSSE'
import EnvironmentSelector from '@/components/EnvironmentSelector'
import ActionPlanCard from '@/components/ActionPlanCard'
import StreamingMessage from '@/components/StreamingMessage'
import {
  postChat, listSessions, listMessages,
} from '@/api/chat'
import { confirmAction } from '@/api/executor'
import { listLlmConfigs } from '@/api/llm'
import type { ChatMessage, ActionPlan, LlmConfig } from '@/types'

const { Sider, Content } = Layout
const { Text, Title } = Typography
const { TextArea } = Input

export default function Chat() {
  const { id: sessionIdParam } = useParams<{ id: string }>()
  const token = useAuthStore((s) => s.token)
  const {
    sessions, messages, setSessions, setMessages, addMessage,
    appendToLastAssistant, isStreaming, setStreaming, setCurrentSession,
  } = useChatStore()
  const { connect, disconnect } = useSSE()
  const navigate = useNavigate()
  const { message: antMsg } = App.useApp()

  const [inputValue, setInputValue] = useState('')
  const [environment, setEnvironment] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [sessionsLoading, setSessionsLoading] = useState(false)
  const [rightTab, setRightTab] = useState<'plans' | 'assist' | 'rca' | null>(null)
  const [assistInput, setAssistInput] = useState('')
  const [rcaInput, setRcaInput] = useState('')
  const [selectedModel, setSelectedModel] = useState<string>('')
  const [models, setModels] = useState<LlmConfig[]>([])

  useEffect(() => {
    listLlmConfigs().then((data) => {
      const list = Array.isArray(data) ? data : []
      setModels(list.filter((m) => m.active))
    }).catch(() => {})
  }, [])

  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => { scrollToBottom() }, [messages])

  const loadSessions = useCallback(async () => {
    setSessionsLoading(true)
    try {
      const data = await listSessions()
      setSessions(Array.isArray(data) ? data : [])
    } catch { /* ignore */ }
    finally { setSessionsLoading(false) }
  }, [setSessions])

  useEffect(() => { loadSessions() }, [loadSessions])

  // 加载指定会话的消息
  useEffect(() => {
    if (!sessionIdParam) {
      setMessages([])
      setCurrentSession(null)
      return
    }
    const sid = Number(sessionIdParam)
    if (isNaN(sid)) return
    setCurrentSession(sid)
    listMessages(sid).then((data) => {
      setMessages(Array.isArray(data) ? data : [])
    }).catch(() => {})
  }, [sessionIdParam, setMessages, setCurrentSession])

  const sendMessage = async () => {
    const text = inputValue.trim()
    if (!text || isStreaming) return
    if (!environment) {
      antMsg.warning('请先选择环境')
      return
    }

    const userMsg: ChatMessage = {
      id: Date.now(),
      session_id: 0,
      role: 'user',
      content: text,
      created_at: new Date().toISOString(),
    }
    addMessage(userMsg)
    setInputValue('')
    setStreaming(true)

    // 空 AI 消息占位
    const aiMsg: ChatMessage = {
      id: Date.now() + 1,
      session_id: 0,
      role: 'assistant',
      content: '',
      created_at: new Date().toISOString(),
    }
    addMessage(aiMsg)

    try {
      const baseUrl = `${window.location.origin}/api/v1`
      const params = new URLSearchParams({ environment, message: text })
      if (selectedModel) params.set('model', selectedModel)
      const url = `${baseUrl}/chat/stream?${params}`

      await connect(url, token, {
        onMessage: (data) => {
          try {
            const parsed = JSON.parse(data)
            if (parsed.text) {
              appendToLastAssistant(parsed.text)
            } else if (typeof parsed === 'string') {
              appendToLastAssistant(parsed)
            }
          } catch {
            appendToLastAssistant(data)
          }
        },
        onComplete: () => {
          setStreaming(false)
          loadSessions()
        },
        onError: () => {
          appendToLastAssistant('\n\n> ⚠️ 连接中断，请重试')
          setStreaming(false)
        },
      })
    } catch {
      appendToLastAssistant('请求失败，请检查后端服务')
      setStreaming(false)
    }
  }

  const handleConfirm = async (executionId: number, approved: boolean) => {
    try {
      await confirmAction({ executionId, approved })
      antMsg.success(approved ? '已确认执行' : '已拒绝')
    } catch {
      antMsg.error('操作失败')
    }
  }

  const newChat = () => {
    setMessages([])
    setCurrentSession(null)
    navigate('/chat')
    setInputValue('')
  }

  // 从消息中提取 ActionPlan
  const extractActionPlans = (msg: ChatMessage): ActionPlan[] => {
    if (msg.action_plans) return msg.action_plans
    return []
  }

  const allPlans = messages.flatMap(extractActionPlans)

  return (
    <Layout style={{ background: '#fff', borderRadius: 8, overflow: 'hidden', height: 'calc(100vh - 112px)' }}>
      {/* 左侧会话列表 */}
      <Sider
        width={220}
        style={{ background: '#fafafa', borderRight: '1px solid #f0f0f0', overflow: 'auto' }}
      >
        <div style={{ padding: '12px' }}>
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <Text strong>历史会话</Text>
            <Button size="small" type="text" icon={<PlusOutlined />} onClick={newChat} />
          </Space>
        </div>
        <Divider style={{ margin: 0 }} />
        {sessionsLoading ? (
          <div style={{ textAlign: 'center', padding: 20 }}><Spin size="small" /></div>
        ) : sessions.length === 0 ? (
          <Empty description="暂无会话" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <List
            size="small"
            dataSource={sessions}
            renderItem={(s) => (
              <List.Item
                style={{
                  padding: '8px 12px',
                  cursor: 'pointer',
                  background: Number(sessionIdParam) === s.id ? '#e6f4ff' : 'transparent',
                }}
                onClick={() => navigate(`/chat/sessions/${s.id}`)}
              >
                <Text ellipsis style={{ fontSize: 13 }}>{s.title || `会话 #${s.id}`}</Text>
              </List.Item>
            )}
          />
        )}
      </Sider>

      {/* 中间对话区 */}
      <Content style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <div style={{ flex: 1, overflow: 'auto', padding: '16px 24px' }}>
          {messages.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '120px 0', color: '#999' }}>
              <RobotOutlined style={{ fontSize: 48, marginBottom: 16 }} />
              <Title level={5} type="secondary">AI 运维助手</Title>
              <Text type="secondary">
                选择环境，输入运维需求，AI 将自动分析并提供操作建议
              </Text>
            </div>
          ) : (
            messages.map((msg, idx) => (
              <div
                key={msg.id || idx}
                style={{
                  marginBottom: 16,
                  display: 'flex',
                  justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
                }}
              >
                {msg.role !== 'user' && (
                  <Avatar
                    icon={<RobotOutlined />}
                    style={{ background: '#1677ff', marginRight: 10, flexShrink: 0 }}
                  />
                )}
                <div style={{ maxWidth: '75%' }}>
                  {msg.role === 'user' ? (
                    <div
                      style={{
                        background: '#1677ff',
                        color: '#fff',
                        padding: '10px 16px',
                        borderRadius: 12,
                        borderBottomRightRadius: 4,
                      }}
                    >
                      <Text style={{ color: '#fff' }}>{msg.content}</Text>
                    </div>
                  ) : (
                    <div
                      style={{
                        background: '#f5f5f5',
                        padding: '12px 16px',
                        borderRadius: 12,
                        borderBottomLeftRadius: 4,
                        border: '1px solid #e8e8e8',
                      }}
                    >
                      {(idx === messages.length - 1 && isStreaming && msg.content === '')
                        ? <Text type="secondary" className="streaming-cursor">思考中</Text>
                        : <StreamingMessage
                            content={msg.content}
                            isStreaming={idx === messages.length - 1 && isStreaming}
                          />
                      }
                      {/* Action Plans */}
                      {msg.action_plans && msg.action_plans.length > 0 && (
                        <div style={{ marginTop: 12 }}>
                          {msg.action_plans.map((plan, pi) => (
                            <ActionPlanCard key={pi} plan={plan} />
                          ))}
                          {/* ASK 确认按钮 */}
                          {msg.action_plans.some((p) => p.risk === 'write') && (
                            <Space style={{ marginTop: 8 }}>
                              <Button
                                type="primary"
                                size="small"
                                onClick={() => handleConfirm(msg.id || 0, true)}
                              >
                                确认执行
                              </Button>
                              <Button
                                danger
                                size="small"
                                onClick={() => handleConfirm(msg.id || 0, false)}
                              >
                                拒绝
                              </Button>
                            </Space>
                          )}
                        </div>
                      )}
                    </div>
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
          <div ref={messagesEndRef} />
        </div>

        {/* 底部输入 */}
        <div style={{ padding: '12px 24px', borderTop: '1px solid #f0f0f0', background: '#fff' }}>
          {/* 工具栏 — 参照 CodeBuddy 风格 */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              marginBottom: 8,
              padding: '6px 10px',
              background: '#1e1e2e',
              borderRadius: 8,
            }}
          >
            <EnvironmentSelector
              value={environment}
              onChange={setEnvironment}
              style={{
                minWidth: 140,
                background: '#2d2d3d',
                borderRadius: 6,
              }}
              placeholder="选择环境"
            />
            <Select
              value={selectedModel || undefined}
              onChange={setSelectedModel}
              placeholder={models.length > 0 ? '选择模型' : '暂无模型'}
              allowClear
              disabled={models.length === 0}
              options={models.map((m) => ({ label: m.name, value: m.name }))}
              style={{
                minWidth: 140,
                background: '#2d2d3d',
                borderRadius: 6,
              }}
              dropdownStyle={{ minWidth: 180 }}
            />
            {models.length === 0 && (
              <Button
                type="link"
                size="small"
                style={{ color: '#a0a0b0', padding: 0 }}
                onClick={() => navigate('/llm-config')}
              >
                去配置模型 →
              </Button>
            )}
          </div>

          <Space.Compact style={{ width: '100%' }}>
            <TextArea
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault()
                  sendMessage()
                }
              }}
              placeholder="输入运维需求，如：查看所有 Pod 状态 | Shift+Enter 换行"
              autoSize={{ minRows: 1, maxRows: 4 }}
              style={{ flex: 1 }}
            />
            {isStreaming ? (
              <Button icon={<StopOutlined />} danger onClick={disconnect}>
                停止
              </Button>
            ) : (
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={sendMessage}
              >
                发送
              </Button>
            )}
          </Space.Compact>
        </div>
      </Content>

      {/* 右侧详情面板 */}
      <Sider width={280} style={{ background: '#fafafa', borderLeft: '1px solid #f0f0f0', overflow: 'auto', padding: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <div>
            <Text strong>操作面板</Text>
            <Divider style={{ margin: '8px 0' }} />
            <Space size={4}>
              <Button
                size="small"
                type={rightTab === 'plans' ? 'primary' : 'default'}
                onClick={() => setRightTab(rightTab === 'plans' ? null : 'plans')}
              >
                ActionPlans
              </Button>
              <Button
                size="small"
                type={rightTab === 'assist' ? 'primary' : 'default'}
                onClick={() => setRightTab(rightTab === 'assist' ? null : 'assist')}
              >
                修复辅助
              </Button>
              <Button
                size="small"
                type={rightTab === 'rca' ? 'primary' : 'default'}
                onClick={() => setRightTab(rightTab === 'rca' ? null : 'rca')}
              >
                RCA
              </Button>
            </Space>
          </div>

          {rightTab === 'plans' && (
            <div>
              <Text type="secondary" style={{ fontSize: 12 }}>
                当前对话中的 ActionPlan ({allPlans.length})
              </Text>
              {allPlans.length === 0 ? (
                <Empty description="无 ActionPlan" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ marginTop: 20 }} />
              ) : (
                allPlans.map((plan, i) => <ActionPlanCard key={i} plan={plan} />)
              )}
            </div>
          )}

          {rightTab === 'assist' && (
            <div>
              <Text type="secondary" style={{ fontSize: 12 }}>输入错误信息，获取修复建议</Text>
              <TextArea
                rows={3}
                value={assistInput}
                onChange={(e) => setAssistInput(e.target.value)}
                placeholder="如：pod crash, CrashLoopBackOff"
                style={{ marginTop: 8 }}
              />
              <Button type="primary" size="small" style={{ marginTop: 8 }} block>
                分析
              </Button>
            </div>
          )}

          {rightTab === 'rca' && (
            <div>
              <Text type="secondary" style={{ fontSize: 12 }}>输入事件描述，进行根因分析</Text>
              <TextArea
                rows={3}
                value={rcaInput}
                onChange={(e) => setRcaInput(e.target.value)}
                placeholder="如：服务响应延迟增加"
                style={{ marginTop: 8 }}
              />
              <Button type="primary" size="small" style={{ marginTop: 8 }} block>
                分析
              </Button>
            </div>
          )}
        </Space>
      </Sider>
    </Layout>
  )
}
