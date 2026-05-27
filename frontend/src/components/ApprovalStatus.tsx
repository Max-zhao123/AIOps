import { Tag, Tooltip, Space } from 'antd'
import { ClockCircleOutlined, CheckCircleOutlined, CloseCircleOutlined, MinusCircleOutlined } from '@ant-design/icons'
import type { Approver } from '@/types'

interface ApprovalStatusProps {
  approvers: Approver[]
  timeoutAt?: string
}

/** 审批状态标签组件：展示多人审批状态 + 超时倒计时 */
export default function ApprovalStatus({ approvers, timeoutAt }: ApprovalStatusProps) {
  const total = approvers.length
  const approved = approvers.filter((a) => a.status === 'approved').length
  const rejected = approvers.filter((a) => a.status === 'rejected').length
  const pending = approvers.filter((a) => a.status === 'pending').length

  // 整体状态判断
  let overallStatus: 'pending' | 'partially_approved' | 'approved' | 'rejected'
  if (rejected > 0) {
    overallStatus = 'rejected'
  } else if (approved === total) {
    overallStatus = 'approved'
  } else if (approved > 0) {
    overallStatus = 'partially_approved'
  } else {
    overallStatus = 'pending'
  }

  const statusConfig: Record<string, { color: string; icon: React.ReactNode; text: string }> = {
    pending: { color: 'default', icon: <ClockCircleOutlined />, text: '待审批' },
    partially_approved: { color: 'processing', icon: <MinusCircleOutlined />, text: `部分审批 ${approved}/${total}` },
    approved: { color: 'success', icon: <CheckCircleOutlined />, text: '已通过' },
    rejected: { color: 'error', icon: <CloseCircleOutlined />, text: '已拒绝' },
  }

  const config = statusConfig[overallStatus]

  // 超时倒计时
  const isTimeoutWarning = (() => {
    if (!timeoutAt) return false
    const remaining = new Date(timeoutAt).getTime() - Date.now()
    return remaining > 0 && remaining < 30 * 60 * 1000 // 30 分钟内即将超时
  })()

  const isTimeout = (() => {
    if (!timeoutAt) return false
    return new Date(timeoutAt).getTime() < Date.now()
  })()

  return (
    <Space direction="vertical" size={4}>
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
      {isTimeout && (
        <Tag color="error">审批已超时</Tag>
      )}
      {isTimeoutWarning && !isTimeout && (
        <Tooltip title={`审批截止: ${new Date(timeoutAt!).toLocaleString()}`}>
          <Tag color="warning">即将超时</Tag>
        </Tooltip>
      )}
      {approvers.length > 0 && (
        <div style={{ fontSize: 12, color: '#999' }}>
          {approvers.map((a, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <Tag
                color={a.status === 'approved' ? 'green' : a.status === 'rejected' ? 'red' : 'default'}
                style={{ fontSize: 11, padding: '0 4px', margin: 0 }}
              >
                {a.username || a.user_id}
              </Tag>
            </div>
          ))}
        </div>
      )}
    </Space>
  )
}
