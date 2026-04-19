import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Space, Table, Tag, Typography, message } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { getPlan } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { RenewalPlan } from '../types';

const { Text, Title } = Typography;

interface SummaryRow {
  key: string;
  dimension: string;
  target: string;
  coveredValue: string;
  coveredServers: number;
  renewalValue: string;
  renewalServers: number;
  currentValue: string;
  currentServers: number;
}

export default function PlanDetailPage() {
  const navigate = useNavigate();
  const { planId = '' } = useParams();
  const [loading, setLoading] = useState(false);
  const [plan, setPlan] = useState<RenewalPlan | null>(null);

  useEffect(() => {
    if (!planId) return;
    loadPlan(planId);
  }, [planId]);

  async function loadPlan(id: string) {
    setLoading(true);
    try {
      const resp = ensureApiOk(await getPlan(id));
      setPlan(resp.data);
    } catch (e) {
      message.error(parseApiError(e, '查询方案失败'));
    } finally {
      setLoading(false);
    }
  }

  const summaryRows = useMemo(() => (plan ? buildSummaryRows(plan) : []), [plan]);

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

          <Card title="方案总结（目标/保内/续保）" loading={loading}>
            <Table<SummaryRow>
              rowKey="key"
              dataSource={summaryRows}
              pagination={false}
              columns={[
                { title: '维度', dataIndex: 'dimension', width: 120 },
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

          <Card title="续保清单" loading={loading}>
            <Table
              rowKey="sn"
              dataSource={plan.items}
              pagination={withTotalPagination(10)}
              columns={[
                { title: '排名', dataIndex: 'rank', width: 70, render: (v: number) => formatInt(v) },
                { title: '栏目', dataIndex: 'bucket', width: 100 },
                { title: 'SN', dataIndex: 'sn', width: 160 },
                { title: '服务器型号', dataIndex: 'model', width: 160 },
                { title: '配置类型', dataIndex: 'config_type', width: 140 },
                { title: '场景大类', dataIndex: 'scene_category', width: 120 },
                { title: 'CPU核数', dataIndex: 'cpu_logical_cores', width: 100, render: (v: number) => formatInt(v) },
                { title: 'GPU卡数', dataIndex: 'gpu_card_count', width: 100, render: (v: number) => formatInt(v) },
                { title: '存储(TB)', dataIndex: 'storage_capacity_tb', width: 100, render: (v: number) => toTB(v) },
                { title: '最终分', dataIndex: 'final_score', width: 110, render: (v: number) => formatFloat(v) }
              ]}
              scroll={{ x: 1320 }}
            />
          </Card>

          <Card title="不续保清单（含原因）" loading={loading}>
            <Table
              rowKey={(r) => `${r.sn}-${r.reason_code}-${r.rank_in_bucket || 0}`}
              dataSource={plan.non_renewal_items || []}
              pagination={withTotalPagination(10)}
              columns={[
                { title: 'SN', dataIndex: 'sn', width: 160 },
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
      dimension: '计算型',
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
      dimension: '温存储',
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
      dimension: '热存储',
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
      dimension: 'GPU（例外）',
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

function toTB(v: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
