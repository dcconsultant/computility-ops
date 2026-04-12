import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Row, Space, Tag, Typography, message } from 'antd';
import { ExpandOutlined, ReloadOutlined, ShrinkOutlined } from '@ant-design/icons';
import {
  listFailureAgeTrendPoints,
  listHostPackages,
  listOverallFailureRates,
  listPackageFailureRates,
  listPackageModelFailureRates
} from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type {
  FailureAgeTrendPoint,
  FailureRateSummary,
  HostPackageConfig,
  PackageFailureRate,
  PackageModelFailureRate
} from '../types';

const { Title, Text } = Typography;

const scopeDefs: Array<{ key: 'all' | 'product' | 'devtest'; label: string }> = [
  { key: 'all', label: '整体' },
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
  const [trendPoints, setTrendPoints] = useState<FailureAgeTrendPoint[]>([]);
  const [hostPackages, setHostPackages] = useState<HostPackageConfig[]>([]);
  const [fp, setFp] = useState<PackageFailureRate[]>([]);
  const [fpm, setFpm] = useState<PackageModelFailureRate[]>([]);
  const [fullScreen, setFullScreen] = useState(false);
  const [groupIndex, setGroupIndex] = useState(0);

  async function reload() {
    try {
      const [o, t, hp, p, pm] = await Promise.all([
        listOverallFailureRates(),
        listFailureAgeTrendPoints(),
        listHostPackages(),
        listPackageFailureRates(),
        listPackageModelFailureRates()
      ]);
      setOverall(ensureApiOk(o).data.list || []);
      setTrendPoints(ensureApiOk(t).data.list || []);
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

  const currentYear = useMemo(() => {
    return overall.find((x) => x.period === 'year')?.year || new Date().getFullYear();
  }, [overall]);

  const storageTrend = useMemo(
    () => fillAges(trendPoints.filter((x) => x.segment === 'storage')),
    [trendPoints]
  );
  const nonStorageTrend = useMemo(
    () => fillAges(trendPoints.filter((x) => x.segment === 'non_storage')),
    [trendPoints]
  );

  const storageYearTrend = useMemo(
    () => buildYearTrend(overall, 'storage', currentYear),
    [overall, currentYear]
  );
  const nonStorageYearTrend = useMemo(
    () => buildYearTrend(overall, 'non_storage', currentYear),
    [overall, currentYear]
  );

  const pkgBucketMap = useMemo(() => {
    const m = new Map<string, string>();
    hostPackages.forEach((x) => m.set((x.config_type || '').trim(), normalizeBucket(x.scene_category || '')));
    return m;
  }, [hostPackages]);

  const pkgConfigMap = useMemo(() => {
    const m = new Map<string, HostPackageConfig>();
    hostPackages.forEach((x) => m.set((x.config_type || '').trim(), x));
    return m;
  }, [hostPackages]);

  const activeGroup = bucketGroups[groupIndex];

  const fpHistory = useMemo(() => fp.filter((x) => (x.period || 'history') === 'history'), [fp]);
  const fpYear = useMemo(() => fp.filter((x) => x.period === 'year' && x.year === currentYear), [fp, currentYear]);
  const fpmHistory = useMemo(() => fpm.filter((x) => (x.period || 'history') === 'history'), [fpm]);
  const fpmYear = useMemo(() => fpm.filter((x) => x.period === 'year' && x.year === currentYear), [fpm, currentYear]);

  const topPackage = useMemo(() => {
    return fpHistory
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .sort((a, b) => b.failure_rate - a.failure_rate)
      .slice(0, 8)
      .map((x) => ({
        name: x.config_type,
        rate: x.failure_rate,
        scale: scaleLabelByConfig((x.config_type || '').trim(), activeGroup.key, pkgConfigMap)
      }));
  }, [fpHistory, pkgBucketMap, activeGroup.key, pkgConfigMap]);

  const topPackageYear = useMemo(() => {
    return fpYear
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .sort((a, b) => b.failure_rate - a.failure_rate)
      .slice(0, 8)
      .map((x) => ({
        name: x.config_type,
        rate: x.failure_rate,
        scale: scaleLabelByConfig((x.config_type || '').trim(), activeGroup.key, pkgConfigMap)
      }));
  }, [fpYear, pkgBucketMap, activeGroup.key, pkgConfigMap]);

  const topPackageModel = useMemo(() => {
    return fpmHistory
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .sort((a, b) => b.failure_rate - a.failure_rate)
      .slice(0, 8)
      .map((x) => ({
        name: `${x.config_type}/${x.model}`,
        rate: x.failure_rate,
        scale: scaleLabelByConfig((x.config_type || '').trim(), activeGroup.key, pkgConfigMap)
      }));
  }, [fpmHistory, pkgBucketMap, activeGroup.key, pkgConfigMap]);

  const topPackageModelYear = useMemo(() => {
    return fpmYear
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .sort((a, b) => b.failure_rate - a.failure_rate)
      .slice(0, 8)
      .map((x) => ({
        name: `${x.config_type}/${x.model}`,
        rate: x.failure_rate,
        scale: scaleLabelByConfig((x.config_type || '').trim(), activeGroup.key, pkgConfigMap)
      }));
  }, [fpmYear, pkgBucketMap, activeGroup.key, pkgConfigMap]);

  const topModel = useMemo(() => {
    const grouped = new Map<string, { rate: number; scale: string }>();
    fpmHistory
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .forEach((x) => {
        const k = `${x.manufacturer}/${x.model}`;
        const candidate = {
          rate: x.failure_rate,
          scale: scaleLabelByConfig((x.config_type || '').trim(), activeGroup.key, pkgConfigMap)
        };
        const old = grouped.get(k);
        if (!old || candidate.rate > old.rate) grouped.set(k, candidate);
      });
    return Array.from(grouped.entries())
      .map(([name, v]) => ({ name, rate: v.rate, scale: v.scale }))
      .sort((a, b) => b.rate - a.rate)
      .slice(0, 8);
  }, [fpmHistory, pkgBucketMap, activeGroup.key, pkgConfigMap]);

  const topModelYear = useMemo(() => {
    const grouped = new Map<string, { rate: number; scale: string }>();
    fpmYear
      .filter((x) => pkgBucketMap.get((x.config_type || '').trim()) === activeGroup.key)
      .forEach((x) => {
        const k = `${x.manufacturer}/${x.model}`;
        const candidate = {
          rate: x.failure_rate,
          scale: scaleLabelByConfig((x.config_type || '').trim(), activeGroup.key, pkgConfigMap)
        };
        const old = grouped.get(k);
        if (!old || candidate.rate > old.rate) grouped.set(k, candidate);
      });
    return Array.from(grouped.entries())
      .map(([name, v]) => ({ name, rate: v.rate, scale: v.scale }))
      .sort((a, b) => b.rate - a.rate)
      .slice(0, 8);
  }, [fpmYear, pkgBucketMap, activeGroup.key, pkgConfigMap]);

  async function toggleFullscreen() {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen();
      setFullScreen(true);
    } else {
      await document.exitFullscreen();
      setFullScreen(false);
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

        <Card title={<span style={{ color: '#dbeafe' }}>故障率概览</span>} className="oc-neon-panel" style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }}>
          <Row gutter={10}>
            {scopeDefs.map((scope) => (
              <Col span={8} key={scope.key}>
                <ScopeGroup
                  title={scope.label}
                  year={currentYear}
                  nonStorage={buildOverviewPair(overall, scope.key, 'non_storage')}
                  storage={buildOverviewPair(overall, scope.key, 'storage')}
                />
              </Col>
            ))}
          </Row>
        </Card>

        <Card title={<span style={{ color: '#dbeafe' }}>故障率趋势</span>} className="oc-neon-panel" style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }}>
          <Row gutter={12}>
            <Col span={12}><TrendCard title="存储（1-10年）" points={storageTrend} /></Col>
            <Col span={12}><TrendCard title="非存储（1-10年）" points={nonStorageTrend} /></Col>
          </Row>
          <Row gutter={12} style={{ marginTop: 12 }}>
            <Col span={12}><YearTrendCard title="存储历年故障率（2021-至今）" points={storageYearTrend} /></Col>
            <Col span={12}><YearTrendCard title="非存储历年故障率（2021-至今）" points={nonStorageYearTrend} /></Col>
          </Row>
          {storageTrend.every((x) => x.denominator_exposure === 0) && nonStorageTrend.every((x) => x.denominator_exposure === 0)
            ? <Alert style={{ marginTop: 12 }} type="info" showIcon message="暂无趋势数据，请先执行故障清单分析" />
            : null}
        </Card>

        <Card
          title={<span style={{ color: '#dbeafe' }}>历史TOP故障率清单：{activeGroup.label}</span>}
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

        <Card
          title={<span style={{ color: '#dbeafe' }}>{currentYear}年TOP故障率清单：{activeGroup.label}</span>}
          className="oc-neon-panel"
          style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }}
          bodyStyle={{ paddingTop: 10 }}
        >
          <Row gutter={12}>
            <Col span={8}><RankCard title="型号 TOP8" rows={topModelYear} /></Col>
            <Col span={8}><RankCard title="套餐 TOP8" rows={topPackageYear} /></Col>
            <Col span={8}><RankCard title="套餐型号 TOP8" rows={topPackageModelYear} /></Col>
          </Row>
        </Card>
      </Space>
    </div>
  );
}

function ScopeGroup({ title, year, nonStorage, storage }: {
  title: string;
  year: number;
  nonStorage: OverviewPair;
  storage: OverviewPair;
}) {
  return (
    <Card className="oc-neon-card" style={{ background: 'rgba(15,23,42,0.55)', border: '1px solid #334155' }} bodyStyle={{ padding: 10 }}>
      <Text style={{ color: '#93c5fd', fontSize: 15 }}>{title}</Text>
      <Space direction="vertical" size={8} style={{ width: '100%', marginTop: 8 }}>
        <MiniOverviewCard segmentLabel="非存储" year={year} data={nonStorage} glow="#22d3ee" />
        <MiniOverviewCard segmentLabel="存储" year={year} data={storage} glow="#a78bfa" />
      </Space>
    </Card>
  );
}

type OverviewPair = {
  yearRate: number;
  historyRate: number;
  historyOverRate: number;
  yearFault: number;
  yearDen: number;
  historyFault: number;
  historyDen: number;
};

function MiniOverviewCard({ segmentLabel, year, data, glow }: {
  segmentLabel: string;
  year: number;
  data: OverviewPair;
  glow: string;
}) {
  return (
    <Card
      className="oc-neon-card"
      style={{
        background: 'linear-gradient(160deg, rgba(15,23,42,0.96), rgba(15,23,42,0.72))',
        border: `1px solid ${glow}`,
        boxShadow: `0 0 14px ${glow}44`
      }}
      bodyStyle={{ padding: 10 }}
    >
      <Text style={{ color: '#bfdbfe', fontSize: 13 }}>{segmentLabel}</Text>
      <Row gutter={8} style={{ marginTop: 6 }}>
        <Col span={12}>
          <Text style={{ color: '#94a3b8', fontSize: 12 }}>{year}年故障率</Text>
          <div style={{ color: '#22d3ee', fontSize: 20, fontWeight: 700 }}>{formatPercent(data.yearRate)}</div>
          <Text style={{ color: '#cbd5e1', fontSize: 12 }}>故障 {formatInt(data.yearFault)} / 分母 {formatFloat(data.yearDen)}</Text>
        </Col>
        <Col span={12}>
          <Text style={{ color: '#94a3b8', fontSize: 12 }}>历史平均故障率</Text>
          <div style={{ color: '#a78bfa', fontSize: 20, fontWeight: 700 }}>{formatPercent(data.historyRate)}</div>
          <Text style={{ color: '#cbd5e1', fontSize: 12 }}>故障 {formatInt(data.historyFault)} / 分母 {formatFloat(data.historyDen)}</Text>
          <div style={{ color: '#f0abfc', fontSize: 12, marginTop: 2 }}>过保故障率 {formatPercent(data.historyOverRate)}</div>
        </Col>
      </Row>
    </Card>
  );
}

function TrendCard({ title, points }: { title: string; points: FailureAgeTrendPoint[] }) {
  const width = 520;
  const height = 240;
  const padding = 30;
  const maxY = Math.max(0.01, ...points.map((p) => p.fault_rate));
  const linePoints = points
    .map((p, idx) => {
      const x = padding + (idx * (width - padding * 2)) / 9;
      const y = height - padding - (p.fault_rate / maxY) * (height - padding * 2);
      return `${x},${y}`;
    })
    .join(' ');

  return (
    <Card className="oc-neon-card" style={{ background: 'rgba(15,23,42,0.75)', border: '1px solid #334155' }} bodyStyle={{ padding: 12 }}>
      <Text style={{ color: '#93c5fd' }}>{title}</Text>
      <svg viewBox={`0 0 ${width} ${height}`} style={{ width: '100%', height: 240, marginTop: 8 }}>
        <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} stroke="#334155" />
        <line x1={padding} y1={padding} x2={padding} y2={height - padding} stroke="#334155" />
        <polyline fill="none" stroke="#22d3ee" strokeWidth="3" points={linePoints} />
        {points.map((p, idx) => {
          const x = padding + (idx * (width - padding * 2)) / 9;
          const y = height - padding - (p.fault_rate / maxY) * (height - padding * 2);
          return (
            <g key={`${p.segment}-${p.age_bucket}`}>
              <circle cx={x} cy={y} r={4} fill="#22d3ee" />
              <text x={x} y={height - 10} textAnchor="middle" fill="#94a3b8" fontSize="11">{p.age_bucket}年</text>
              <text x={x} y={y - 8} textAnchor="middle" fill="#cbd5e1" fontSize="10">{(p.fault_rate * 100).toFixed(1)}%</text>
            </g>
          );
        })}
      </svg>
    </Card>
  );
}

type YearTrendPoint = { year: number; rate: number; fault: number; denominator: number };

function YearTrendCard({ title, points }: { title: string; points: YearTrendPoint[] }) {
  const width = 520;
  const height = 240;
  const padding = 30;
  const maxY = Math.max(0.01, ...points.map((p) => p.rate));
  const n = Math.max(points.length - 1, 1);
  const linePoints = points
    .map((p, idx) => {
      const x = padding + (idx * (width - padding * 2)) / n;
      const y = height - padding - (p.rate / maxY) * (height - padding * 2);
      return `${x},${y}`;
    })
    .join(' ');

  return (
    <Card className="oc-neon-card" style={{ background: 'rgba(15,23,42,0.75)', border: '1px solid #334155' }} bodyStyle={{ padding: 12 }}>
      <Text style={{ color: '#93c5fd' }}>{title}</Text>
      <svg viewBox={`0 0 ${width} ${height}`} style={{ width: '100%', height: 240, marginTop: 8 }}>
        <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} stroke="#334155" />
        <line x1={padding} y1={padding} x2={padding} y2={height - padding} stroke="#334155" />
        <polyline fill="none" stroke="#a78bfa" strokeWidth="3" points={linePoints} />
        {points.map((p, idx) => {
          const x = padding + (idx * (width - padding * 2)) / n;
          const y = height - padding - (p.rate / maxY) * (height - padding * 2);
          return (
            <g key={`${p.year}`}>
              <circle cx={x} cy={y} r={4} fill="#a78bfa" />
              <text x={x} y={height - 10} textAnchor="middle" fill="#94a3b8" fontSize="11">{p.year}</text>
              <text x={x} y={y - 8} textAnchor="middle" fill="#cbd5e1" fontSize="10">{(p.rate * 100).toFixed(1)}%</text>
            </g>
          );
        })}
      </svg>
    </Card>
  );
}

function RankCard({ title, rows }: { title: string; rows: Array<{ name: string; rate: number; scale?: string }> }) {
  return (
    <Card className="oc-neon-card" style={{ background: 'rgba(15,23,42,0.75)', border: '1px solid #334155' }} bodyStyle={{ padding: 10 }}>
      <Text style={{ color: '#93c5fd' }}>{title}</Text>
      <div style={{ marginTop: 8 }}>
        {rows.map((x, idx) => (
          <div key={x.name} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 8px', borderRadius: 6, background: idx % 2 ? 'rgba(30,41,59,0.55)' : 'rgba(15,23,42,0.75)', marginBottom: 4 }}>
            <span style={{ color: '#e2e8f0', maxWidth: '72%', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{idx + 1}. {x.name}</span>
            <span style={{ display: 'inline-flex', gap: 8, alignItems: 'baseline' }}>
              <span style={{ color: '#22d3ee', fontWeight: 700 }}>{formatPercent(x.rate)}</span>
              <span style={{ color: '#94a3b8', fontSize: 12 }}>{x.scale || ''}</span>
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
}

function buildOverviewPair(rows: FailureRateSummary[], scope: 'all' | 'product' | 'devtest', segment: 'storage' | 'non_storage'): OverviewPair {
  const y = rows.find((x) => x.period === 'year' && x.scope === scope && x.segment === segment);
  const h = rows.find((x) => x.period === 'history' && x.scope === scope && x.segment === segment);
  return {
    yearRate: y?.full_cycle_failure_rate || 0,
    historyRate: h?.full_cycle_failure_rate || 0,
    historyOverRate: h?.over_warranty_failure_rate || 0,
    yearFault: y?.fault_count || 0,
    yearDen: y?.server_years || 0,
    historyFault: h?.fault_count || 0,
    historyDen: h?.server_years || 0
  };
}

function buildYearTrend(rows: FailureRateSummary[], segment: 'storage' | 'non_storage', currentYear: number): YearTrendPoint[] {
  const m = new Map<number, FailureRateSummary>();
  rows
    .filter((x) => x.period === 'year_trend' && x.scope === 'all' && x.segment === segment && typeof x.year === 'number')
    .forEach((x) => m.set(x.year as number, x));

  const startYear = 2021;
  const out: YearTrendPoint[] = [];
  for (let y = startYear; y <= currentYear; y++) {
    const row = m.get(y);
    out.push({
      year: y,
      rate: row?.full_cycle_failure_rate || 0,
      fault: row?.fault_count || 0,
      denominator: row?.server_years || 0
    });
  }
  return out;
}

function fillAges(rows: FailureAgeTrendPoint[]) {
  const m = new Map<number, FailureAgeTrendPoint>();
  rows.forEach((x) => m.set(x.age_bucket, x));
  const segment = rows[0]?.segment || 'unknown';
  const out: FailureAgeTrendPoint[] = [];
  for (let age = 1; age <= 10; age++) {
    out.push(m.get(age) || {
      segment,
      age_bucket: age,
      numerator_fault_count: 0,
      denominator_exposure: 0,
      fault_rate: 0
    });
  }
  return out;
}

function scaleLabelByConfig(configType: string, bucket: string, pkgMap: Map<string, HostPackageConfig>) {
  const cfg = pkgMap.get(configType);
  if (!cfg) return '';
  if (bucket === 'warm_storage' || bucket === 'hot_storage' || bucket === 'storage') {
    const tb = Number(cfg.storage_capacity_tb || 0);
    const pb = tb / 1000;
    return `${pb.toFixed(2)}PB`;
  }
  if (bucket === 'gpu') {
    return `${Number(cfg.gpu_card_count || 0)}卡`;
  }
  return `${Number(cfg.cpu_logical_cores || 0)}核`;
}

function normalizeBucket(scene?: string) {
  const n = String(scene || '').toLowerCase().replace(/[\s_-]/g, '');
  if (['计算型', '计算', 'compute', 'generalcompute', '通用计算', 'cpu'].includes(n)) return 'compute';
  if (['温存储', '温', 'warmstorage', 'warm', 'coldstorage', '温储'].includes(n)) return 'warm_storage';
  if (['热存储', '热', 'hotstorage', 'hot', '热储'].includes(n)) return 'hot_storage';
  if (['gpu', 'gpu型', 'gpu计算', 'gpucompute'].includes(n)) return 'gpu';
  return 'compute';
}

function formatPercent(v?: number) {
  const p = Number(((v || 0) * 100).toFixed(2));
  return `${p.toFixed(2)}%`;
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
