import { useState, useEffect, useCallback } from 'react'
import { Card, Row, Col, Select, App, Typography, Table, Tag, Button, Modal, Form, Input, InputNumber, Space, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { listSLADefinitions, createSLADefinition, updateSLADefinition, deleteSLADefinition, getSLAStats } from '@/api/sla'
import type { SLADefinition, SLAStats, CreateSLARequest } from '@/types'
import { useAuthStore } from '@/stores/authStore'
import EnvironmentSelector from '@/components/EnvironmentSelector'

const { Title } = Typography

/** F-045: SLA 仪表盘 */
export default function SLA() {
  const [stats, setStats] = useState<SLAStats | null>(null)
  const [definitions, setDefinitions] = useState<SLADefinition[]>([])
  const [loading, setLoading] = useState(false)
  const [period, setPeriod] = useState('7d')
  const [environment, setEnvironment] = useState<string>()
  const { message } = App.useApp()
  const hasRole = useAuthStore((s) => s.hasRole)
  const isAdmin = hasRole('admin')

  const loadStats = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getSLAStats({ period, environment })
      setStats(data)
    } catch {
      message.error('加载 SLA 统计失败')
    } finally {
      setLoading(false)
    }
  }, [period, environment, message])

  const loadDefinitions = useCallback(async () => {
    try {
      const data = await listSLADefinitions()
      setDefinitions(Array.isArray(data) ? data : [])
    } catch {
      message.error('加载 SLA 定义失败')
    }
  }, [message])

  useEffect(() => { loadStats() }, [loadStats])
  useEffect(() => { loadDefinitions() }, [loadDefinitions])

  // SLA 达标率饼图
  const slaPieOption = {
    tooltip: { trigger: 'item', formatter: '{b}: {c}% ({d}%)' },
    legend: { top: '5%', left: 'center' },
    series: [{
      name: 'SLA 达标率',
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}\n{c}%' },
      data: stats ? [
        { value: Number((stats.overall_sla_rate * 100).toFixed(1)), name: '达标', itemStyle: { color: '#52c41a' } },
        { value: Number(((1 - stats.overall_sla_rate) * 100).toFixed(1)), name: '未达标', itemStyle: { color: '#ff4d4f' } },
      ] : [],
    }],
  }

  // MTTR 趋势折线图（模拟数据，基于当前统计）
  const mttrLineOption = {
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: stats?.severity_stats
        ? Object.keys(stats.severity_stats)
        : [],
    },
    yAxis: { type: 'value', name: 'MTTR (分钟)' },
    series: [{
      name: '平均修复时间',
      type: 'line',
      smooth: true,
      data: stats?.severity_stats
        ? Object.values(stats.severity_stats).map((s) => Number(s.avg_mttr_min.toFixed(1)))
        : [],
      itemStyle: { color: '#1890ff' },
      areaStyle: { color: 'rgba(24,144,255,0.1)' },
    }],
  }

  // 各严重级别 SLA 柱状图
  const severityBarOption = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['响应 SLA 达标率', '解决 SLA 达标率'] },
    xAxis: {
      type: 'category',
      data: stats?.severity_stats ? Object.keys(stats.severity_stats) : [],
    },
    yAxis: { type: 'value', name: '达标率 (%)', max: 100 },
    series: [
      {
        name: '响应 SLA 达标率',
        type: 'bar',
        data: stats?.severity_stats
          ? Object.values(stats.severity_stats).map((s) =>
              s.total > 0 ? Number(((s.response_sla_met / s.total) * 100).toFixed(1)) : 0)
          : [],
        itemStyle: { color: '#1890ff' },
      },
      {
        name: '解决 SLA 达标率',
        type: 'bar',
        data: stats?.severity_stats
          ? Object.values(stats.severity_stats).map((s) =>
              s.total > 0 ? Number(((s.resolution_sla_met / s.total) * 100).toFixed(1)) : 0)
          : [],
        itemStyle: { color: '#52c41a' },
      },
    ],
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>SLA 仪表盘</Title>
        <Space>
          <EnvironmentSelector
            value={environment}
            onChange={setEnvironment}
            placeholder="全部环境"
            style={{ width: 160 }}
          />
          <Select
            value={period}
            onChange={setPeriod}
            style={{ width: 120 }}
            options={[
              { label: '近 7 天', value: '7d' },
              { label: '近 30 天', value: '30d' },
              { label: '近 90 天', value: '90d' },
            ]}
          />
        </Space>
      </div>

      {/* 概览卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 36, fontWeight: 700, color: stats && stats.overall_sla_rate >= 0.9 ? '#52c41a' : '#ff4d4f' }}>
                {stats ? `${(stats.overall_sla_rate * 100).toFixed(1)}%` : '-'}
              </div>
              <div style={{ color: '#666', marginTop: 4 }}>整体 SLA 达标率</div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 36, fontWeight: 700, color: '#1890ff' }}>
                {stats ? `${stats.overall_mttr_min.toFixed(1)} min` : '-'}
              </div>
              <div style={{ color: '#666', marginTop: 4 }}>平均 MTTR</div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 36, fontWeight: 700, color: '#722ed1' }}>
                {definitions.length}
              </div>
              <div style={{ color: '#666', marginTop: 4 }}>SLA 定义数</div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card title="SLA 达标率" size="small">
            <ReactECharts option={slaPieOption} style={{ height: 280 }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="MTTR 趋势" size="small">
            <ReactECharts option={mttrLineOption} style={{ height: 280 }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="各级别 SLA 达标率" size="small">
            <ReactECharts option={severityBarOption} style={{ height: 280 }} />
          </Card>
        </Col>
      </Row>

      {/* SLA 定义配置表 */}
      {isAdmin && <SLADefinitionTable definitions={definitions} onRefresh={loadDefinitions} />}
    </div>
  )
}

/** SLA 定义配置表（仅 admin 可操作） */
function SLADefinitionTable({
  definitions,
  onRefresh,
}: {
  definitions: SLADefinition[]
  onRefresh: () => void
}) {
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<SLADefinition | null>(null)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (record: SLADefinition) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      severity: record.severity,
      response_time_min: record.response_time_min,
      resolution_time_min: record.resolution_time_min,
    })
    setModalOpen(true)
  }

  const handleSubmit = async (values: CreateSLARequest) => {
    try {
      if (editing) {
        await updateSLADefinition(editing.id, values)
        message.success('更新成功')
      } else {
        await createSLADefinition(values)
        message.success('创建成功')
      }
      setModalOpen(false)
      setEditing(null)
      form.resetFields()
      onRefresh()
    } catch {
      message.error(editing ? '更新失败' : '创建失败')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteSLADefinition(id)
      message.success('删除成功')
      onRefresh()
    } catch {
      message.error('删除失败')
    }
  }

  return (
    <>
      <Card
        title="SLA 定义配置"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建定义</Button>}
      >
        <Table
          dataSource={definitions}
          rowKey="id"
          pagination={false}
          columns={[
            { title: '名称', dataIndex: 'name', key: 'name', width: 180 },
            {
              title: '严重级别', dataIndex: 'severity', key: 'severity', width: 120,
              render: (v: string) => {
                const colorMap: Record<string, string> = { critical: 'red', high: 'orange', medium: 'blue', low: 'green' }
                return <Tag color={colorMap[v] || 'default'}>{v}</Tag>
              },
            },
            { title: '响应时间 (min)', dataIndex: 'response_time_min', key: 'response_time_min', width: 140 },
            { title: '解决时间 (min)', dataIndex: 'resolution_time_min', key: 'resolution_time_min', width: 140 },
            {
              title: '操作', key: 'actions', width: 160,
              render: (_: unknown, r: SLADefinition) => (
                <Space>
                  <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(r)}>编辑</Button>
                  <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
                    <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={editing ? '编辑 SLA 定义' : '新建 SLA 定义'}
        open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditing(null); form.resetFields() }}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：P0 事故 SLA" />
          </Form.Item>
          <Form.Item name="severity" label="严重级别" rules={[{ required: true, message: '请选择严重级别' }]}>
            <Select options={[
              { label: 'Critical', value: 'critical' },
              { label: 'High', value: 'high' },
              { label: 'Medium', value: 'medium' },
              { label: 'Low', value: 'low' },
            ]} />
          </Form.Item>
          <Form.Item name="response_time_min" label="响应时间 (分钟)" rules={[{ required: true, message: '请输入响应时间' }]}>
            <InputNumber min={1} style={{ width: '100%' }} placeholder="15" />
          </Form.Item>
          <Form.Item name="resolution_time_min" label="解决时间 (分钟)" rules={[{ required: true, message: '请输入解决时间' }]}>
            <InputNumber min={1} style={{ width: '100%' }} placeholder="60" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
