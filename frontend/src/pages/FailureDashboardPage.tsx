import { useEffect, useMemo, useState } from 'react';
import { Button, Card, Col, Row, Space, Tag, Typography, message } from 'antd';
import { ExpandOutlined, ReloadOutlined, ShrinkOutlined } from '@ant-design/icons';
import { listHostPackages, listOverallFailureRates, listPackageFailureRates, listPackageModelFailureRates } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { FailureRateSummary, HostPackageConfig, PackageFailureRate, PackageModelFailureRate } from '../types';

const { Title, Text } = Typography;

const scopes: Array<{ key: 'all' | 'product' | 'devtest'; label: string }> = [
  { key: 'all', label: '全部环境' },
  { key: 'product', label: '生产' },
  { key: 'devtest', label: '开测' }
];

const bucketGroups: Array<{ key: string; label: string }> = [
  { key: 'compute', label: '计算' },
  { key: 'warm_storage', label: '温存储' },
  { key: 'hot_storage', label: '热存储' },
  { key: 'gpu', label: 'GPU' }
];

export default function FailureDashboardPage() {
  const [overall, setOverall] = useState<FailureRateSummary[]>([]);
  const [hostPackages, setHostPackages] = useState<HostPackageConfig[]>([]);
  const [fp, setFp] = useState<PackageFailureRate[]>([]);
  const [fpm, setFpm] = useState<PackageModelFailureRate[]>([]);
  const [fullScreen, setFullScreen] = useState(false);
  const [groupIndex, setGroupIndex] = useState(0);

  async function reload() {
    try {
      const [o, hp, p, pm] = await Promise.all([
        listOverallFailureRates(),
        listHostPackages(),
        listPackageFailureRates(),
        listPackageModelFailureRates()
      ]);
      setOverall(ensureApiOk(o).data.list || []);
      setHostPackages(ensureApiOk(hp).data.list || []);
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
    const timer = setInterval(() => {
      setGroupIndex((prev) => (prev + 1) % bucketGroups.length);
    }, 10000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    const onChange = () => setFullScreen(!!document.fullscreenElement);
    document.addEventListener('fullscreenchange', onChange);
    return () => document.removeEventListener('fullscreenchange', onChange);
  }, []);

  const currentYear = useMemo(() => {
    const y = overall.find((x) => x.period === 'year')?.year;
    return y || new Date().getFullYear();
  }, [overall]);

  const pkgBucketMap = useMemo(() => {
    const m = new Map<string, string>();
    hostPackages.forEach((x) => m.set((x.config_type || '').trim(), normalizeBucket(x.scene_category || '')));
    return m;
  }, [hostPackages]);

  const activeGroup = bucketGroups[groupIndex];

  const topPackage = useMemo(() => {
    return fp
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .sort((a, b) => b.failure_rate - a.failure_rate)
      .slice(0, 8)
      .map((x) => ({ name: x.config_type, rate: x.failure_rate }));
  }, [fp, pkgBucketMap, activeGroup.key]);

  const topPackageModel = useMemo(() => {
    return fpm
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .sort((a, b) => b.failure_rate - a.failure_rate)
      .slice(0, 8)
      .map((x) => ({ name: `${x.config_type}/${x.model}`, rate: x.failure_rate }));
  }, [fpm, pkgBucketMap, activeGroup.key]);

  const topModel = useMemo(() => {
    const grouped = new Map<string, number>();
    fpm
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .forEach((x) => {
        const k = `${x.manufacturer}/${x.model}`;
        const old = grouped.get(k) ?? 0;
        if (x.failure_rate > old) grouped.set(k, x.failure_rate);
      });
    return Array.from(grouped.entries())
      .map(([name, rate]) => ({ name, rate }))
      .sort((a, b) => b.rate - a.rate)
      .slice(0, 8);
  }, [fpm, pkgBucketMap, activeGroup.key]);

  async function toggleFullscreen() {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen();
    } else {
      await document.exitFullscreen();
    }
  }

  return (
    <div style={{ minHeight: '100vh', padding: 18, background: 'radial-gradient(circle at 20% 20%, #1d4ed8 0, #0f172a 35%, #020617 100%)' }}>
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Space style={{ justifyContent: 'space-between', width: '100%' }}>
          <Space>
            <Title level={2} style={{ margin: 0, color: '#e2e8f0' }}>⚡ 故障率分析看板</Title>
            <Tag color="blue">10s轮播</Tag>
          </Space>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={reload}>刷新</Button>
            <Button type="primary" icon={fullScreen ? <ShrinkOutlined /> : <ExpandOutlined />} onClick={toggleFullscreen}>
              {fullScreen ? '退出全屏' : '全屏展示'}
            </Button>
          </Space>
        </Space>

        <SummaryGroup title="历史平均故障率" period="history" data={overall} />
        <SummaryGroup title={`${currentYear}年故障率`} period="year" data={overall} />

        <Card
          title={<span style={{ color: '#dbeafe' }}>TOP故障率轮播：{activeGroup.label}</span>}
          className="oc-neon-panel"
          style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }}
          bodyStyle={{ paddingTop: 10 }}
          extra={<Text style={{ color: '#93c5fd' }}>每10秒自动切换</Text>}
        >
          <Row gutter={12}>
            <Col span={8}><RankCard title="型号 TOP8" rows={topModel} /></Col>
            <Col span={8}><RankCard title="套餐 TOP8" rows={topPackage} /></Col>
            <Col span={8}><RankCard title="套餐型号 TOP8" rows={topPackageModel} /></Col>
          </Row>
        </Card>
      </Space>
    </div>
  );
}

function SummaryGroup({ title, period, data }: { title: string; period: 'history' | 'year'; data: FailureRateSummary[] }) {
  return (
    <Card title={<span style={{ color: '#dbeafe' }}>{title}</span>} style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }} bodyStyle={{ paddingTop: 10 }}>
      <Row gutter={12}>
        {scopes.map((scope) => (
          <Col span={8} key={`${period}-${scope.key}`}>
            <Space direction="vertical" size={8} style={{ width: '100%' }}>
              <MetricCard title={`${scope.label} - 非存储`} data={findRate(data, period, scope.key, 'non_storage')} glow="#22d3ee" />
              <MetricCard title={`${scope.label} - 存储`} data={findRate(data, period, scope.key, 'storage')} glow="#a78bfa" />
            </Space>
          </Col>
        ))}
      </Row>
    </Card>
  );
}

function MetricCard({ title, data, glow }: { title: string; data?: FailureRateSummary; glow: string }) {
  return (
    <Card
      className="oc-neon-card"
      style={{
        background: 'linear-gradient(160deg, rgba(15,23,42,0.96), rgba(15,23,42,0.72))',
        border: `1px solid ${glow}`,
        boxShadow: `0 0 18px ${glow}55`
      }}
      bodyStyle={{ padding: 12 }}
    >
      <Space direction="vertical" style={{ width: '100%' }} size={6}>
        <Text style={{ color: '#bfdbfe', fontSize: 15 }}>{title}</Text>
        <Row gutter={8}>
          <Col span={12}>
            <Text style={{ color: '#94a3b8' }}>全周期</Text>
            <div style={{ color: '#e2e8f0', fontSize: 22, fontWeight: 700 }}>{formatPercent(data?.full_cycle_failure_rate)}</div>
          </Col>
          <Col span={12}>
            <Text style={{ color: '#94a3b8' }}>过保</Text>
            <div style={{ color: '#f8fafc', fontSize: 22, fontWeight: 700 }}>{formatPercent(data?.over_warranty_failure_rate)}</div>
          </Col>
        </Row>
        <Text style={{ color: '#cbd5e1' }}>故障数 {formatInt(data?.fault_count)} / 过保故障数 {formatInt(data?.over_warranty_fault_count)}</Text>
      </Space>
    </Card>
  );
}

function RankCard({ title, rows }: { title: string; rows: Array<{ name: string; rate: number }> }) {
  return (
    <Card className="oc-neon-card" style={{ background: 'rgba(15,23,42,0.75)', border: '1px solid #334155' }} bodyStyle={{ padding: 10 }}>
      <Text style={{ color: '#93c5fd' }}>{title}</Text>
      <div style={{ marginTop: 8 }}>
        {rows.map((x, idx) => (
          <div key={x.name} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 8px', borderRadius: 6, background: idx % 2 ? 'rgba(30,41,59,0.55)' : 'rgba(15,23,42,0.75)', marginBottom: 4 }}>
            <span style={{ color: '#e2e8f0', maxWidth: '72%', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{idx + 1}. {x.name}</span>
            <span style={{ color: '#22d3ee', fontWeight: 700 }}>{formatPercent(x.rate)}</span>
          </div>
        ))}
      </div>
    </Card>
  );
}

function normalizeBucket(scene?: string) {
  const n = String(scene || '').toLowerCase().replace(/[\s_-]/g, '');
  if (['计算型', '计算', 'compute', 'generalcompute', '通用计算', 'cpu'].includes(n)) return 'compute';
  if (['温存储', '温', 'warmstorage', 'warm', 'coldstorage', '温储'].includes(n)) return 'warm_storage';
  if (['热存储', '热', 'hotstorage', 'hot', '热储'].includes(n)) return 'hot_storage';
  if (['gpu', 'gpu型', 'gpu计算', 'gpucompute'].includes(n)) return 'gpu';
  return 'compute';
}

function findRate(rows: FailureRateSummary[], period: string, scope: string, segment: string) {
  return rows.find((x) => x.period === period && x.scope === scope && x.segment === segment);
}

function formatPercent(v?: number) {
  const p = Number(((v || 0) * 100).toFixed(2));
  return `${p.toFixed(2)}%`;
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}
