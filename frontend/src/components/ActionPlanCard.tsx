import { Card, Tag, Space } from 'antd'
import {
  CodeOutlined,
  SafetyOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import type { ActionPlan } from '@/types'

const riskColor: Record<string, string> = {
  read: 'blue',
  write: 'orange',
  notify: 'default',
}

const riskLabel: Record<string, string> = {
  read: '只读',
  write: '写入',
  notify: '通知',
}

interface Props {
  plan: ActionPlan
}

export default function ActionPlanCard({ plan }: Props) {
  return (
    <Card
      size="small"
      style={{
        borderLeft: `4px solid ${plan.risk === 'write' ? '#fa8c16' : plan.risk === 'read' ? '#1677ff' : '#d9d9d9'}`,
        background: '#fafafa',
        marginBottom: 8,
      }}
    >
      <Space direction="vertical" size={4} style={{ width: '100%' }}>
        <Space>
          <CodeOutlined />
          <strong>{plan.plugin}</strong>
          <span style={{ color: '#666' }}>/ {plan.action}</span>
          {plan.risk && (
            <Tag color={riskColor[plan.risk] || 'default'}>
              <SafetyOutlined /> {riskLabel[plan.risk] || plan.risk}
            </Tag>
          )}
        </Space>
        {plan.commandPreview && (
          <div
            style={{
              fontFamily: 'monospace',
              fontSize: 13,
              padding: '6px 10px',
              background: '#1e1e1e',
              color: '#d4d4d4',
              borderRadius: 6,
              overflowX: 'auto',
            }}
          >
            $ {plan.commandPreview}
          </div>
        )}
        {plan.summary && (
          <div style={{ color: '#666', fontSize: 13 }}>
            <ThunderboltOutlined style={{ marginRight: 4 }} />
            {plan.summary}
          </div>
        )}
      </Space>
    </Card>
  )
}
