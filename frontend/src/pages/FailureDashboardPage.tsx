import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Row, Space, Typography, message } from 'antd';
import { ExpandOutlined, ReloadOutlined, ShrinkOutlined } from '@ant-design/icons';
import { listFailureAgeTrendPoints, listFailureOverviewCards } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { FailureAgeTrendPoint, FailureOverviewCard } from '../types';

const { Title, Text } = Typography;

export default function FailureDashboardPage() {
  const [overviewCards, setOverviewCards] = useState<FailureOverviewCard[]>([]);
  const [trendPoints, setTrendPoints] = useState<FailureAgeTrendPoint[]>([]);
  const [fullScreen, setFullScreen] = useState(false);

  async function reload() {
    try {
      const [overviewResp, trendResp] = await Promise.all([
        listFailureOverviewCards(),
        listFailureAgeTrendPoints()
      ]);
      setOverviewCards(ensureApiOk(overviewResp).data.list || []);
      setTrendPoints(ensureApiOk(trendResp).data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载看板失败'));
    }
  }

  useEffect(() => {
    reload();
  }, []);

  const storageCard = useMemo(() => overviewCards.find((x) => x.segment === 'storage'), [overviewCards]);
  const nonStorageCard = useMemo(() => overviewCards.find((x) => x.segment === 'non_storage'), [overviewCards]);

  const storageTrend = useMemo(
    () => fillAges(trendPoints.filter((x) => x.segment === 'storage')),
    [trendPoints]
  );
  const nonStorageTrend = useMemo(
    () => fillAges(trendPoints.filter((x) => x.segment === 'non_storage')),
    [trendPoints]
  );

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
          <Title level={2} style={{ margin: 0, color: '#e2e8f0' }}>⚡ 故障率分析看板</Title>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={reload}>刷新</Button>
            <Button type="primary" icon={fullScreen ? <ShrinkOutlined /> : <ExpandOutlined />} onClick={toggleFullscreen}>
              {fullScreen ? '退出全屏' : '全屏展示'}
            </Button>
          </Space>
        </Space>

        <Card title={<span style={{ color: '#dbeafe' }}>故障率概览</span>} className="oc-neon-panel" style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }}>
          <Row gutter={12}>
            <Col span={12}><OverviewCard title="存储" data={storageCard} /></Col>
            <Col span={12}><OverviewCard title="非存储" data={nonStorageCard} /></Col>
          </Row>
          {!storageCard && !nonStorageCard ? <Alert style={{ marginTop: 12 }} type="info" showIcon message="暂无概览数据，请先在故障清单分析中执行重算" /> : null}
        </Card>

        <Card title={<span style={{ color: '#dbeafe' }}>故障率趋势</span>} className="oc-neon-panel" style={{ background: 'rgba(2,6,23,0.65)', border: '1px solid #334155' }}>
          <Row gutter={12}>
            <Col span={12}><TrendCard title="存储（1-10年）" points={storageTrend} /></Col>
            <Col span={12}><TrendCard title="非存储（1-10年）" points={nonStorageTrend} /></Col>
          </Row>
        </Card>
      </Space>
    </div>
  );
}

function OverviewCard({ title, data }: { title: string; data?: FailureOverviewCard }) {
  return (
    <Card className="oc-neon-card" style={{ background: 'rgba(15,23,42,0.75)', border: '1px solid #334155' }} bodyStyle={{ padding: 14 }}>
      <Text style={{ color: '#93c5fd', fontSize: 16 }}>{title}</Text>
      <Row gutter={12} style={{ marginTop: 10 }}>
        <Col span={12}>
          <Text style={{ color: '#94a3b8' }}>{data?.year || new Date().getFullYear()}年故障率</Text>
          <div style={{ color: '#22d3ee', fontSize: 26, fontWeight: 700 }}>{formatPercent(data?.current_year_fault_rate)}</div>
          <Text style={{ color: '#cbd5e1' }}>故障 {formatInt(data?.current_year_fault_count)} / 分母 {formatFloat(data?.current_year_denominator)}</Text>
        </Col>
        <Col span={12}>
          <Text style={{ color: '#94a3b8' }}>历史平均故障率</Text>
          <div style={{ color: '#a78bfa', fontSize: 26, fontWeight: 700 }}>{formatPercent(data?.history_avg_fault_rate)}</div>
          <Text style={{ color: '#cbd5e1' }}>故障 {formatInt(data?.history_fault_count)} / 分母 {formatFloat(data?.history_denominator)}</Text>
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
