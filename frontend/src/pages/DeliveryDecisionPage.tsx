import { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Button, Card, Col, Divider, Form, InputNumber, Row, Select, Space, Statistic, Table, Typography, message } from 'antd';
import { ReloadOutlined, SaveOutlined } from '@ant-design/icons';
import { calculateDeliveryDecision, getDeliveryDecisionConfig, getDeliveryDecisionDefaults, saveDeliveryDecisionConfig } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { DeliveryDecisionInput, DeliveryDecisionResult, DeliveryDecisionSensitivityPoint } from '../types';

const { Text, Title } = Typography;

const fmt = (value?: number | null, digits = 4) => {
  if (value === null || value === undefined || Number.isNaN(value)) return '-';
  return new Intl.NumberFormat('zh-CN', { maximumFractionDigits: digits, minimumFractionDigits: digits }).format(value);
};

const pct = (value?: number | null, digits = 1) => {
  if (value === null || value === undefined || Number.isNaN(value)) return '-';
  return `${(value * 100).toFixed(digits)}%`;
};

const cur = (value?: number | null, digits = 4) => `${fmt(value, digits)} 元`;

const defaults: DeliveryDecisionInput = {
  hw_total: 277000,
  hw_cores: 128,
  hw_tax_rate: 0.13,
  idc_rent_monthly: 4639,
  idc_rack_kw: 3.52,
  idc_fill_rate: 1.2,
  idc_server_power_w: 570,
  idc_network_depreciation: 3.62,
  cloud_memory_ratio: 8,
  cloud_disk_ratio: 0,
  cloud_cpu_daily_price: 0.78543,
  cloud_memory_daily_price: 0.1122,
  cloud_disk_daily_price: 0.00732,
  cloud_tax_rate: 0.06,
  depreciation_years: 7,
  wacc_rate: 0.03,
  residual_rate: 0.05,
  country: 'China',
  currency: 'CNY',
  cloud_current_discount: 0.25
};

type CurveRow = DeliveryDecisionSensitivityPoint & { key: string };

function SimpleLineChart({ title, series, yLabel }: { title: string; series: DeliveryDecisionSensitivityPoint[]; yLabel: string }) {
  const width = 560;
  const height = 220;
  const padding = 28;
  const points = series.map((item, index) => ({
    x: padding + (index * (width - padding * 2)) / Math.max(series.length - 1, 1),
    y: item.final_self_share
  }));
  const path = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${height - padding - p.y * (height - padding * 2)}`).join(' ');
  return (
    <Card title={title} type="inner">
      <svg width="100%" viewBox={`0 0 ${width} ${height}`} role="img" aria-label={title}>
        <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} stroke="#d9d9d9" />
        <line x1={padding} y1={padding} x2={padding} y2={height - padding} stroke="#d9d9d9" />
        <path d={path} fill="none" stroke="#1677ff" strokeWidth="3" />
        {points.map((p, idx) => (
          <g key={series[idx].label}>
            <circle cx={p.x} cy={height - padding - p.y * (height - padding * 2)} r="4" fill="#1677ff" />
            <text x={p.x} y={height - 8} fontSize="10" textAnchor="middle" fill="#666">
              {series[idx].label}
            </text>
          </g>
        ))}
        <text x={8} y={16} fontSize="10" fill="#666">{yLabel}</text>
      </svg>
    </Card>
  );
}

export default function DeliveryDecisionPage() {
  const [form] = Form.useForm<DeliveryDecisionInput>();
  const [input, setInput] = useState<DeliveryDecisionInput>(defaults);
  const [result, setResult] = useState<DeliveryDecisionResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [initLoading, setInitLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const debounceRef = useRef<number | null>(null);

  useEffect(() => {
    void loadConfig();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function loadConfig() {
    setInitLoading(true);
    try {
      const resp = ensureApiOk(await getDeliveryDecisionConfig());
      if (resp.data.found && resp.data.state?.input) {
        const next = resp.data.state.input;
        setInput(next);
        form.setFieldsValue(next);
        await runCalculate(next);
        return;
      }
      await loadDefaults(defaults.country);
    } catch (e) {
      message.error(parseApiError(e, '加载交付决策配置失败'));
      await loadDefaults(defaults.country);
    } finally {
      setInitLoading(false);
    }
  }

  async function loadDefaults(country: string) {
    setInitLoading(true);
    try {
      const resp = ensureApiOk(await getDeliveryDecisionDefaults(country));
      const next = resp.data.input;
      setInput(next);
      form.setFieldsValue(next);
      await runCalculate(next);
    } catch (e) {
      message.error(parseApiError(e, '加载交付决策默认参数失败'));
    } finally {
      setInitLoading(false);
    }
  }

  async function runCalculate(next: DeliveryDecisionInput) {
    setLoading(true);
    try {
      const resp = ensureApiOk(await calculateDeliveryDecision({ input: next }));
      setResult(resp.data);
    } catch (e) {
      setResult(null);
      message.error(parseApiError(e, '交付方式计算失败'));
    } finally {
      setLoading(false);
    }
  }

  function scheduleCalculate(next: DeliveryDecisionInput) {
    setInput(next);
    form.setFieldsValue(next);
    if (debounceRef.current) {
      window.clearTimeout(debounceRef.current);
    }
    debounceRef.current = window.setTimeout(() => {
      void runCalculate(next);
    }, 220);
  }

  async function handleSave() {
    setSaving(true);
    try {
      const resp = ensureApiOk(await saveDeliveryDecisionConfig({ input }));
      const next = resp.data.state.input;
      setInput(next);
      form.setFieldsValue(next);
      await runCalculate(next);
      message.success('配置已保存');
    } catch (e) {
      message.error(parseApiError(e, '保存配置失败'));
    } finally {
      setSaving(false);
    }
  }

  const hardwarePoints = useMemo(
    () => result?.sensitivity_points.filter((p) => p.curve === 'hardware_price') || [],
    [result]
  );
  const cloudPoints = useMemo(
    () => result?.sensitivity_points.filter((p) => p.curve === 'cloud_discount') || [],
    [result]
  );

  const pointColumns = [
    { title: '采样', dataIndex: 'label', width: 88 },
    { title: '输入值', dataIndex: 'x_value', render: (v: number) => fmt(v, 4) },
    { title: '公式占比', dataIndex: 'formula_self_share', render: (v: number) => pct(v, 1) },
    { title: '最终占比', dataIndex: 'final_self_share', render: (v: number) => pct(v, 1) },
    { title: '3年锁定 TCO', dataIndex: 'self_daily_tco_3y', render: (v: number) => cur(v, 6) },
    { title: '对冲失效', dataIndex: 'cloud_hedge_lost', render: (v: boolean) => (v ? '是' : '否') }
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card>
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
            <div>
              <Title level={4} style={{ margin: 0 }}>交付方式决策</Title>
              <Text type="secondary">自建与公有云成本对标、动态配比与价格敏感度分析</Text>
            </div>
            <Space>
              <Button icon={<SaveOutlined />} type="primary" onClick={handleSave} loading={saving}>
                保存配置
              </Button>
              <Button icon={<ReloadOutlined />} onClick={() => loadDefaults(input.country)} loading={initLoading}>
                重置默认值
              </Button>
            </Space>
          </Space>
          <Alert
            type="info"
            showIcon
            message="说明"
            description="国家切换只影响默认模板，币种保持为元；本页结果按当前输入实时计算，点击保存配置后写入后端持久化。"
          />
        </Space>
      </Card>

      <Row gutter={16}>
        <Col xs={24} xl={9}>
          <Card title="参数输入">
            <Form form={form} layout="vertical" initialValues={input}>
              <Form.Item label="国家">
                <Select
                  value={input.country}
                  options={[{ value: 'China', label: 'China' }, { value: 'India', label: 'India' }]}
                  onChange={(value) => void loadDefaults(value)}
                />
              </Form.Item>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="整机含税采购总价"><InputNumber value={input.hw_total} onChange={(v) => scheduleCalculate({ ...input, hw_total: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={2} /></Form.Item></Col>
                <Col span={12}><Form.Item label="整机物理核心数"><InputNumber value={input.hw_cores} onChange={(v) => scheduleCalculate({ ...input, hw_cores: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={0} /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="硬件增值税率"><InputNumber value={input.hw_tax_rate * 100} onChange={(v) => scheduleCalculate({ ...input, hw_tax_rate: Number(v || 0) / 100 })} style={{ width: '100%' }} min={0} max={18} precision={2} addonAfter="%" /></Form.Item></Col>
                <Col span={12}><Form.Item label="折旧年限"><InputNumber value={input.depreciation_years} onChange={(v) => scheduleCalculate({ ...input, depreciation_years: Number(v || 0) })} style={{ width: '100%' }} min={1} max={10} precision={0} /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="机柜月租"><InputNumber value={input.idc_rent_monthly} onChange={(v) => scheduleCalculate({ ...input, idc_rent_monthly: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={2} /></Form.Item></Col>
                <Col span={12}><Form.Item label="机柜功率 (kW)"><InputNumber value={input.idc_rack_kw} onChange={(v) => scheduleCalculate({ ...input, idc_rack_kw: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={2} /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="满柜率"><InputNumber value={input.idc_fill_rate * 100} onChange={(v) => scheduleCalculate({ ...input, idc_fill_rate: Number(v || 0) / 100 })} style={{ width: '100%' }} min={50} max={150} precision={2} addonAfter="%" /></Form.Item></Col>
                <Col span={12}><Form.Item label="单机功率 (W)"><InputNumber value={input.idc_server_power_w} onChange={(v) => scheduleCalculate({ ...input, idc_server_power_w: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={0} /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="网络及机柜月折旧"><InputNumber value={input.idc_network_depreciation} onChange={(v) => scheduleCalculate({ ...input, idc_network_depreciation: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={2} /></Form.Item></Col>
                <Col span={12}><Form.Item label="云服务进项税率"><InputNumber value={input.cloud_tax_rate * 100} onChange={(v) => scheduleCalculate({ ...input, cloud_tax_rate: Number(v || 0) / 100 })} style={{ width: '100%' }} min={0} max={18} precision={2} addonAfter="%" /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="内存比 (GB/Core)"><InputNumber value={input.cloud_memory_ratio} onChange={(v) => scheduleCalculate({ ...input, cloud_memory_ratio: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={2} /></Form.Item></Col>
                <Col span={12}><Form.Item label="磁盘比 (GB/Core)"><InputNumber value={input.cloud_disk_ratio} onChange={(v) => scheduleCalculate({ ...input, cloud_disk_ratio: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={2} /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="CPU 单核日价"><InputNumber value={input.cloud_cpu_daily_price} onChange={(v) => scheduleCalculate({ ...input, cloud_cpu_daily_price: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={5} /></Form.Item></Col>
                <Col span={12}><Form.Item label="内存日价"><InputNumber value={input.cloud_memory_daily_price} onChange={(v) => scheduleCalculate({ ...input, cloud_memory_daily_price: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={5} /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="磁盘日价"><InputNumber value={input.cloud_disk_daily_price} onChange={(v) => scheduleCalculate({ ...input, cloud_disk_daily_price: Number(v || 0) })} style={{ width: '100%' }} min={0} precision={5} /></Form.Item></Col>
                <Col span={12}><Form.Item label="WACC 年资金成本率"><InputNumber value={input.wacc_rate * 100} onChange={(v) => scheduleCalculate({ ...input, wacc_rate: Number(v || 0) / 100 })} style={{ width: '100%' }} min={0} max={100} precision={2} addonAfter="%" /></Form.Item></Col>
              </Row>
              <Row gutter={12}>
                <Col span={12}><Form.Item label="设备残值率"><InputNumber value={input.residual_rate * 100} onChange={(v) => scheduleCalculate({ ...input, residual_rate: Number(v || 0) / 100 })} style={{ width: '100%' }} min={0} max={100} precision={2} addonAfter="%" /></Form.Item></Col>
                <Col span={12}><Form.Item label="当前云折扣"><InputNumber value={input.cloud_current_discount * 100} onChange={(v) => scheduleCalculate({ ...input, cloud_current_discount: Number(v || 0) / 100 })} style={{ width: '100%' }} min={1} max={100} precision={2} addonAfter="折" /></Form.Item></Col>
              </Row>
            </Form>
          </Card>
        </Col>

        <Col xs={24} xl={15}>
          <Row gutter={[16, 16]}>
            <Col xs={24} md={12} xl={8}>
              <Card><Statistic title="公有云单核日 Net TCO" value={result?.formula.cloud_daily_net} precision={6} suffix="元" loading={loading} /></Card>
            </Col>
            <Col xs={24} md={12} xl={8}>
              <Card><Statistic title="自建单核日 TCO" value={result?.formula.self_daily_tco} precision={6} suffix="元" loading={loading} /></Card>
            </Col>
            <Col xs={24} md={12} xl={8}>
              <Card><Statistic title="3年锁定 TCO" value={result?.formula.self_daily_tco_3y} precision={6} suffix="元" loading={loading} /></Card>
            </Col>
            <Col xs={24} md={12} xl={8}>
              <Card><Statistic title="溢价比值 R" value={result?.formula.premium_ratio_r} precision={6} loading={loading} /></Card>
            </Col>
            <Col xs={24} md={12} xl={8}>
              <Card><Statistic title="最终自建占比" value={result ? result.formula.final_self_share * 100 : undefined} precision={2} suffix="%" loading={loading} /></Card>
            </Col>
            <Col xs={24} md={12} xl={8}>
              <Card><Statistic title="参考打平年限" value={result?.formula.break_even_years || undefined} precision={3} suffix="年" loading={loading} /></Card>
            </Col>
          </Row>

          <Card title="关键结论" style={{ marginTop: 16 }}>
            <Space direction="vertical" size={8} style={{ width: '100%' }}>
              <Text>云 TCO：{cur(result?.formula.cloud_daily_net, 6)}</Text>
              <Text>自建 TCO：{cur(result?.formula.self_daily_tco, 6)}</Text>
              <Text>3年锁定 TCO：{cur(result?.formula.self_daily_tco_3y, 6)}</Text>
              <Text>公式配比：自建 {pct(result?.formula.formula_self_share, 2)} / 公有云 {pct(result?.formula.formula_cloud_share, 2)}</Text>
              <Text>最终配比：自建 {pct(result?.formula.final_self_share, 2)} / 公有云 {pct(result?.formula.final_cloud_share, 2)}</Text>
              <Text>对冲判断：{result?.formula.cloud_hedge_lost ? '公有云失去对冲优势，全量自建' : '仍保留动态配比'}</Text>
              <Text type="secondary">算式版本：{result?.snapshot.formula_version || '-'}</Text>
            </Space>
          </Card>
        </Col>
      </Row>

      <Divider style={{ margin: '8px 0' }} />

      <Row gutter={16}>
        <Col xs={24} xl={12}>
          <SimpleLineChart title="自建价格敏感度" series={hardwarePoints} yLabel="最终自建占比" />
        </Col>
        <Col xs={24} xl={12}>
          <SimpleLineChart title="公有云价格敏感度" series={cloudPoints} yLabel="最终自建占比" />
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} xl={12}>
          <Card title="自建价格敏感度采样">
            <Table<CurveRow>
              rowKey={(r) => `${r.curve}-${r.label}`}
              pagination={false}
              dataSource={hardwarePoints as CurveRow[]}
              columns={pointColumns as any}
            />
          </Card>
        </Col>
        <Col xs={24} xl={12}>
          <Card title="公有云价格敏感度采样">
            <Table<CurveRow>
              rowKey={(r) => `${r.curve}-${r.label}`}
              pagination={false}
              dataSource={cloudPoints as CurveRow[]}
              columns={pointColumns as any}
            />
          </Card>
        </Col>
      </Row>

      {result?.formula.cloud_hedge_lost ? (
        <Alert type="warning" showIcon message="公有云失去对冲优势，最终建议为全量自建。" />
      ) : null}
    </Space>
  );
}
