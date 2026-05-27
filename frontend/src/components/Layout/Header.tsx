import { Layout, Dropdown, Avatar, theme, Space } from 'antd'
import { UserOutlined, LogoutOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { useNavigate } from 'react-router-dom'

const { Header: AntHeader } = Layout

export default function Header() {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()
  const { token } = theme.useToken()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const items = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  return (
    <AntHeader
      style={{
        background: token.colorBgContainer,
        padding: '0 24px',
        display: 'flex',
        justifyContent: 'flex-end',
        alignItems: 'center',
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
      }}
    >
      <Dropdown menu={{ items }}>
        <Space style={{ cursor: 'pointer' }}>
          <Avatar size="small" icon={<UserOutlined />} />
          <span>{user?.username || '用户'}</span>
          <span style={{ color: token.colorTextSecondary, fontSize: 12 }}>
            ({user?.role})
          </span>
        </Space>
      </Dropdown>
    </AntHeader>
  )
}
