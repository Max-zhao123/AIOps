import { useState } from 'react'
import { Form, Input, Button, App, Typography, Card, Alert } from 'antd'
import { LockOutlined } from '@ant-design/icons'
import { changePassword } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import PasswordStrength from '@/components/PasswordStrength'

const { Title, Text } = Typography

/** F-049: 改密页面 + 密码过期提示 */
export default function ChangePassword() {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [newPassword, setNewPassword] = useState('')
  const { message } = App.useApp()
  const user = useAuthStore((s) => s.user)
  const isExpired = user?.password_expired

  const handleSubmit = async (values: { oldPassword: string; newPassword: string; confirmPassword: string }) => {
    if (values.newPassword !== values.confirmPassword) {
      message.error('两次输入的密码不一致')
      return
    }
    setLoading(true)
    try {
      await changePassword(user!.id, {
        oldPassword: values.oldPassword,
        newPassword: values.newPassword,
      })
      message.success('密码修改成功，请重新登录')
      form.resetFields()
      setNewPassword('')
    } catch {
      message.error('密码修改失败，请检查旧密码是否正确')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ maxWidth: 480, margin: '0 auto' }}>
      {isExpired && (
        <Alert
          message="密码已过期"
          description="您的密码已过期，请立即修改密码以继续使用系统。"
          type="error"
          showIcon
          style={{ marginBottom: 24 }}
        />
      )}

      <Card>
        <Title level={4} style={{ textAlign: 'center', marginBottom: 24 }}>修改密码</Title>

        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="oldPassword"
            label="旧密码"
            rules={[{ required: true, message: '请输入旧密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入旧密码"
            />
          </Form.Item>

          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 12, message: '密码长度至少 12 位' },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入新密码"
              onChange={(e) => setNewPassword(e.target.value)}
            />
          </Form.Item>

          <PasswordStrength password={newPassword} />

          <div style={{ margin: '8px 0 24px', fontSize: 12, color: '#999' }}>
            <Text type="secondary">
              密码复杂度要求：至少 12 位，包含大小写字母、数字和特殊字符
            </Text>
          </div>

          <Form.Item
            name="confirmPassword"
            label="确认新密码"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请再次输入新密码"
            />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              确认修改
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
