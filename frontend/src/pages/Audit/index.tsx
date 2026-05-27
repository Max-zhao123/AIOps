import { useState, useEffect, useCallback } from 'react'
import { Table, Select, Input, App, Typography, Tag, Space, Button, Dropdown, Modal, Card, Row, Col } from 'antd'
import { SearchOutlined, DownloadOutlined, FileTextOutlined } from '@ant-design/icons'
import { listAudit, exportAudit, getComplianceReport } from '@/api/audit'
import EnvironmentSelector from '@/components/EnvironmentSelector'
import type { AuditLog, ComplianceReport } from '@/types'

const { Title, Text } = Typography

/** F-047: 审计导出 + 合规报告 */
export default function Audit() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(false)
  const [environment, setEnvironment] = useState<string>()
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [exportLoading, setExportLoading] = useState(false)
  const [reportOpen, setReportOpen] = useState(false)
  const [report, setReport] = useState<ComplianceReport | null>(null)
  const [reportLoading, setReportLoading] = useState(false)
  const { message } = App.useApp()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listAudit({ environment, page, pageSize })
      setLogs(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [environment, page, pageSize, message])

  useEffect(() => { load() }, [load])

  /** 导出审计日志 */
  const handleExport = async (format: 'csv' | 'pdf') => {
    setExportLoading(true)
    try {
      const blob = await exportAudit({ format, environment })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `audit_export.${format}`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
      message.success(`导出 ${format.toUpperCase()} 成功`)
    } catch {
      message.error('导出失败')
    } finally {
      setExportLoading(false)
    }
  }

  /** 查看合规报告 */
  const handleComplianceReport = async () => {
    setReportLoading(true)
    try {
      const data = await getComplianceReport({ environment })
      setReport(data)
      setReportOpen(true)
    } catch {
      message.error('获取合规报告失败')
    } finally {
      setReportLoading(false)
    }
  }

  return (
    <div>
      <Title level={4}>审计日志</Title>

      <Space style={{ marginBottom: 16 }} wrap>
        <EnvironmentSelector
          value={environment}
          onChange={(v) => { setEnvironment(v); setPage(1) }}
          placeholder="全部环境"
          style={{ width: 160 }}
        />
        <Input
          prefix={<SearchOutlined />}
          placeholder="搜索关键词"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          style={{ width: 200 }}
          allowClear
        />
        {/* F-047: 导出按钮 */}
        <Dropdown
          menu={{
            items: [
              { key: 'csv', label: '导出 CSV', onClick: () => handleExport('csv') },
              { key: 'pdf', label: '导出 PDF', onClick: () => handleExport('pdf') },
            ],
          }}
        >
          <Button icon={<DownloadOutlined />} loading={exportLoading}>
            导出
          </Button>
        </Dropdown>
        {/* F-047: 合规报告按钮 */}
        <Button
          icon={<FileTextOutlined />}
          onClick={handleComplianceReport}
          loading={reportLoading}
        >
          合规报告
        </Button>
      </Space>

      <Table
        dataSource={logs.filter((l) => !keyword || JSON.stringify(l).includes(keyword))}
        rowKey="id"
        loading={loading}
        pagination={{ current: page, pageSize, onChange: setPage, showTotal: (t) => `共 ${t} 条` }}
        columns={[
          { title: '时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
          { title: '用户', dataIndex: 'username', key: 'username', width: 100 },
          { title: '操作', dataIndex: 'action', key: 'action', width: 120 },
          { title: '资源', dataIndex: 'resource_type', key: 'resource_type', width: 100 },
          { title: '资源ID', dataIndex: 'resource_id', key: 'resource_id', width: 80 },
          {
            title: '环境',
            dataIndex: 'environment',
            key: 'environment',
            width: 120,
            render: (v: string) => v ? <Tag>{v}</Tag> : '-',
          },
          {
            title: '结果',
            dataIndex: 'result',
            key: 'result',
            width: 80,
            render: (v: string) => (
              <Tag color={v === 'success' ? 'green' : v === 'denied' ? 'red' : 'default'}>
                {v}
              </Tag>
            ),
          },
          { title: '详情', dataIndex: 'detail', key: 'detail', ellipsis: true },
        ]}
      />

      {/* 合规报告 Modal */}
      <Modal
        title="合规报告"
        open={reportOpen}
        onCancel={() => setReportOpen(false)}
        footer={null}
        width={700}
      >
        {report ? (
          <div>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small">
                  <Text type="secondary">统计周期</Text>
                  <div style={{ fontSize: 18, fontWeight: 600 }}>{report.period || '-'}</div>
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small">
                  <Text type="secondary">事件总数</Text>
                  <div style={{ fontSize: 18, fontWeight: 600 }}>{report.total_events}</div>
                </Card>
              </Col>
            </Row>

            {report.by_action && Object.keys(report.by_action).length > 0 && (
              <Card title="按操作类型" size="small" style={{ marginTop: 16 }}>
                {Object.entries(report.by_action).map(([key, val]) => (
                  <div key={key} style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <Text>{key}</Text>
                    <Text strong>{val}</Text>
                  </div>
                ))}
              </Card>
            )}

            {report.by_result && Object.keys(report.by_result).length > 0 && (
              <Card title="按结果" size="small" style={{ marginTop: 16 }}>
                {Object.entries(report.by_result).map(([key, val]) => (
                  <div key={key} style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <Tag color={key === 'success' ? 'green' : key === 'denied' ? 'red' : 'default'}>{key}</Tag>
                    <Text strong>{val}</Text>
                  </div>
                ))}
              </Card>
            )}

            {report.by_environment && Object.keys(report.by_environment).length > 0 && (
              <Card title="按环境" size="small" style={{ marginTop: 16 }}>
                {Object.entries(report.by_environment).map(([key, val]) => (
                  <div key={key} style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <Tag>{key}</Tag>
                    <Text strong>{val}</Text>
                  </div>
                ))}
              </Card>
            )}
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>暂无数据</div>
        )}
      </Modal>
    </div>
  )
}
