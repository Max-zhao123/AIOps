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
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import type { MenuProps } from 'antd'

const { Sider } = Layout

type MenuItem = Required<MenuProps>['items'][number]

const menuItems: MenuItem[] = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '概览仪表盘' },
  { key: '/chat', icon: <MessageOutlined />, label: 'AI 对话' },
  { key: '/environments', icon: <EnvironmentOutlined />, label: '环境管理' },
  { key: '/policies', icon: <SafetyOutlined />, label: '安全策略' },
  { key: '/audit', icon: <AuditOutlined />, label: '审计日志' },
  { key: '/plugins', icon: <ApiOutlined />, label: '插件列表' },
  { key: '/knowledge', icon: <BookOutlined />, label: '知识库' },
  { key: '/data-query', icon: <LineChartOutlined />, label: '数据查询' },
  { key: '/inspections', icon: <SearchOutlined />, label: '巡检任务' },
  { key: '/risk-alerts', icon: <AlertOutlined />, label: '高危预警' },
  { key: '/runbooks', icon: <FileTextOutlined />, label: '自愈手册' },
  { key: '/actions', icon: <ClockCircleOutlined />, label: '待确认操作' },
  { key: '/helpdesk', icon: <CustomerServiceOutlined />, label: 'HelpDesk' },
]

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  const selectedKey = '/' + location.pathname.split('/')[1]

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={setCollapsed}
      style={{ borderRight: '1px solid #f0f0f0' }}
    >
      <div
        style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderBottom: '1px solid #f0f0f0',
        }}
      >
        <span
          style={{
            color: '#1677ff',
            fontWeight: 700,
            fontSize: collapsed ? 16 : 20,
            whiteSpace: 'nowrap',
          }}
        >
          {collapsed ? 'AI' : 'AIOps'}
        </span>
      </div>
      <Menu
        mode="inline"
        selectedKeys={[selectedKey]}
        items={menuItems}
        onClick={({ key }) => navigate(key)}
        style={{ borderInlineEnd: 'none' }}
      />
    </Sider>
  )
}
