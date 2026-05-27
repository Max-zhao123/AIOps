import { useState } from 'react'
import { Button, Modal, Input, App, Tag, Space } from 'antd'
import { UndoOutlined, LoadingOutlined } from '@ant-design/icons'
import { rollbackAction } from '@/api/executor'
import type { ExecutionRecord } from '@/types'

interface RollbackButtonProps {
  record: ExecutionRecord
  onSuccess?: () => void
}

/** 回滚操作按钮 + 确认弹窗 */
export default function RollbackButton({ record, onSuccess }: RollbackButtonProps) {
  const [open, setOpen] = useState(false)
  const [reason, setReason] = useState('')
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()

  // 只有成功状态的操作才能回滚
  const canRollback = record.status === 'succeeded' && record.rollback_available !== false

  if (!canRollback) {
    if (record.status === 'rolled_back') {
      return <Tag color="purple">已回滚</Tag>
    }
    if (record.status === 'rolling_back') {
      return <Tag color="processing">回滚中</Tag>
    }
    return null
  }

  const handleRollback = async () => {
    setLoading(true)
    try {
      await rollbackAction(record.id, {
        reason: reason || undefined,
      })
      message.success('回滚已提交')
      setOpen(false)
      setReason('')
      onSuccess?.()
    } catch {
      message.error('回滚失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Button
        size="small"
        danger
        icon={<UndoOutlined />}
        onClick={() => setOpen(true)}
      >
        回滚
      </Button>
      <Modal
        title="确认回滚操作"
        open={open}
        onCancel={() => { setOpen(false); setReason('') }}
        onOk={handleRollback}
        confirmLoading={loading}
        okText="确认回滚"
        okButtonProps={{ danger: true }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <strong>操作摘要：</strong>
            {record.plugin} / {record.action}
          </div>
          <div>
            <strong>环境：</strong>{record.environment}
          </div>
          <div style={{ color: '#ff4d4f', fontSize: 13 }}>
            ⚠️ 回滚操作将尝试撤销已执行的变更，请确认此操作。
          </div>
          <Input.TextArea
            rows={3}
            placeholder="回滚原因（可选）"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
          />
        </Space>
      </Modal>
    </>
  )
}
