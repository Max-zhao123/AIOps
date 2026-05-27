import { useState } from 'react'
import { Card, Button, App, Typography, Row, Col, Tag, Space, Descriptions } from 'antd'
import { ExperimentOutlined, ArrowLeftOutlined } from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { simulatePolicy } from '@/api/policies'
import type { PolicyEvaluateResult } from '@/types'
import Editor from '@monaco-editor/react'

const { Title, Text } = Typography

const DECISION_STYLE: Record<string, { color: string; bg: string }> = {
  ALLOW: { color: '#fff', bg: '#52c41a' },
  ASK: { color: '#fff', bg: '#fa8c16' },
  DENY: { color: '#fff', bg: '#ff4d4f' },
}

const DEFAULT_JSON = `{
  "plugin": "kubernetes",
  "action": "list",
  "parameters": {
    "namespace": "default",
    "resource": "pods"
  },
  "risk": "read",
  "summary": "查看 default 命名空间的 Pod 列表",
  "commandPreview": "kubectl get pods -n default"
}
`

export default function Simulate() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [jsonValue, setJsonValue] = useState(DEFAULT_JSON)
  const [result, setResult] = useState<PolicyEvaluateResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { message } = App.useApp()

  const handleSimulate = async () => {
    setError('')
    setResult(null)
    try {
      const planData = JSON.parse(jsonValue)
      setLoading(true)
      const res = await simulatePolicy(Number(id), {
        plan: planData,
        securitySpecYAML: '',
      })
      setResult(res)
    } catch (err) {
      if (err instanceof SyntaxError) {
        setError('JSON 格式错误，请检查')
      } else {
        setError('模拟评估失败')
      }
      message.error('模拟评估失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/policies')}>返回</Button>
          <Title level={4} style={{ margin: 0 }}>策略模拟器</Title>
          {id && <Tag>策略 #{id}</Tag>}
        </Space>
      </div>

      <Row gutter={16}>
        <Col xs={24} lg={12}>
          <Card
            title="ActionPlan 输入"
            size="small"
            extra={
              <Button
                type="primary"
                size="small"
                icon={<ExperimentOutlined />}
                onClick={handleSimulate}
                loading={loading}
              >
                运行模拟
              </Button>
            }
            style={{ marginBottom: 16 }}
          >
            <div style={{ border: '1px solid #d9d9d9', borderRadius: 6, overflow: 'hidden' }}>
              <Editor
                height="420px"
                language="json"
                value={jsonValue}
                onChange={(val) => setJsonValue(val || '')}
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
            {error && (
              <Text type="danger" style={{ marginTop: 8, display: 'block' }}>{error}</Text>
            )}
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="评估结果" size="small">
            {result ? (
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <div
                  style={{
                    padding: '16px 24px',
                    borderRadius: 8,
                    background: DECISION_STYLE[result.decision]?.bg || '#f5f5f5',
                    color: DECISION_STYLE[result.decision]?.color || '#333',
                    textAlign: 'center',
                    fontSize: 20,
                    fontWeight: 700,
                  }}
                >
                  {result.decision}
                </div>
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="匹配规则">
                    {result.matched_rule_id ? (
                      <Tag color="blue">{result.matched_rule_id}</Tag>
                    ) : (
                      <Tag>default</Tag>
                    )}
                  </Descriptions.Item>
                  <Descriptions.Item label="说明">
                    <Text>{result.message || '-'}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="风险级别">
                    <Tag color={result.decision === 'DENY' ? 'red' : result.decision === 'ASK' ? 'orange' : 'green'}>
                      {result.decision === 'ALLOW' ? '低风险' : result.decision === 'ASK' ? '需确认' : '已拒绝'}
                    </Tag>
                  </Descriptions.Item>
                </Descriptions>
              </Space>
            ) : (
              <div style={{ textAlign: 'center', padding: 60, color: '#999' }}>
                <ExperimentOutlined style={{ fontSize: 40, marginBottom: 16 }} />
                <div>输入 ActionPlan JSON，点击"运行模拟"查看策略评估结果</div>
              </div>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  )
}
