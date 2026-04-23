import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { exportNonRenewalPlan, getPlan, listRenewalUnitPrices } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { RenewalPlan, RenewalUnitPrice } from '../types';

const { Text, Title } = Typography;

interface SummaryRow {
  key: string;
  sceneCategory: string;
  target: string;
  coveredValue: string;
  coveredServers: number;
  renewalValue: string;
  renewalServers: number;
  currentValue: string;
  currentServers: number;
}

interface RegionSummaryRow {
  key: string;
  region: string;
  renewalCount: number;
  estimatedCost: number;
  activeRatio: number;
  budget: number;
  budgetExecutionRate: number | null;
}

interface CostRow {
  key: string;
  region: string;
  bucket: string;
  unitPrice: number;
  servers: number;
  amount: number;
}

export default function PlanDetailPage() {
  const navigate = useNavigate();
  const { planId = '' } = useParams();
  const [loading, setLoading] = useState(false);
  const [plan, setPlan] = useState<RenewalPlan | null>(null);
  const [unitPrices, setUnitPrices] = useState<RenewalUnitPrice[]>([]);

  useEffect(() => {
    if (!planId) return;
    loadPlan(planId);
  }, [planId]);

  async function loadPlan(id: string) {
    setLoading(true);
    try {
      const [planResp, unitPriceResp] = await Promise.all([getPlan(id), listRenewalUnitPrices()]);
      setPlan(ensureApiOk(planResp).data);
      setUnitPrices(ensureApiOk(unitPriceResp).data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '查询方案失败'));
    } finally {
      setLoading(false);
    }
  }

  const summaryRows = useMemo(() => (plan ? buildSummaryRows(plan) : []), [plan]);
  const costRows = useMemo(() => (plan ? buildCostRows(plan, unitPrices) : []), [plan, unitPrices]);
  const totalRenewalAmount = useMemo(() => costRows.reduce((sum, r) => sum + r.amount, 0), [costRows]);
  const regionSummaryRows = useMemo(() => (plan ? buildRegionSummaryRows(plan, costRows) : []), [plan, costRows]);

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Space style={{ justifyContent: 'space-between', width: '100%' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/plan')}>返回续保管理</Button>
          <Title level={4} style={{ margin: 0 }}>方案详情 {planId}</Title>
        </Space>
      </Space>

      {!plan && !loading && (
        <Alert type="warning" showIcon message="未找到方案" description="请返回列表重新选择方案。" />
      )}

      {plan && (
        <>
          <Card title="基础信息" loading={loading}>
            <Space wrap>
              <Tag>目标时间: {plan.target_date || '-'}</Tag>
              <Tag>计算目标: {formatInt(plan.target_cores)} </Tag>
              <Tag>温存储目标: {toTB(plan.warm_target_storage_tb || 0)}TB</Tag>
              <Tag>热存储目标: {toTB(plan.hot_target_storage_tb || 0)}TB</Tag>
              <Tag color="blue">入选台数: {formatInt(plan.selected_count)}</Tag>
            </Space>
            <div style={{ marginTop: 12 }}>
              <Text type="secondary">异常详情已收敛到列表页“异常”列提示，本页仅展示方案汇总与明细。</Text>
            </div>
          </Card>

          <Tabs
            items={[
              {
                key: 'overview',
                label: '方案总览',
                children: (
                  <Space direction="vertical" size="large" style={{ width: '100%' }}>
                    <Card title="续保总览-算力" loading={loading}>
                      <Table<SummaryRow>
                        rowKey="key"
                        dataSource={summaryRows}
                        pagination={false}
                        columns={[
                          { title: '场景大类', dataIndex: 'sceneCategory', width: 120 },
                          { title: '目标', dataIndex: 'target', width: 120 },
                          { title: '保内满足', dataIndex: 'coveredValue', width: 130 },
                          { title: '保内台数', dataIndex: 'coveredServers', width: 100, render: (v: number) => formatInt(v) },
                          { title: '续保满足', dataIndex: 'renewalValue', width: 130 },
                          { title: '续保台数', dataIndex: 'renewalServers', width: 100, render: (v: number) => formatInt(v) },
                          { title: '当前总量', dataIndex: 'currentValue', width: 130 },
                          { title: '当前台数', dataIndex: 'currentServers', width: 100, render: (v: number) => formatInt(v) }
                        ]}
                      />
                    </Card>

                    <Card title="续保总览" loading={loading}>
                      <Table<RegionSummaryRow>
                        rowKey="key"
                        dataSource={regionSummaryRows}
                        pagination={false}
                        columns={[
                          { title: '地域', dataIndex: 'region', width: 140 },
                          { title: '续保数量', dataIndex: 'renewalCount', width: 120, render: (v: number) => formatInt(v) },
                          { title: '预估花费(CNY)', dataIndex: 'estimatedCost', width: 140, render: (v: number) => formatMoney(v) },
                          { title: '在役占比', dataIndex: 'activeRatio', width: 120, render: (v: number) => formatPercent(v) },
                          { title: '26预算', dataIndex: 'budget', width: 120, render: (v: number) => formatMoney(v) },
                          {
                            title: '预算执行率',
                            dataIndex: 'budgetExecutionRate',
                            width: 130,
                            render: (v: number | null) => (v == null ? '-' : formatPercent(v))
                          }
                        ]}
                      />
                    </Card>

                    <Card title="续保金额估算（按机房区分印度/国内）" loading={loading}>
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Text type="secondary">规则：机房以 IN 开头归类印度，其余归类国内；按“国家+场景大类”读取续保管理中的最新单价。</Text>
                        <Text strong>续保总金额：{formatMoney(totalRenewalAmount)}</Text>
                        <Table<CostRow>
                          rowKey="key"
                          dataSource={costRows}
                          pagination={false}
                          columns={[
                            { title: '区域', dataIndex: 'region', width: 120 },
                            { title: '栏目', dataIndex: 'bucket', width: 120 },
                            { title: '单价', dataIndex: 'unitPrice', width: 120, render: (v: number) => formatMoney(v) },
                            { title: '台数', dataIndex: 'servers', width: 100, render: (v: number) => formatInt(v) },
                            { title: '金额', dataIndex: 'amount', width: 140, render: (v: number) => formatMoney(v) }
                          ]}
                        />
                      </Space>
                    </Card>
                  </Space>
                )
              },
              {
                key: 'renewal',
                label: `续保清单（${formatInt((plan.items || []).length)}）`,
                children: (
                  <Card title="续保清单" loading={loading}>
                    <Table
                      rowKey="sn"
                      dataSource={plan.items}
                      pagination={withTotalPagination(10)}
                      columns={[
                        { title: '排名', dataIndex: 'rank', width: 70, render: (v: number) => formatInt(v) },
                        { title: '栏目', dataIndex: 'bucket', width: 100 },
                        { title: 'SN', dataIndex: 'sn', width: 160 },
                        { title: '机房', dataIndex: 'idc', width: 110 },
                        { title: '服务器型号', dataIndex: 'model', width: 160 },
                        { title: '详细配置', dataIndex: 'detailed_config', width: 220, ellipsis: true },
                        { title: '配置类型', dataIndex: 'config_type', width: 140 },
                        { title: '场景大类', dataIndex: 'scene_category', width: 120 },
                        { title: 'CPU核数', dataIndex: 'cpu_logical_cores', width: 100, render: (v: number) => formatInt(v) },
                        { title: 'GPU卡数', dataIndex: 'gpu_card_count', width: 100, render: (v: number) => formatInt(v) },
                        { title: '存储(TB)', dataIndex: 'storage_capacity_tb', width: 100, render: (v: number) => toTB(v) },
                        { title: '最近1年故障率', dataIndex: 'recent_1y_failure_rate', width: 130, render: (v: number) => formatPercent(v) },
                        { title: '最终分', dataIndex: 'final_score', width: 110, render: (v: number) => formatFloat(v) }
                      ]}
                      scroll={{ x: 1560 }}
                    />
                  </Card>
                )
              },
              {
                key: 'non-renewal',
                label: `不续保清单（${formatInt((plan.non_renewal_items || []).length)}）`,
                children: (
                  <Card
                    title="不续保清单（含原因）"
                    loading={loading}
                    extra={<Button onClick={() => exportNonRenewalPlan(plan.plan_id)}>下载Excel</Button>}
                  >
                    <Table
                      rowKey={(r) => `${r.sn}-${r.reason_code}-${r.rank_in_bucket || 0}`}
                      dataSource={plan.non_renewal_items || []}
                      pagination={withTotalPagination(10)}
                      columns={[
                        { title: 'SN', dataIndex: 'sn', width: 160 },
                        { title: '机房', dataIndex: 'idc', width: 110 },
                        { title: '栏目', dataIndex: 'bucket', width: 110 },
                        { title: '配置类型', dataIndex: 'config_type', width: 140 },
                        { title: 'PSA', dataIndex: 'psa', width: 120, render: (v?: string) => (v && v.trim() ? v : '-') },
                        { title: '价值分', dataIndex: 'final_score', width: 110, render: (v: number) => formatFloat(v) },
                        { title: '桶内排名', dataIndex: 'rank_in_bucket', width: 100, render: (v: number) => (v ? formatInt(v) : '-') },
                        { title: '不续保理由', dataIndex: 'reason', width: 140 },
                        { title: '原因说明', dataIndex: 'reason_detail', width: 260 }
                      ]}
                      scroll={{ x: 1300 }}
                    />
                  </Card>
                )
              }
            ]}
          />
        </>
      )}
    </Space>
  );
}

function withTotalPagination(pageSize: number) {
  return {
    pageSize,
    showTotal: (total: number) => `共${total}条，${Math.ceil(total / pageSize)}页`
  };
}

function buildSummaryRows(plan: RenewalPlan): SummaryRow[] {
  const compute = plan.sections?.find((s) => s.bucket === 'compute');
  const warm = plan.sections?.find((s) => s.bucket === 'warm_storage');
  const hot = plan.sections?.find((s) => s.bucket === 'hot_storage');

  const computeCovered = compute?.covered_cores || 0;
  const computeRenewal = compute?.selected_cores || 0;
  const computeCoveredServers = compute?.covered_count || 0;
  const computeRenewalServers = compute?.selected_count || 0;

  const warmCovered = warm?.covered_storage_tb || 0;
  const warmRenewal = warm?.selected_storage_tb || 0;
  const warmCoveredServers = warm?.covered_count || 0;
  const warmRenewalServers = warm?.selected_count || 0;

  const hotCovered = hot?.covered_storage_tb || 0;
  const hotRenewal = hot?.selected_storage_tb || 0;
  const hotCoveredServers = hot?.covered_count || 0;
  const hotRenewalServers = hot?.selected_count || 0;

  return [
    {
      key: 'compute',
      sceneCategory: '计算',
      target: `${formatInt(plan.target_cores || 0)} 核`,
      coveredValue: `${formatInt(computeCovered)} 核`,
      coveredServers: computeCoveredServers,
      renewalValue: `${formatInt(computeRenewal)} 核`,
      renewalServers: computeRenewalServers,
      currentValue: `${formatInt(computeCovered + computeRenewal)} 核`,
      currentServers: computeCoveredServers + computeRenewalServers
    },
    {
      key: 'warm_storage',
      sceneCategory: '温存储',
      target: `${plan.warm_target_storage_tb || 0} TB`,
      coveredValue: `${toTB(warmCovered)} TB`,
      coveredServers: warmCoveredServers,
      renewalValue: `${toTB(warmRenewal)} TB`,
      renewalServers: warmRenewalServers,
      currentValue: `${toTB(warmCovered + warmRenewal)} TB`,
      currentServers: warmCoveredServers + warmRenewalServers
    },
    {
      key: 'hot_storage',
      sceneCategory: '热存储',
      target: `${plan.hot_target_storage_tb || 0} TB`,
      coveredValue: `${toTB(hotCovered)} TB`,
      coveredServers: hotCoveredServers,
      renewalValue: `${toTB(hotRenewal)} TB`,
      renewalServers: hotRenewalServers,
      currentValue: `${toTB(hotCovered + hotRenewal)} TB`,
      currentServers: hotCoveredServers + hotRenewalServers
    },
    {
      key: 'gpu',
      sceneCategory: 'GPU',
      target: '-',
      coveredValue: `${formatInt(plan.gpu_covered_cards || 0)} 卡`,
      coveredServers: plan.gpu_covered_servers || 0,
      renewalValue: `${formatInt(plan.gpu_renewal_cards || 0)} 卡`,
      renewalServers: plan.gpu_renewal_servers || 0,
      currentValue: `${formatInt(plan.gpu_current_cards || 0)} 卡`,
      currentServers: plan.gpu_current_servers || 0
    }
  ];
}

function buildRegionSummaryRows(plan: RenewalPlan, costRows: CostRow[]): RegionSummaryRow[] {
  const domesticRenewalCount = (plan.items || []).filter((x) => !isIndiaIDC(x.idc)).length;
  const indiaRenewalCount = (plan.items || []).filter((x) => isIndiaIDC(x.idc)).length;
  const domesticEstimatedCost = costRows.filter((x) => x.region === '国内').reduce((s, x) => s + x.amount, 0);
  const indiaEstimatedCost = costRows.filter((x) => x.region === '印度').reduce((s, x) => s + x.amount, 0);

  const domesticTotal = Number(plan.domestic_servers_no_psa || 0);
  const indiaTotal = Number(plan.india_servers_no_psa || 0);
  const total = Number(plan.total_servers_no_psa || (domesticTotal + indiaTotal));
  const domesticBudget = Number(plan.domestic_budget || 0);
  const indiaBudget = Number(plan.india_budget || 0);

  const rows: RegionSummaryRow[] = [
    {
      key: 'domestic',
      region: '国内',
      renewalCount: domesticRenewalCount,
      estimatedCost: domesticEstimatedCost,
      activeRatio: domesticTotal > 0 ? domesticRenewalCount / domesticTotal : 0,
      budget: domesticBudget,
      budgetExecutionRate: domesticBudget > 0 ? domesticEstimatedCost / domesticBudget : null
    },
    {
      key: 'india',
      region: '印度',
      renewalCount: indiaRenewalCount,
      estimatedCost: indiaEstimatedCost,
      activeRatio: indiaTotal > 0 ? indiaRenewalCount / indiaTotal : 0,
      budget: indiaBudget,
      budgetExecutionRate: indiaBudget > 0 ? indiaEstimatedCost / indiaBudget : null
    }
  ];

  const totalBudget = domesticBudget + indiaBudget;
  rows.push({
    key: 'total',
    region: '合计',
    renewalCount: domesticRenewalCount + indiaRenewalCount,
    estimatedCost: domesticEstimatedCost + indiaEstimatedCost,
    activeRatio: total > 0 ? (domesticRenewalCount + indiaRenewalCount) / total : 0,
    budget: totalBudget,
    budgetExecutionRate: totalBudget > 0 ? (domesticEstimatedCost + indiaEstimatedCost) / totalBudget : null
  });
  return rows;
}

function buildCostRows(plan: RenewalPlan, unitPrices: RenewalUnitPrice[]): CostRow[] {
  const bucketLabel: Record<string, string> = {
    compute: '计算型',
    warm_storage: '温存储',
    hot_storage: '热存储',
    gpu: 'GPU'
  };
  const priceMap = new Map(unitPrices.map((p) => [`${p.country}|${p.scene_category}`, Number(p.unit_price || 0)]));
  const amountByKey = new Map<string, CostRow>();
  for (const item of plan.items || []) {
    const region = isIndiaIDC(item.idc) ? '印度' : '国内';
    const bucket = item.bucket || 'compute';
    const unitPrice = estimateUnitPrice(priceMap, region, bucket);
    const key = `${region}|${bucket}`;
    const old = amountByKey.get(key) || {
      key,
      region,
      bucket: bucketLabel[bucket] || bucket,
      unitPrice,
      servers: 0,
      amount: 0
    };
    old.servers += 1;
    old.amount += unitPrice;
    amountByKey.set(key, old);
  }
  const rows = Array.from(amountByKey.values());
  rows.sort((a, b) => {
    if (a.region !== b.region) return a.region.localeCompare(b.region);
    return a.bucket.localeCompare(b.bucket);
  });
  return rows;
}

function isIndiaIDC(idc?: string) {
  const v = (idc || '').trim().toUpperCase();
  return v.startsWith('IN');
}

function estimateUnitPrice(priceMap: Map<string, number>, region: string, bucket: string) {
  const key = `${region}|${bucket}`;
  return Number(priceMap.get(key) || 0);
}

function toTB(v: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatMoney(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 });
}

function formatPercent(v?: number) {
  return `${(Number(v || 0) * 100).toFixed(2)}%`;
}
