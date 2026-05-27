import { Routes, Route, Navigate } from 'react-router-dom'
import { Suspense, lazy } from 'react'
import { Spin } from 'antd'
import AppLayout from '@/components/Layout/AppLayout'
import AuthGuard from '@/components/AuthGuard'
import ErrorBoundary from '@/components/ErrorBoundary'

const Login = lazy(() => import('@/pages/Login'))
const Dashboard = lazy(() => import('@/pages/Dashboard'))
const Chat = lazy(() => import('@/pages/Chat'))
const Environments = lazy(() => import('@/pages/Environments'))
const Policies = lazy(() => import('@/pages/Policies'))
const PolicySimulate = lazy(() => import('@/pages/Policies/Simulate'))
const Audit = lazy(() => import('@/pages/Audit'))
const Plugins = lazy(() => import('@/pages/Plugins'))
const Knowledge = lazy(() => import('@/pages/Knowledge'))
const DataQuery = lazy(() => import('@/pages/DataQuery'))
const Inspections = lazy(() => import('@/pages/Inspections'))
const InspectionReports = lazy(() => import('@/pages/Inspections/Reports'))
const RiskAlerts = lazy(() => import('@/pages/RiskAlerts'))
const Runbooks = lazy(() => import('@/pages/Runbooks'))
const Actions = lazy(() => import('@/pages/Actions'))
const HelpDesk = lazy(() => import('@/pages/HelpDesk'))
const LlmConfig = lazy(() => import('@/pages/LlmConfig'))
// Phase E 新增页面
const Schedules = lazy(() => import('@/pages/Schedules'))
const Notifications = lazy(() => import('@/pages/Notifications'))
const Credentials = lazy(() => import('@/pages/Credentials'))
const SLA = lazy(() => import('@/pages/SLA'))
const Monitoring = lazy(() => import('@/pages/Monitoring'))
const ChangePassword = lazy(() => import('@/pages/ChangePassword'))
const NotFound = lazy(() => import('@/pages/NotFound'))

const PageLoader = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '60vh' }}>
    <Spin size="large" />
  </div>
)

function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            element={
              <AuthGuard>
                <AppLayout />
              </AuthGuard>
            }
          >
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/chat" element={<Chat />} />
            <Route path="/chat/sessions/:id" element={<Chat />} />
            <Route path="/environments" element={<Environments />} />
            <Route path="/policies" element={<Policies />} />
            <Route path="/policies/:id/simulate" element={<PolicySimulate />} />
            <Route path="/audit" element={<Audit />} />
            <Route path="/plugins" element={<Plugins />} />
            <Route path="/knowledge" element={<Knowledge />} />
            <Route path="/data-query" element={<DataQuery />} />
            <Route path="/inspections" element={<Inspections />} />
            <Route path="/inspections/:id/reports" element={<InspectionReports />} />
            <Route path="/risk-alerts" element={<RiskAlerts />} />
            <Route path="/runbooks" element={<Runbooks />} />
            <Route path="/actions" element={<Actions />} />
            <Route path="/helpdesk" element={<HelpDesk />} />
            <Route path="/llm-config" element={<LlmConfig />} />
            {/* Phase E 新增路由 */}
            <Route path="/schedules" element={<Schedules />} />
            <Route path="/notifications" element={<Notifications />} />
            <Route path="/credentials" element={<Credentials />} />
            <Route path="/sla" element={<SLA />} />
            <Route path="/monitoring" element={<Monitoring />} />
            <Route path="/change-password" element={<ChangePassword />} />
          </Route>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Suspense>
    </ErrorBoundary>
  )
}

export default App
