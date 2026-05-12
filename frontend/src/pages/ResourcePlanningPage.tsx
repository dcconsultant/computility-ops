import { useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Divider, Form, Input, InputNumber, Row, Space, Statistic, Table, Tag, Typography, message } from 'antd';
import { DownloadOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons';
import { listHostPackages, listServers } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { HostPackageConfig, ServerItem } from '../types';

const { Text, Title } = Typography;

type Region = 'domestic' | 'india';

interface PlanRow {
  action: '新机采购' | '改配利旧' | '准系统采购利旧' | '续保' | '自维修' | '处置';
  region: Region;
  reason: string;
  estimate_units: number;
  estimate_cores: number;
  estimate_budget: number;
  priority: number;
}

interface SnapshotRecord {
  name: string;
  createdAt: string;
  params: any;
}

const SNAPSHOT_KEY = 'resource_planning_snapshots_v1';

function readSnapshots(): SnapshotRecord[] {
  try {
    const raw = localStorage.getItem(SNAPSHOT_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function writeSnapshots(list: SnapshotRecord[]) {
  localStorage.setItem(SNAPSHOT_KEY, JSON.stringify(list.slice(0, 20)));
}

function exportRowsCSV(rows: PlanRow[]) {
  const headers = ['优先级', '区域', '策略动作', '建议数量', '覆盖算力(核)', '预算影响(元)', '触发原因'];
  const body = rows.map((r) => [
    r.priority,
    r.region === 'domestic' ? '国内' : '印度',
    r.action,
    r.estimate_units,
    r.estimate_cores,
    r.estimate_budget,
    `"${String(r.reason || '').replace(/"/g, '""')}"`
  ]);
  const csv = [headers.join(','), ...body.map((x) => x.join(','))].join('\n');
  const blob = new Blob(['\ufeff' + csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `资源规划建议_${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')}.csv`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export default function ResourcePlanningPage() {
  const [loading, setLoading] = useState(false);
  const [rows, setRows] = useState<PlanRow[]>([]);
  const [summary, setSummary] = useState({
    total_units: 0,
    total_cores: 0,
    total_budget: 0,
    risk_count: 0
  });
  const [lastInput, setLastInput] = useState<any>(null);
  const [form] = Form.useForm();
  const [snapshotName, setSnapshotName] = useState('');
  const [snapshots, setSnapshots] = useState<SnapshotRecord[]>(() => readSnapshots());

  async function generatePlan(values: any) {
    setLoading(true);
    try {
      const [s1, s2] = await Promise.all([listServers(), listHostPackages()]);
      const servers = (ensureApiOk(s1).data.list || []) as ServerItem[];
      const packages = (ensureApiOk(s2).data.list || []) as HostPackageConfig[];

      const byRegion = {
        domestic: servers.filter((s) => !/india/i.test(s.psa || '')),
        india: servers.filter((s) => /india/i.test(s.psa || ''))
      };

      const avgCore = Math.max(1, Math.round((packages.reduce((acc, p) => acc + Number(p.cpu_logical_cores || 0), 0) / Math.max(1, packages.length)) || 64));

      const planRows: PlanRow[] = [];
      (['domestic', 'india'] as Region[]).forEach((region, idx) => {
        const targetCores = Number(values[`${region}_target_cores`] || 0);
        const budget = Number(values[`${region}_budget`] || 0);
        const pool = byRegion[region];
        const poolSize = pool.length;

        const keepRatio = Number(values.keep_under_5y_ratio || 0.9);
        const buyRatio = Math.min(0.25, Number(values.new_purchase_ratio || 0.1));
        const newUnits = Math.max(0, Math.floor((targetCores * buyRatio) / avgCore));
        const retrofitUnits = Math.max(0, Math.floor(poolSize * 0.08));
        const renewUnits = Math.max(0, Math.floor(poolSize * keepRatio * 0.12));
        const selfRepairUnits = Math.max(0, Math.floor(poolSize * 0.05));
        const disposeUnits = Math.max(0, Math.floor(poolSize * (1 - keepRatio) * 0.2));

        const unitNewCost = Number(values.new_unit_cost || 38000);
        const unitRetrofitCost = Number(values.retrofit_unit_cost || 8000);
        const unitRenewCost = Number(values.renew_unit_cost || 6000);
        const unitRepairCost = Number(values.repair_unit_cost || 3000);

        planRows.push(
          {
            action: '新机采购',
            region,
            reason: `按需求算力的 ${(buyRatio * 100).toFixed(0)}% 估算，遵循≤25%上限`,
            estimate_units: newUnits,
            estimate_cores: newUnits * avgCore,
            estimate_budget: newUnits * unitNewCost,
            priority: 1 + idx
          },
          {
            action: '改配利旧',
            region,
            reason: '优先改配可达标机器，补齐中短期缺口',
            estimate_units: retrofitUnits,
            estimate_cores: retrofitUnits * Math.floor(avgCore * 0.6),
            estimate_budget: retrofitUnits * unitRetrofitCost,
            priority: 3 + idx
          },
          {
            action: '准系统采购利旧',
            region,
            reason: '当高价值配件库存充足且成本低于新机阈值时触发',
            estimate_units: Math.max(0, Math.floor(newUnits * 0.35)),
            estimate_cores: Math.max(0, Math.floor(newUnits * 0.35)) * Math.floor(avgCore * 0.8),
            estimate_budget: Math.max(0, Math.floor(newUnits * 0.35)) * Math.floor(unitNewCost * 0.65),
            priority: 5 + idx
          },
          {
            action: '续保',
            region,
            reason: '非开测场景按价值分优先，覆盖稳态需求',
            estimate_units: renewUnits,
            estimate_cores: renewUnits * Math.floor(avgCore * 0.5),
            estimate_budget: renewUnits * unitRenewCost,
            priority: 7 + idx
          },
          {
            action: '自维修',
            region,
            reason: '开测环境优先，受备件可得性与成功率约束',
            estimate_units: selfRepairUnits,
            estimate_cores: selfRepairUnits * Math.floor(avgCore * 0.4),
            estimate_budget: selfRepairUnits * unitRepairCost,
            priority: 9 + idx
          },
          {
            action: '处置',
            region,
            reason: '不可修复或无改配价值设备拆配后处置',
            estimate_units: disposeUnits,
            estimate_cores: 0,
            estimate_budget: -disposeUnits * Math.floor(unitRetrofitCost * 0.35),
            priority: 11 + idx
          }
        );

        const usedBudget = planRows.filter((r) => r.region === region).reduce((a, b) => a + b.estimate_budget, 0);
        if (usedBudget > budget) {
          message.warning(`${region === 'domestic' ? '国内' : '印度'}预算可能不足：预计 ${usedBudget.toLocaleString()} > 预算 ${budget.toLocaleString()}`);
        }
      });

      const total_units = planRows.reduce((a, b) => a + b.estimate_units, 0);
      const total_cores = planRows.reduce((a, b) => a + b.estimate_cores, 0);
      const total_budget = planRows.reduce((a, b) => a + b.estimate_budget, 0);
      const risk_count = Number(total_budget > (Number(values.domestic_budget || 0) + Number(values.india_budget || 0))) + Number((values.new_purchase_ratio || 0.1) > 0.25);

      setRows(planRows.sort((a, b) => a.priority - b.priority));
      setSummary({ total_units, total_cores, total_budget, risk_count });
      setLastInput(values);
    } catch (e) {
      message.error(parseApiError(e, '资源规划生成失败'));
    } finally {
      setLoading(false);
    }
  }

  const regionStats = useMemo(() => {
    return {
      domestic: rows.filter((r) => r.region === 'domestic').reduce((a, b) => a + b.estimate_budget, 0),
      india: rows.filter((r) => r.region === 'india').reduce((a, b) => a + b.estimate_budget, 0)
    };
  }, [rows]);

  function saveSnapshot() {
    const params = form.getFieldsValue();
    const name = (snapshotName || '').trim() || `快照-${new Date().toLocaleString()}`;
    const next: SnapshotRecord[] = [{ name, createdAt: new Date().toISOString(), params }, ...snapshots];
    writeSnapshots(next);
    setSnapshots(next.slice(0, 20));
    setSnapshotName('');
    message.success('参数快照已保存');
  }

  function loadSnapshot(item: SnapshotRecord) {
    form.setFieldsValue(item.params || {});
    message.success(`已加载快照：${item.name}`);
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card>
        <Title level={4} style={{ marginTop: 0 }}>资源规划（首版）</Title>
        <Text type="secondary">依据你给的文档先做可执行骨架：参数配置 → 方案生成 → 风险提示。后续规则细化可直接迭代到同页面。</Text>
      </Card>

      <Card title="规划输入条件">
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            domestic_target_cores: 200000,
            india_target_cores: 50000,
            domestic_budget: 18000000,
            india_budget: 4500000,
            new_purchase_ratio: 0.1,
            keep_under_5y_ratio: 0.9,
            new_unit_cost: 38000,
            retrofit_unit_cost: 8000,
            renew_unit_cost: 6000,
            repair_unit_cost: 3000
          }}
          onFinish={generatePlan}
        >
          <Row gutter={16}>
            <Col span={6}><Form.Item name="domestic_target_cores" label="国内目标算力(核)" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="india_target_cores" label="印度目标算力(核)" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="domestic_budget" label="国内预算(元)" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="india_budget" label="印度预算(元)" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>

            <Col span={6}><Form.Item name="new_purchase_ratio" label="新机采购比例(≤0.25)"><InputNumber min={0} max={0.25} step={0.01} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="keep_under_5y_ratio" label="5年内保留比例"><InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="new_unit_cost" label="新机单台成本"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="retrofit_unit_cost" label="改配单台成本"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>

            <Col span={6}><Form.Item name="renew_unit_cost" label="续保单台成本"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="repair_unit_cost" label="自维修单台成本"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
          <Space wrap>
            <Button type="primary" htmlType="submit" loading={loading}>生成规划方案</Button>
            <Button icon={<ReloadOutlined />} onClick={() => form.resetFields()}>重置参数</Button>
            <Input
              placeholder="快照名称（可选）"
              value={snapshotName}
              onChange={(e) => setSnapshotName(e.target.value)}
              style={{ width: 220 }}
            />
            <Button icon={<SaveOutlined />} onClick={saveSnapshot}>保存参数快照</Button>
          </Space>

          {snapshots.length > 0 ? (
            <Space direction="vertical" size={6} style={{ marginTop: 12, width: '100%' }}>
              <Text type="secondary">参数快照（最近20条）</Text>
              <Space wrap>
                {snapshots.map((s) => (
                  <Button key={`${s.name}-${s.createdAt}`} size="small" onClick={() => loadSnapshot(s)}>
                    {s.name}
                  </Button>
                ))}
              </Space>
            </Space>
          ) : null}
        </Form>
      </Card>

      <Row gutter={16}>
        <Col span={6}><Card><Statistic title="动作总数量" value={summary.total_units} /></Card></Col>
        <Col span={6}><Card><Statistic title="预计覆盖算力(核)" value={summary.total_cores} /></Card></Col>
        <Col span={6}><Card><Statistic title="预计净预算(元)" value={summary.total_budget} precision={0} /></Card></Col>
        <Col span={6}><Card><Statistic title="风险提示数" value={summary.risk_count} /></Card></Col>
      </Row>

      {lastInput ? (
        <Alert
          type="info"
          showIcon
          message="预算对照"
          description={`国内预计 ${regionStats.domestic.toLocaleString()} / 预算 ${(Number(lastInput.domestic_budget || 0)).toLocaleString()}；印度预计 ${regionStats.india.toLocaleString()} / 预算 ${(Number(lastInput.india_budget || 0)).toLocaleString()}`}
        />
      ) : null}

      <Card
        title="策略动作建议（可复盘草案）"
        extra={<Button icon={<DownloadOutlined />} onClick={() => exportRowsCSV(rows)} disabled={!rows.length}>导出建议CSV</Button>}
      >
        <Table
          rowKey={(r) => `${r.region}-${r.action}-${r.priority}`}
          dataSource={rows}
          pagination={false}
          columns={[
            { title: '优先级', dataIndex: 'priority', width: 88 },
            {
              title: '区域', dataIndex: 'region', width: 100,
              render: (v: Region) => v === 'domestic' ? <Tag color="blue">国内</Tag> : <Tag color="gold">印度</Tag>
            },
            { title: '策略动作', dataIndex: 'action', width: 120 },
            { title: '建议数量', dataIndex: 'estimate_units', width: 110 },
            { title: '覆盖算力(核)', dataIndex: 'estimate_cores', width: 130 },
            { title: '预算影响(元)', dataIndex: 'estimate_budget', width: 140, render: (v: number) => v.toLocaleString() },
            { title: '触发原因', dataIndex: 'reason' }
          ]}
        />

        <Divider />
        <Text type="secondary">说明：本版先按文档规则做“可执行前端估算”。后续可接入后端策略引擎与审批流，把“参数快照、版本号、执行回写”补齐。</Text>
      </Card>
    </Space>
  );
}
