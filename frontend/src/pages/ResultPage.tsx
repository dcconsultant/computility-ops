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
  Typography
} from 'antd';
import { useNavigate, useParams } from 'react-router-dom';
import { exportPlan, getPlan } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { PlanItem, RenewalPlan } from '../types';

const { Title } = Typography;

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

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="查询续保结果">
        <Space>
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
            <Statistic title="目标核数" value={plan.target_cores} />
            <Statistic title="已选核数" value={plan.selected_cores} />
            <Statistic title="入选台数" value={plan.selected_count} />
          </Space>

          <Input.Search
            allowClear
            placeholder="按 SN / 型号 过滤"
            onSearch={setKeyword}
            onChange={(e) => setKeyword(e.target.value)}
            style={{ width: 320, marginBottom: 12 }}
          />

          <Table
            rowKey="sn"
            loading={loading}
            dataSource={filteredItems}
            pagination={{ pageSize: 20 }}
            columns={[
              { title: '排名', dataIndex: 'rank', sorter: (a, b) => a.rank - b.rank },
              { title: 'SN', dataIndex: 'sn' },
              { title: '制造商', dataIndex: 'manufacturer' },
              { title: '型号', dataIndex: 'model' },
              { title: '配置类型', dataIndex: 'config_type' },
              { title: 'CPU逻辑核数', dataIndex: 'cpu_logical_cores' },
              { title: 'PSA', dataIndex: 'psa' },
              { title: '架构标准化系数', dataIndex: 'arch_standardized_factor' },
              {
                title: '最终分',
                dataIndex: 'final_score',
                sorter: (a, b) => a.final_score - b.final_score,
                defaultSortOrder: 'descend'
              },
              { title: '特殊策略', dataIndex: 'special_policy' }
            ]}
          />
        </Card>
      )}
    </Space>
  );
}
