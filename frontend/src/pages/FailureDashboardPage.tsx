import { useEffect, useMemo, useState } from 'react';
import { Button, Card, Col, Progress, Row, Space, Statistic, Table, Typography, message } from 'antd';
import { ExpandOutlined, ReloadOutlined, ShrinkOutlined } from '@ant-design/icons';
import { listModelFailureRates, listOverallFailureRates, listPackageFailureRates, listPackageModelFailureRates } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { FailureRateSummary, ModelFailureRate, PackageFailureRate, PackageModelFailureRate } from '../types';

const { Title, Text } = Typography;

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

  const storage = overall.find((x) => x.segment === 'storage');
  const nonStorage = overall.find((x) => x.segment === 'non_storage');

  async function toggleFullscreen() {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen();
    } else {
      await document.exitFullscreen();
    }
  }

  return (
    <div style={{ minHeight: 'calc(100vh - 96px)', padding: 12, background: 'linear-gradient(135deg,#0f172a,#1e293b,#111827)' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Space style={{ justifyContent: 'space-between', width: '100%' }}>
          <Title level={3} style={{ margin: 0, color: '#fff' }}>⚡ 故障率分析看板</Title>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={reload}>刷新</Button>
            <Button type="primary" icon={fullScreen ? <ShrinkOutlined /> : <ExpandOutlined />} onClick={toggleFullscreen}>
              {fullScreen ? '退出全屏' : '全屏展示'}
            </Button>
          </Space>
        </Space>

        <Row gutter={16}>
          {[storage, nonStorage].map((item) => (
            <Col span={12} key={item?.segment || 'na'}>
              <Card style={{ background: '#0b1220', border: '1px solid #334155' }}>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text style={{ color: '#93c5fd' }}>{item?.segment === 'storage' ? '存储' : '非存储'}</Text>
                  <Row gutter={16}>
                    <Col span={12}><Statistic title="全周期故障率" value={ratePercent(item?.full_cycle_failure_rate)} suffix="%" valueStyle={{ color: '#22d3ee' }} /></Col>
                    <Col span={12}><Statistic title="过保故障率" value={ratePercent(item?.over_warranty_failure_rate)} suffix="%" valueStyle={{ color: '#f472b6' }} /></Col>
                  </Row>
                  <Progress percent={Math.min(100, ratePercent(item?.full_cycle_failure_rate))} strokeColor="#22d3ee" showInfo={false} />
                  <Progress percent={Math.min(100, ratePercent(item?.over_warranty_failure_rate))} strokeColor="#f472b6" showInfo={false} />
                  <Text style={{ color: '#cbd5e1' }}>故障数 {formatInt(item?.fault_count)} / 过保故障数 {formatInt(item?.over_warranty_fault_count)}</Text>
                  <Text style={{ color: '#94a3b8' }}>全周期台年 {formatFloat(item?.server_years)} / 过保台年 {formatFloat(item?.over_warranty_years)}</Text>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>

        <Row gutter={16}>
          <Col span={8}>
            <Card title="型号故障率 TOP8" style={{ background: '#0b1220', border: '1px solid #334155' }} headStyle={{ color: '#fff' }}>
              <Table
                size="small"
                pagination={false}
                rowKey={(r) => `${r.manufacturer}-${r.model}`}
                dataSource={topModel}
                columns={[
                  { title: '型号', dataIndex: 'model' },
                  { title: '年化', dataIndex: 'failure_rate', render: (v: number) => formatPercent(v) }
                ]}
              />
            </Card>
          </Col>
          <Col span={8}>
            <Card title="套餐故障率 TOP8" style={{ background: '#0b1220', border: '1px solid #334155' }} headStyle={{ color: '#fff' }}>
              <Table
                size="small"
                pagination={false}
                rowKey="config_type"
                dataSource={topPackage}
                columns={[
                  { title: '套餐', dataIndex: 'config_type' },
                  { title: '年化', dataIndex: 'failure_rate', render: (v: number) => formatPercent(v) }
                ]}
              />
            </Card>
          </Col>
          <Col span={8}>
            <Card title="套餐型号故障率 TOP8" style={{ background: '#0b1220', border: '1px solid #334155' }} headStyle={{ color: '#fff' }}>
              <Table
                size="small"
                pagination={false}
                rowKey={(r) => `${r.config_type}-${r.manufacturer}-${r.model}`}
                dataSource={topPackageModel}
                columns={[
                  { title: '套餐', dataIndex: 'config_type' },
                  { title: '年化', dataIndex: 'failure_rate', render: (v: number) => formatPercent(v) }
                ]}
              />
            </Card>
          </Col>
        </Row>
      </Space>
    </div>
  );
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
