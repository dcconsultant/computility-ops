import { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Empty,
  Input,
  message,
  Space,
  Statistic,
  Table,
  Tabs,
  Typography
} from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { exportPlan, getPlan } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { PlanItem, RenewalPlan, RenewalPlanSection } from '../types';

const { Title, Text } = Typography;

const bucketLabel: Record<string, string> = {
  compute: '计算型',
  warm_storage: '温存储',
  hot_storage: '热存储',
  gpu: 'GPU'
};

export default function ResultPage() {
  const params = useParams();
  const navigate = useNavigate();
  const [planIdInput, setPlanIdInput] = useState(params.planId ?? '');
  const [plan, setPlan] = useState<RenewalPlan | null>(null);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');

  async function fetchPlan(planId: string) {
    if (!planId) return;
    setLoading(true);
    try {
      const resp = ensureApiOk(await getPlan(planId));
      setPlan(resp.data);
      setPlanIdInput(resp.data.plan_id);
    } catch (e) {
      setPlan(null);
      message.error(parseApiError(e, '查询方案失败'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (params.planId) {
      fetchPlan(params.planId);
    }
  }, [params.planId]);

  const filteredItems = useMemo(() => {
    if (!plan) return [] as PlanItem[];
    const kw = keyword.trim().toLowerCase();
    if (!kw) return plan.items;
    return plan.items.filter((x) =>
      x.sn.toLowerCase().includes(kw) || (x.model || '').toLowerCase().includes(kw)
    );
  }, [plan, keyword]);

  const sections = useMemo(() => {
    if (!plan) return [] as RenewalPlanSection[];
    if (plan.sections?.length) return plan.sections;
    return [{ bucket: 'compute', selected_count: plan.items.length, items: plan.items } as RenewalPlanSection];
  }, [plan]);

  const columns = [
    { title: '排名', dataIndex: 'rank', sorter: (a: PlanItem, b: PlanItem) => a.rank - b.rank },
    { title: 'SN', dataIndex: 'sn' },
    { title: '制造商', dataIndex: 'manufacturer' },
    { title: '型号', dataIndex: 'model' },
    { title: '环境', dataIndex: 'environment' },
    { title: '配置类型', dataIndex: 'config_type' },
    { title: 'CPU逻辑核数', dataIndex: 'cpu_logical_cores' },
    { title: '存储容量(TB)', dataIndex: 'storage_capacity_tb' },
    { title: '基础分', dataIndex: 'base_score' },
    { title: 'AFR_old', dataIndex: 'afr_old' },
    { title: 'AFR_avg', dataIndex: 'afr_avg' },
    { title: '故障率系数', dataIndex: 'failure_adjust_factor' },
    {
      title: '最终分',
      dataIndex: 'final_score',
      sorter: (a: PlanItem, b: PlanItem) => a.final_score - b.final_score,
      defaultSortOrder: 'descend' as const
    },
    { title: '特殊策略', dataIndex: 'special_policy' }
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="查询续保结果">
        <Space wrap>
          <Input
            placeholder="输入 plan_id"
            value={planIdInput}
            onChange={(e) => setPlanIdInput(e.target.value)}
            style={{ width: 260 }}
          />
          <Button
            type="primary"
            onClick={() => {
              if (!planIdInput.trim()) {
                message.warning('请输入 plan_id');
                return;
              }
              navigate(`/result/${planIdInput.trim()}`);
            }}
          >
            查询
          </Button>
          {plan && (
            <>
              <Button onClick={() => exportPlan(plan.plan_id, 'xlsx')}>导出 XLSX</Button>
              <Button onClick={() => exportPlan(plan.plan_id, 'csv')}>导出 CSV</Button>
            </>
          )}
        </Space>
      </Card>

      {!plan ? (
        <Card><Empty description={loading ? '加载中...' : '请输入 plan_id 查询'} /></Card>
      ) : (
        <Card title={<Title level={5} style={{ margin: 0 }}>方案详情：{plan.plan_id}</Title>}>
          <Space size={24} wrap style={{ marginBottom: 16 }}>
            <Statistic title="续保目标时间" value={plan.target_date || '-'} />
            <Statistic title="计算型目标核数" value={plan.target_cores} />
            <Statistic title="总已选核数" value={plan.selected_cores} />
            <Statistic title="总已选存储(TB)" value={plan.selected_storage_tb || 0} />
            <Statistic title="总入选台数" value={plan.selected_count} />
          </Space>

          <Text type="secondary">
            排除环境：{(plan.excluded_environments || []).join('、') || '无'}
          </Text>

          <Input.Search
            allowClear
            placeholder="按 SN / 型号 过滤"
            onSearch={setKeyword}
            onChange={(e) => setKeyword(e.target.value)}
            style={{ width: 320, marginTop: 12, marginBottom: 12 }}
          />

          <Tabs
            items={sections.map((sec) => {
              const tableData = sec.items.filter((x) => filteredItems.some((y) => y.sn === x.sn));
              return {
                key: sec.bucket,
                label: bucketLabel[sec.bucket] || sec.bucket,
                children: (
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Space size={24} wrap>
                      <Statistic title="目标核数" value={sec.target_cores || 0} />
                      <Statistic title="目标存储(TB)" value={sec.target_storage_tb || 0} />
                      <Statistic title="已选核数" value={sec.selected_cores || 0} />
                      <Statistic title="已选存储(TB)" value={sec.selected_storage_tb || 0} />
                      <Statistic title="入选台数" value={sec.selected_count || 0} />
                    </Space>
                    <Table
                      rowKey="sn"
                      loading={loading}
                      dataSource={tableData}
                      pagination={{ pageSize: 20 }}
                      columns={columns}
                    />
                  </Space>
                )
              };
            })}
          />
        </Card>
      )}
    </Space>
  );
}
