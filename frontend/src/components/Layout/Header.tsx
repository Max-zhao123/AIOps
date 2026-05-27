import { Layout, Dropdown, Avatar, theme, Space, Badge } from 'antd'
import { UserOutlined, LogoutOutlined, BellOutlined } from '@ant-design/icons'
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
        borderBottom: 'none',
        boxShadow: '0 1px 4px rgba(0,0,0,0.04)',
        zIndex: 10,
      }}
    >
      <Space size={20}>
        <Badge size="small">
          <BellOutlined style={{ fontSize: 18, color: token.colorTextSecondary, cursor: 'pointer' }} />
        </Badge>
        <Dropdown menu={{ items }} placement="bottomRight">
          <Space
            style={{
              cursor: 'pointer',
              padding: '4px 8px',
              borderRadius: 8,
              transition: 'background 0.2s',
            }}
            onMouseEnter={(e) => (e.currentTarget.style.background = token.colorFillSecondary)}
            onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
          >
            <Avatar size="small" icon={<UserOutlined />} style={{ background: '#3b82f6' }} />
            <span style={{ fontWeight: 500 }}>{user?.username || '用户'}</span>
            <span
              style={{
                color: token.colorTextSecondary,
                fontSize: 12,
                background: token.colorFillTertiary,
                padding: '0 8px',
                borderRadius: 4,
              }}
            >
              {user?.role}
            </span>
          </Space>
        </Dropdown>
      </Space>
    </AntHeader>
  )
}
