import { useState } from 'react'
import { Layout, Menu } from 'antd'
import {
  DashboardOutlined,
  MessageOutlined,
  EnvironmentOutlined,
  SafetyOutlined,
  AuditOutlined,
  ApiOutlined,
  BookOutlined,
  LineChartOutlined,
  SearchOutlined,
  AlertOutlined,
  FileTextOutlined,
  ClockCircleOutlined,
  CustomerServiceOutlined,
  DoubleLeftOutlined,
  DoubleRightOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import type { MenuProps } from 'antd'

const { Sider } = Layout

type MenuItem = Required<MenuProps>['items'][number]

const coreItems: MenuItem[] = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '概览仪表盘' },
  { key: '/chat', icon: <MessageOutlined />, label: 'AI 对话' },
]

const opsItems: MenuItem[] = [
  { key: '/environments', icon: <EnvironmentOutlined />, label: '环境管理' },
  { key: '/policies', icon: <SafetyOutlined />, label: '安全策略' },
  { key: '/inspections', icon: <SearchOutlined />, label: '巡检任务' },
  { key: '/data-query', icon: <LineChartOutlined />, label: '数据查询' },
  { key: '/runbooks', icon: <FileTextOutlined />, label: '自愈手册' },
  { key: '/plugins', icon: <ApiOutlined />, label: '插件列表' },
]

const secItems: MenuItem[] = [
  { key: '/audit', icon: <AuditOutlined />, label: '审计日志' },
  { key: '/risk-alerts', icon: <AlertOutlined />, label: '高危预警' },
  { key: '/actions', icon: <ClockCircleOutlined />, label: '待确认操作' },
]

const kbItems: MenuItem[] = [
  { key: '/knowledge', icon: <BookOutlined />, label: '知识库' },
  { key: '/helpdesk', icon: <CustomerServiceOutlined />, label: 'HelpDesk' },
]

function Group({
  title,
  children,
  collapsed,
}: {
  title: string
  children: React.ReactNode
  collapsed: boolean
}) {
  if (collapsed) return <>{children}</>
  return (
    <div style={{ marginTop: 8 }}>
      <div
        style={{
          padding: '8px 24px 4px',
          color: 'rgba(255,255,255,0.35)',
          fontSize: 11,
          fontWeight: 600,
          letterSpacing: 1,
          textTransform: 'uppercase',
        }}
      >
        {title}
      </div>
      {children}
    </div>
  )
}

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  const selectedKey = '/' + location.pathname.split('/')[1]

  const menuStyle: React.CSSProperties = {
    background: 'transparent',
    borderInlineEnd: 'none',
    color: 'rgba(255,255,255,0.7)',
  }

  const menuItemStyle = `
    .custom-sider .ant-menu-item {
      border-radius: 8px;
      margin: 2px 12px !important;
      padding: 0 14px !important;
      width: calc(100% - 24px) !important;
      font-size: 14px;
      height: 42px;
      line-height: 42px;
    }
    .custom-sider .ant-menu-item .anticon {
      font-size: 18px;
    }
    .custom-sider .ant-menu-item-selected {
      background: rgba(59, 130, 246, 0.15) !important;
      color: #60a5fa !important;
      box-shadow: inset 3px 0 0 #3b82f6;
    }
    .custom-sider .ant-menu-item:not(.ant-menu-item-selected):hover {
      background: rgba(255,255,255,0.06) !important;
      color: rgba(255,255,255,0.9) !important;
    }
    .custom-sider .ant-menu-item::after {
      display: none;
    }
  `

  return (
    <>
      <style>{menuItemStyle}</style>
      <Sider
        width={220}
        collapsible
        collapsed={collapsed}
        trigger={null}
        className="custom-sider"
        style={{
          background: 'linear-gradient(180deg, #0f172a 0%, #1e293b 100%)',
          borderRight: 'none',
          boxShadow: '4px 0 20px rgba(0,0,0,0.15)',
          position: 'relative',
        }}
      >
        {/* Logo */}
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: collapsed ? 'center' : 'flex-start',
            padding: collapsed ? 0 : '0 20px',
            borderBottom: '1px solid rgba(255,255,255,0.06)',
          }}
        >
          {collapsed ? (
            <span
              style={{
                background: 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                fontWeight: 800,
                fontSize: 20,
              }}
            >
              AI
            </span>
          ) : (
            <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span
                style={{
                  background: 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  fontWeight: 800,
                  fontSize: 22,
                }}
              >
                AIOps
              </span>
              <span
                style={{
                  color: 'rgba(255,255,255,0.3)',
                  fontSize: 11,
                  fontWeight: 500,
                  letterSpacing: 1,
                }}
              >
                v1.0
              </span>
            </span>
          )}
        </div>

        {/* Menu */}
        <div style={{ paddingTop: 8, overflow: 'auto', height: 'calc(100vh - 104px)' }}>
          <Group title="核心" collapsed={collapsed}>
            <Menu
              mode="inline"
              selectedKeys={[selectedKey]}
              items={coreItems}
              onClick={({ key }) => navigate(key)}
              style={menuStyle}
              inlineCollapsed={collapsed}
            />
          </Group>
          <Group title="运维" collapsed={collapsed}>
            <Menu
              mode="inline"
              selectedKeys={[selectedKey]}
              items={opsItems}
              onClick={({ key }) => navigate(key)}
              style={menuStyle}
              inlineCollapsed={collapsed}
            />
          </Group>
          <Group title="安全" collapsed={collapsed}>
            <Menu
              mode="inline"
              selectedKeys={[selectedKey]}
              items={secItems}
              onClick={({ key }) => navigate(key)}
              style={menuStyle}
              inlineCollapsed={collapsed}
            />
          </Group>
          <Group title="知识" collapsed={collapsed}>
            <Menu
              mode="inline"
              selectedKeys={[selectedKey]}
              items={kbItems}
              onClick={({ key }) => navigate(key)}
              style={menuStyle}
              inlineCollapsed={collapsed}
            />
          </Group>
        </div>

        {/* Collapse toggle */}
        <div
          style={{
            height: 40,
            borderTop: '1px solid rgba(255,255,255,0.06)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            color: 'rgba(255,255,255,0.4)',
            transition: 'color 0.2s',
          }}
          onClick={() => setCollapsed(!collapsed)}
          onMouseEnter={(e) => (e.currentTarget.style.color = 'rgba(255,255,255,0.8)')}
          onMouseLeave={(e) => (e.currentTarget.style.color = 'rgba(255,255,255,0.4)')}
        >
          {collapsed ? <DoubleRightOutlined /> : <DoubleLeftOutlined />}
        </div>
      </Sider>
    </>
  )
}
