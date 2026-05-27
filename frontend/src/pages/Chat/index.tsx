import { useState, useEffect, useRef, useCallback } from 'react'
import {
  Layout, Input, Button, List, Typography, Space, Tag, Avatar, App, Spin, Empty, Divider, Select, Tooltip, Image, Popconfirm,
} from 'antd'
import {
  SendOutlined, PlusOutlined, UserOutlined, RobotOutlined,
  StopOutlined, ClearOutlined, PictureOutlined, PaperClipOutlined,
  CloseCircleOutlined, InboxOutlined, DeleteOutlined,
} from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { useChatStore } from '@/stores/chatStore'
import { useSSE } from '@/hooks/useSSE'
import EnvironmentSelector from '@/components/EnvironmentSelector'
import ActionPlanCard from '@/components/ActionPlanCard'
import StreamingMessage from '@/components/StreamingMessage'
import {
  postChat, listSessions, listMessages, archiveSession, deleteSession,
} from '@/api/chat'
import { confirmAction } from '@/api/executor'
import { listLlmConfigs } from '@/api/llm'
import type { ChatMessage, ActionPlan, LlmConfig, ChatSession } from '@/types'

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
  const [uploadedImages, setUploadedImages] = useState<string[]>([])
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (!files) return
    Array.from(files).forEach((file) => {
      if (!file.type.startsWith('image/')) {
        antMsg.warning(`${file.name} 不是图片文件`)
        return
      }
      if (file.size > 5 * 1024 * 1024) {
        antMsg.warning(`${file.name} 超过 5MB 限制`)
        return
      }
      const reader = new FileReader()
      reader.onload = () => {
        setUploadedImages((prev) => [...prev, reader.result as string])
      }
      reader.readAsDataURL(file)
    })
    e.target.value = ''
  }

  const removeImage = (idx: number) => {
    setUploadedImages((prev) => prev.filter((_, i) => i !== idx))
  }

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
            if (parsed.type === 'error') {
              appendToLastAssistant(`\n\n> ⚠️ 错误: ${parsed.content}`)
              setStreaming(false)
            } else if (parsed.type === 'delta' && parsed.content) {
              appendToLastAssistant(parsed.content)
            } else if (parsed.text) {
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

  /** F-048: 归档会话 */
  const handleArchive = async (sessionId: number) => {
    try {
      await archiveSession(sessionId)
      antMsg.success('已归档')
      loadSessions()
    } catch {
      antMsg.error('归档失败')
    }
  }

  /** F-048: 删除会话 */
  const handleDeleteSession = async (sessionId: number) => {
    try {
      await deleteSession(sessionId)
      antMsg.success('已删除')
      // 如果当前在删除的会话中，跳转到 /chat
      if (Number(sessionIdParam) === sessionId) {
        navigate('/chat')
      }
      loadSessions()
    } catch {
      antMsg.error('删除失败')
    }
  }

  // 从消息中提取 ActionPlan
  const extractActionPlans = (msg: ChatMessage): ActionPlan[] => {
    if (msg.action_plans) return msg.action_plans
    return []
  }

  const allPlans = messages.flatMap(extractActionPlans)

  // F-048: 分区：活跃会话 + 归档会话
  const activeSessions = sessions.filter((s) => s.status !== 'archived')
  const archivedSessions = sessions.filter((s) => s.status === 'archived')

  const renderSessionItem = (s: ChatSession, isArchived: boolean) => (
    <List.Item
      style={{
        padding: '8px 12px',
        cursor: 'pointer',
        background: Number(sessionIdParam) === s.id ? '#e6f4ff' : 'transparent',
      }}
      onClick={() => navigate(`/chat/sessions/${s.id}`)}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
        <Text ellipsis style={{ fontSize: 13, flex: 1 }}>{s.title || `会话 #${s.id}`}</Text>
        <Space size={2}>
          {!isArchived && (
            <Tooltip title="归档">
              <Button
                type="text"
                size="small"
                icon={<InboxOutlined />}
                style={{ color: '#999' }}
                onClick={(e) => { e.stopPropagation(); handleArchive(s.id) }}
              />
            </Tooltip>
          )}
          <Popconfirm title="确定删除此会话？" onConfirm={(e) => { e?.stopPropagation(); handleDeleteSession(s.id) }}>
            <Button
              type="text"
              size="small"
              icon={<DeleteOutlined />}
              style={{ color: '#ff4d4f' }}
              onClick={(e) => e.stopPropagation()}
            />
          </Popconfirm>
        </Space>
      </div>
    </List.Item>
  )

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
          <div>
            {/* 活跃会话分区 */}
            {activeSessions.length > 0 && (
              <div>
                <div style={{ padding: '8px 12px 4px', fontSize: 11, color: '#999', fontWeight: 600 }}>
                  活跃会话 ({activeSessions.length})
                </div>
                <List
                  size="small"
                  dataSource={activeSessions}
                  renderItem={(s) => renderSessionItem(s, false)}
                />
              </div>
            )}
            {/* 归档会话分区 */}
            {archivedSessions.length > 0 && (
              <div>
                <div style={{ padding: '12px 12px 4px', fontSize: 11, color: '#999', fontWeight: 600, borderTop: '1px solid #f0f0f0' }}>
                  <InboxOutlined style={{ marginRight: 4 }} />
                  归档会话 ({archivedSessions.length})
                </div>
                <List
                  size="small"
                  dataSource={archivedSessions}
                  renderItem={(s) => renderSessionItem(s, true)}
                />
              </div>
            )}
          </div>
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
        <div style={{ padding: '12px 24px', background: '#f7f8fa' }}>
          {/* 整体圆角输入卡片 */}
          <div
            style={{
              background: '#fff',
              border: '1px solid #e8e8ed',
              borderRadius: 16,
              boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
              overflow: 'hidden',
            }}
          >
            {/* 顶部工具栏 */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '8px 14px 0',
              }}
            >
              {/* 左侧工具按钮 */}
              <Space size={4}>
                <Tooltip title="上传图片">
                  <Button
                    type="text"
                    size="small"
                    icon={<PictureOutlined />}
                    style={{ color: '#8c8c8c' }}
                    onClick={() => fileInputRef.current?.click()}
                  />
                </Tooltip>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  multiple
                  style={{ display: 'none' }}
                  onChange={handleImageSelect}
                />
                <Tooltip title="上传附件">
                  <Button
                    type="text"
                    size="small"
                    icon={<PaperClipOutlined />}
                    style={{ color: '#8c8c8c' }}
                    onClick={() => antMsg.info('附件功能即将上线')}
                  />
                </Tooltip>
              </Space>

              {/* 右侧环境+模型 */}
              <Space size={4}>
                <EnvironmentSelector
                  value={environment}
                  onChange={setEnvironment}
                  placeholder="环境"
                  variant="borderless"
                  size="small"
                  style={{ minWidth: 110, fontSize: 12, color: '#666' }}
                />
                <Select
                  value={selectedModel || undefined}
                  onChange={setSelectedModel}
                  placeholder="模型"
                  allowClear
                  variant="borderless"
                  size="small"
                  disabled={models.length === 0}
                  options={models.map((m) => ({ label: m.name, value: m.name }))}
                  style={{ minWidth: 90, fontSize: 12 }}
                  dropdownStyle={{ minWidth: 160 }}
                />
                {models.length === 0 && (
                  <Button
                    type="link"
                    size="small"
                    style={{ color: '#999', padding: 0, fontSize: 12 }}
                    onClick={() => navigate('/llm-config')}
                  >
                    去配置模型
                  </Button>
                )}
              </Space>
            </div>

            {/* 图片预览区 */}
            {uploadedImages.length > 0 && (
              <div style={{ display: 'flex', gap: 8, padding: '8px 14px 0', flexWrap: 'wrap' }}>
                {uploadedImages.map((src, idx) => (
                  <div key={idx} style={{ position: 'relative' }}>
                    <Image
                      src={src}
                      width={64}
                      height={64}
                      style={{ borderRadius: 8, objectFit: 'cover', border: '1px solid #eee' }}
                      preview={{ mask: null }}
                    />
                    <Button
                      type="text"
                      size="small"
                      icon={<CloseCircleOutlined />}
                      danger
                      style={{
                        position: 'absolute',
                        top: -8,
                        right: -8,
                        padding: 0,
                        minWidth: 20,
                        height: 20,
                        background: '#fff',
                        borderRadius: '50%',
                      }}
                      onClick={() => removeImage(idx)}
                    />
                  </div>
                ))}
              </div>
            )}

            {/* 输入框 */}
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
              autoSize={{ minRows: 1, maxRows: 6 }}
              bordered={false}
              style={{
                padding: '10px 14px',
                fontSize: 14,
                resize: 'none',
              }}
            />

            {/* 底部发送栏 */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'flex-end',
                padding: '6px 12px 10px',
                gap: 8,
              }}
            >
              {isStreaming ? (
                <Button
                  icon={<StopOutlined />}
                  danger
                  shape="circle"
                  onClick={disconnect}
                />
              ) : (
                <Button
                  type="primary"
                  shape="circle"
                  icon={<SendOutlined />}
                  disabled={!inputValue.trim() && uploadedImages.length === 0}
                  onClick={sendMessage}
                />
              )}
            </div>
          </div>
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
