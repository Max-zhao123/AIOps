import { useState } from 'react'
import { Card, Select, Button, Input, App, Spin, Table, Row, Col, Alert, Typography, Empty, Tabs } from 'antd'
import { SendOutlined } from '@ant-design/icons'
import EnvironmentSelector from '@/components/EnvironmentSelector'
import { queryData } from '@/api/dataQuery'
import type { DataQueryResponse } from '@/types'
import ReactECharts from 'echarts-for-react'

const { Title, Text } = Typography

export default function DataQuery() {
  const [plugin, setPlugin] = useState<string>('kubernetes')
  const [environment, setEnvironment] = useState<string>()
  const [query, setQuery] = useState('')
  const [result, setResult] = useState<DataQueryResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()

  const handleQuery = async () => {
    if (!query.trim()) {
      message.warning('请输入查询内容')
      return
    }
    setLoading(true)
    try {
      const res = await queryData({ plugin, query, environment })
      setResult(res)
    } catch {
      message.error('查询失败')
    } finally {
      setLoading(false)
    }
  }

  const renderResult = () => {
    if (!result) return <Empty description="输入查询并点击发送" />

    const { data, plugin: sourcePlugin } = result

    if (!data || data.length === 0) {
      return <Empty description="无查询结果" />
    }

    if (sourcePlugin === 'prometheus' || sourcePlugin === 'prom') {
      const first = data[0] as Record<string, unknown>
      const values = first?.values as [number, string][] | undefined
      if (values) {
        const option = {
          tooltip: { trigger: 'axis' },
          xAxis: { type: 'time' },
          yAxis: { type: 'value' },
          series: [{
            data: values.map(([ts, val]) => [ts * 1000, parseFloat(val)]),
            type: 'line',
            smooth: true,
            areaStyle: { opacity: 0.15 },
          }],
        }
        return <ReactECharts option={option} style={{ height: 320 }} />
      }
    }

    if (Array.isArray(data) && data.length > 0 && typeof data[0] === 'object') {
      const keys = Object.keys(data[0] as Record<string, unknown>)
      return (
        <Table
          size="small"
          dataSource={data as Record<string, unknown>[]}
          rowKey={(_, i) => String(i)}
          columns={keys.map((k) => ({ title: k, dataIndex: k, key: k, ellipsis: true }))}
          pagination={data.length > 20 ? { pageSize: 20 } : false}
          scroll={{ x: 'max-content' }}
        />
      )
    }

    return <pre style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(data, null, 2)}</pre>
  }

  return (
    <div>
      <Title level={4}>数据查询</Title>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12} align="middle">
          <Col>
            <Select
              value={plugin}
              onChange={setPlugin}
              style={{ width: 160 }}
              options={[
                { label: 'Kubernetes', value: 'kubernetes' },
                { label: 'Prometheus', value: 'prometheus' },
                { label: 'Logs', value: 'logs' },
              ]}
            />
          </Col>
          <Col>
            <EnvironmentSelector
              value={environment}
              onChange={setEnvironment}
              placeholder="选择环境（可选）"
              style={{ width: 180 }}
            />
          </Col>
          <Col flex={1}>
            <Input.Search
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="输入查询语句或自然语言"
              enterButton={<><SendOutlined /> 查询</>}
              onSearch={handleQuery}
              loading={loading}
            />
          </Col>
        </Row>
      </Card>

      {result?.truncated && (
        <Alert
          message={`结果已截断（总行数: ${result.total_rows}，显示前 500 行）`}
          type="warning"
          showIcon
          closable
          style={{ marginBottom: 16 }}
        />
      )}

      <Card size="small">
        {loading ? <Spin style={{ display: 'block', margin: '40px auto' }} /> : renderResult()}
      </Card>
    </div>
  )
}
