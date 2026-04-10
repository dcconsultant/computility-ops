import { useEffect, useMemo, useState } from 'react';
import { Button, Card, Col, Progress, Row, Space, Table, Tag, Typography, message } from 'antd';
import { ExpandOutlined, ReloadOutlined, ShrinkOutlined } from '@ant-design/icons';
import { listModelFailureRates, listOverallFailureRates, listPackageFailureRates, listPackageModelFailureRates } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { FailureRateSummary, ModelFailureRate, PackageFailureRate, PackageModelFailureRate } from '../types';

const { Title, Text } = Typography;

const scopes: Array<{ key: 'all' | 'product' | 'devtest'; label: string }> = [
  { key: 'all', label: '整体' },
  { key: 'product', label: '生产' },
  { key: 'devtest', label: '开测' }
];

export default function FailureDashboardPage() {
  const [overall, setOverall] = useState<FailureRateSummary[]>([]);
  const [fm, setFm] = useState<ModelFailureRate[]>([]);
  const [fp, setFp] = useState<PackageFailureRate[]>([]);
  const [fpm, setFpm] = useState<PackageModelFailureRate[]>([]);
  const [fullScreen, setFullScreen] = useState(false);

  async function reload() {
    try {
      const [o, m, p, pm] = await Promise.all([
        listOverallFailureRates(),
        listModelFailureRates(),
        listPackageFailureRates(),
        listPackageModelFailureRates()
      ]);
      setOverall(ensureApiOk(o).data.list || []);
      setFm(ensureApiOk(m).data.list || []);
      setFp(ensureApiOk(p).data.list || []);
      setFpm(ensureApiOk(pm).data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载看板失败'));
    }
  }

  useEffect(() => {
    reload();
  }, []);

  useEffect(() => {
    const onChange = () => setFullScreen(!!document.fullscreenElement);
    document.addEventListener('fullscreenchange', onChange);
    return () => document.removeEventListener('fullscreenchange', onChange);
  }, []);

  const topModel = useMemo(() => [...fm].sort((a, b) => b.failure_rate - a.failure_rate).slice(0, 8), [fm]);
  const topPackage = useMemo(() => [...fp].sort((a, b) => b.failure_rate - a.failure_rate).slice(0, 8), [fp]);
  const topPackageModel = useMemo(() => [...fpm].sort((a, b) => b.failure_rate - a.failure_rate).slice(0, 8), [fpm]);

  const rows = useMemo(() => {
    return scopes.map((s) => ({
      scope: s,
      storage: findRate(overall, s.key, 'storage'),
      nonStorage: findRate(overall, s.key, 'non_storage')
    }));
  }, [overall]);

  async function toggleFullscreen() {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen();
    } else {
      await document.exitFullscreen();
    }
  }

  return (
    <div style={{ minHeight: '100vh', padding: 20, background: 'radial-gradient(circle at 20% 20%, #1d4ed8 0, #0f172a 35%, #020617 100%)' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Space style={{ justifyContent: 'space-between', width: '100%' }}>
          <Space>
            <Title level={2} style={{ margin: 0, color: '#e2e8f0' }}>⚡ 故障率分析看板</Title>
            <Tag color="blue">全屏友好</Tag>
          </Space>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={reload}>刷新</Button>
            <Button type="primary" icon={fullScreen ? <ShrinkOutlined /> : <ExpandOutlined />} onClick={toggleFullscreen}>
              {fullScreen ? '退出全屏' : '全屏展示'}
            </Button>
          </Space>
        </Space>

        {rows.map((row) => (
          <Card
            key={row.scope.key}
            title={<span style={{ color: '#dbeafe' }}>{row.scope.label}维度</span>}
            style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155', boxShadow: '0 0 40px rgba(56,189,248,0.15)' }}
            bodyStyle={{ paddingTop: 12 }}
          >
            <Row gutter={16}>
              <Col span={12}>
                <MetricCard title="存储" data={row.storage} glow="#22d3ee" />
              </Col>
              <Col span={12}>
                <MetricCard title="非存储" data={row.nonStorage} glow="#f472b6" />
              </Col>
            </Row>
          </Card>
        ))}

        <Row gutter={16}>
          <Col span={8}>
            <RankCard title="型号故障率 TOP8" rows={topModel.map((x) => ({ name: `${x.manufacturer}/${x.model}`, rate: x.failure_rate }))} />
          </Col>
          <Col span={8}>
            <RankCard title="套餐故障率 TOP8" rows={topPackage.map((x) => ({ name: x.config_type, rate: x.failure_rate }))} />
          </Col>
          <Col span={8}>
            <RankCard title="套餐型号故障率 TOP8" rows={topPackageModel.map((x) => ({ name: `${x.config_type}/${x.model}`, rate: x.failure_rate }))} />
          </Col>
        </Row>
      </Space>
    </div>
  );
}

function MetricCard({ title, data, glow }: { title: string; data?: FailureRateSummary; glow: string }) {
  return (
    <Card
      style={{
        background: 'linear-gradient(160deg, rgba(15,23,42,0.95), rgba(15,23,42,0.75))',
        border: `1px solid ${glow}`,
        boxShadow: `0 0 24px ${glow}55`
      }}
      bodyStyle={{ padding: 14 }}
    >
      <Space direction="vertical" style={{ width: '100%' }}>
        <Text style={{ color: '#bfdbfe', fontSize: 16 }}>{title}</Text>
        <Row gutter={12}>
          <Col span={12}>
            <Text style={{ color: '#94a3b8' }}>全周期</Text>
            <div style={{ color: '#e2e8f0', fontSize: 24, fontWeight: 700 }}>{formatPercent(data?.full_cycle_failure_rate)}</div>
          </Col>
          <Col span={12}>
            <Text style={{ color: '#94a3b8' }}>过保</Text>
            <div style={{ color: '#f8fafc', fontSize: 24, fontWeight: 700 }}>{formatPercent(data?.over_warranty_failure_rate)}</div>
          </Col>
        </Row>
        <Progress percent={Math.min(100, ratePercent(data?.full_cycle_failure_rate))} showInfo={false} strokeColor={glow} />
        <Progress percent={Math.min(100, ratePercent(data?.over_warranty_failure_rate))} showInfo={false} strokeColor="#a78bfa" />
        <Text style={{ color: '#cbd5e1' }}>故障数 {formatInt(data?.fault_count)} / 过保故障数 {formatInt(data?.over_warranty_fault_count)}</Text>
        <Text style={{ color: '#94a3b8' }}>全周期台年 {formatFloat(data?.server_years)} / 过保台年 {formatFloat(data?.over_warranty_years)}</Text>
      </Space>
    </Card>
  );
}

function RankCard({ title, rows }: { title: string; rows: Array<{ name: string; rate: number }> }) {
  return (
    <Card title={<span style={{ color: '#dbeafe' }}>{title}</span>} style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }}>
      <Table
        size="small"
        pagination={false}
        rowKey={(r) => r.name}
        dataSource={rows}
        columns={[
          { title: '对象', dataIndex: 'name' },
          { title: '故障率', dataIndex: 'rate', render: (v: number) => formatPercent(v) }
        ]}
      />
    </Card>
  );
}

function findRate(rows: FailureRateSummary[], scope: string, segment: string) {
  return rows.find((x) => x.scope === scope && x.segment === segment);
}

function ratePercent(v?: number) {
  return Number(((v || 0) * 100).toFixed(2));
}

function formatPercent(v?: number) {
  return `${ratePercent(v).toFixed(2)}%`;
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
