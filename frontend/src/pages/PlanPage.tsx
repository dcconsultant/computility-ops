import { useEffect, useState } from 'react';
import {
  Button,
  Card,
  DatePicker,
  Dropdown,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tooltip,
  Typography,
  Upload,
  message
} from 'antd';
import { ExclamationCircleOutlined, UploadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import type { UploadProps } from 'antd';
import { useNavigate } from 'react-router-dom';
import { createPlan, deletePlan, exportPlan, importSpecialRules, listPlans, listSpecialRules } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { RenewalPlan, SpecialRule } from '../types';

const { Text } = Typography;
const { RangePicker } = DatePicker;

export default function PlanPage() {
  const navigate = useNavigate();
  const [targetDate, setTargetDate] = useState(dayjs().format('YYYY-MM-DD'));
  const [excludeEnvs, setExcludeEnvs] = useState('开发,测试');
  const [excludePSAs, setExcludePSAs] = useState('');
  const [targetCores, setTargetCores] = useState<number>(1200);
  const [warmTargetStorageTB, setWarmTargetStorageTB] = useState<number>(0);
  const [hotTargetStorageTB, setHotTargetStorageTB] = useState<number>(0);
  const [loading, setLoading] = useState(false);

  const [plans, setPlans] = useState<RenewalPlan[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [specialRules, setSpecialRules] = useState<SpecialRule[]>([]);
  const [specialLoading, setSpecialLoading] = useState(false);
  const [specialUploading, setSpecialUploading] = useState(false);

  const [queryPlanID, setQueryPlanID] = useState('');
  const [queryTargetDateRange, setQueryTargetDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null);
  const [queryExcludedPSA, setQueryExcludedPSA] = useState('');
  const [queryExcludedEnv, setQueryExcludedEnv] = useState('');

  async function reloadPlans() {
    setListLoading(true);
    try {
      const params = {
        plan_id: queryPlanID.trim() || undefined,
        target_date_from: queryTargetDateRange?.[0]?.format('YYYY-MM-DD') || undefined,
        target_date_to: queryTargetDateRange?.[1]?.format('YYYY-MM-DD') || undefined,
        excluded_psa: queryExcludedPSA.trim() || undefined,
        excluded_environment: queryExcludedEnv.trim() || undefined
      };
      const resp = ensureApiOk(await listPlans(params));
      setPlans(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载历史方案失败'));
    } finally {
      setListLoading(false);
    }
  }

  async function reloadSpecialRules() {
    setSpecialLoading(true);
    try {
      const resp = ensureApiOk(await listSpecialRules());
      setSpecialRules(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载例外清单失败'));
    } finally {
      setSpecialLoading(false);
    }
  }

  useEffect(() => {
    reloadPlans();
    reloadSpecialRules();
  }, []);

  async function onCreatePlan() {
    if (!targetDate) {
      message.warning('请选择续保目标时间');
      return;
    }
    if (!targetCores || targetCores <= 0) {
      message.warning('请输入有效计算型目标核数');
      return;
    }
    if (warmTargetStorageTB < 0 || hotTargetStorageTB < 0) {
      message.warning('温/热存储目标容量不能为负数');
      return;
    }

    setLoading(true);
    try {
      const excluded = excludeEnvs.split(/[，,]/).map((x) => x.trim()).filter(Boolean);
      const excludedPSAList = excludePSAs.split(/[，,]/).map((x) => x.trim()).filter(Boolean);
      const resp = ensureApiOk(await createPlan({
        target_date: targetDate,
        excluded_environments: excluded,
        excluded_psas: excludedPSAList,
        target_cores: targetCores,
        warm_target_storage_tb: warmTargetStorageTB,
        hot_target_storage_tb: hotTargetStorageTB
      }));
      message.success(`方案已生成：${resp.data.plan_id}`);
      await reloadPlans();
      navigate(`/plan/${resp.data.plan_id}`);
    } catch (e) {
      message.error(parseApiError(e, '生成失败'));
    } finally {
      setLoading(false);
    }
  }

  async function onDelete(planId: string) {
    try {
      ensureApiOk(await deletePlan(planId));
      message.success('方案已删除');
      await reloadPlans();
    } catch (e) {
      message.error(parseApiError(e, '删除失败'));
    }
  }

  async function onResetHistoryFilters() {
    setQueryPlanID('');
    setQueryTargetDateRange(null);
    setQueryExcludedPSA('');
    setQueryExcludedEnv('');
    setListLoading(true);
    try {
      const resp = ensureApiOk(await listPlans());
      setPlans(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '重置查询失败'));
    } finally {
      setListLoading(false);
    }
  }

  const specialUploadProps: UploadProps = {
    maxCount: 1,
    showUploadList: true,
    accept: '.xlsx',
    customRequest: async (options) => {
      const file = options.file as File;
      setSpecialUploading(true);
      try {
        const resp = ensureApiOk(await importSpecialRules(file));
        message.success(`例外清单导入完成：成功 ${resp.data.success} 条`);
        await reloadSpecialRules();
        options.onSuccess?.({}, new XMLHttpRequest());
      } catch (e) {
        message.error(parseApiError(e, '导入例外清单失败'));
        options.onError?.(new Error('import failed'));
      } finally {
        setSpecialUploading(false);
      }
    }
  };

  const columns = [
    { title: '方案ID', dataIndex: 'plan_id', width: 140 },
    { title: '续保目标时间', dataIndex: 'target_date', width: 120 },
    {
      title: '排除PSA',
      dataIndex: 'excluded_psas',
      render: (v: string[]) => (v && v.length ? v.join('、') : '-'),
      width: 180
    },
    {
      title: '排除环境',
      dataIndex: 'excluded_environments',
      render: (v: string[]) => (v && v.length ? v.join('、') : '-'),
      width: 160
    },
    { title: '计算目标核数', dataIndex: 'target_cores', width: 120, render: (v: number) => formatInt(v) },
    { title: '温存储空间需求(TB)', dataIndex: 'warm_target_storage_tb', width: 160, render: (v: number) => formatFloat(v) },
    { title: '热存储空间需求(TB)', dataIndex: 'hot_target_storage_tb', width: 160, render: (v: number) => formatFloat(v) },
    {
      title: '摘要',
      width: 260,
      render: (_: unknown, r: RenewalPlan) => {
        const summary = buildPlanSummary(r);
        return (
          <Space direction="vertical" size={2}>
            <Text type={summary.computeRate >= 100 ? undefined : 'warning'}>算力达成: {summary.computeRate.toFixed(1)}%</Text>
            <Text type={summary.warmRate >= 100 ? undefined : 'warning'}>温存储达成: {summary.warmRate.toFixed(1)}%</Text>
            <Text type={summary.hotRate >= 100 ? undefined : 'warning'}>热存储达成: {summary.hotRate.toFixed(1)}%</Text>
            <Text type="secondary">入选台数: {formatInt(r.selected_count || 0)}</Text>
          </Space>
        );
      }
    },
    {
      title: '异常',
      width: 90,
      render: (_: unknown, r: RenewalPlan) => {
        const anomaly = analyzeAnomalies(r);
        const report = buildAnomalyReport(r);
        return (
          <Tooltip title={report}>
            <Button
              type={anomaly.blockers.length ? 'primary' : 'default'}
              danger={anomaly.blockers.length > 0}
              icon={<ExclamationCircleOutlined />}
              size="small"
              onClick={() => Modal.info({ title: `方案 ${r.plan_id} 异常报告`, width: 760, content: <pre style={{ whiteSpace: 'pre-wrap' }}>{report}</pre> })}
            >
              {anomaly.blockers.length + anomaly.warnings.length || 0}
            </Button>
          </Tooltip>
        );
      }
    },
    {
      title: '操作',
      fixed: 'right' as const,
      width: 280,
      render: (_: unknown, r: RenewalPlan) => (
        <Space>
          <Button size="small" onClick={() => navigate(`/plan/${r.plan_id}`)}>查看</Button>
          <Dropdown
            menu={{
              items: [
                { key: 'xlsx', label: '下载 Excel (.xlsx)' },
                { key: 'csv', label: '下载 CSV (.csv)' }
              ],
              onClick: ({ key }) => exportPlan(r.plan_id, key as 'xlsx' | 'csv')
            }}
          >
            <Button size="small">下载</Button>
          </Dropdown>
          <Popconfirm title="确认删除该方案？" onConfirm={() => onDelete(r.plan_id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="续保管理 - 生成方案">
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Space wrap>
            <Text>续保目标时间</Text>
            <DatePicker
              value={targetDate ? dayjs(targetDate) : undefined}
              onChange={(d) => setTargetDate(d ? d.format('YYYY-MM-DD') : '')}
              allowClear={false}
            />
          </Space>

          <Space wrap>
            <Text>排除环境</Text>
            <Input style={{ width: 300 }} value={excludeEnvs} onChange={(e) => setExcludeEnvs(e.target.value)} placeholder="开发,测试" />
            <Text type="secondary">多个环境用逗号分隔</Text>
          </Space>

          <Space wrap>
            <Text>排除PSA</Text>
            <Input style={{ width: 300 }} value={excludePSAs} onChange={(e) => setExcludePSAs(e.target.value)} placeholder="例如：A,B,C 或 10,20" />
            <Text type="secondary">排除的PSA不参与总量和续保清单统计</Text>
          </Space>

          <Space wrap>
            <InputNumber min={1} value={targetCores} onChange={(v) => setTargetCores(v ?? 0)} addonBefore="计算目标核数" />
            <InputNumber min={0} step={1} value={warmTargetStorageTB} onChange={(v) => setWarmTargetStorageTB(v ?? 0)} addonBefore="温存储需求(TB)" />
            <InputNumber min={0} step={1} value={hotTargetStorageTB} onChange={(v) => setHotTargetStorageTB(v ?? 0)} addonBefore="热存储需求(TB)" />
            <Button type="primary" loading={loading} onClick={onCreatePlan}>生成方案</Button>
          </Space>
        </Space>
      </Card>

      <Card
        title="续保管理 - 历史方案列表"
        extra={(
          <Space wrap>
            <Input
              style={{ width: 160 }}
              value={queryPlanID}
              onChange={(e) => setQueryPlanID(e.target.value)}
              placeholder="方案ID"
              allowClear
            />
            <RangePicker
              value={queryTargetDateRange}
              onChange={(v) => setQueryTargetDateRange((v as [dayjs.Dayjs | null, dayjs.Dayjs | null]) || null)}
              allowEmpty={[true, true]}
            />
            <Input
              style={{ width: 140 }}
              value={queryExcludedPSA}
              onChange={(e) => setQueryExcludedPSA(e.target.value)}
              placeholder="排除PSA"
              allowClear
            />
            <Input
              style={{ width: 140 }}
              value={queryExcludedEnv}
              onChange={(e) => setQueryExcludedEnv(e.target.value)}
              placeholder="排除环境"
              allowClear
            />
            <Button type="primary" onClick={reloadPlans} loading={listLoading}>搜索</Button>
            <Button onClick={onResetHistoryFilters} loading={listLoading}>重置</Button>
          </Space>
        )}
      >
        <Table
          rowKey="plan_id"
          loading={listLoading}
          dataSource={plans}
          columns={columns}
          scroll={{ x: 1500 }}
          pagination={withTotalPagination(10)}
        />
      </Card>

      <Card
        title="续保管理 - 例外清单"
        extra={(
          <Space>
            <Button onClick={reloadSpecialRules} loading={specialLoading}>刷新</Button>
            <Upload {...specialUploadProps}>
              <Button icon={<UploadOutlined />} loading={specialUploading}>上传并导入</Button>
            </Upload>
          </Space>
        )}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Text type="secondary">导入模板建议三列：SN、策略（加白/加黑）、原因（可选）。其余字段会自动从服务器管理表按 SN 补全。</Text>
          <Table
            rowKey="sn"
            loading={specialLoading}
            dataSource={specialRules}
            pagination={withTotalPagination(10)}
            columns={[
              { title: 'SN', dataIndex: 'sn', width: 160 },
              { title: '制造商', dataIndex: 'manufacturer', width: 120 },
              { title: '型号', dataIndex: 'model', width: 140 },
              { title: 'PSA', dataIndex: 'psa', width: 100 },
              { title: '套餐', dataIndex: 'package_type', width: 140 },
              { title: '策略', dataIndex: 'policy', width: 100 },
              { title: '原因', dataIndex: 'reason', width: 220 }
            ]}
            scroll={{ x: 1140 }}
          />
        </Space>
      </Card>
    </Space>
  );
}

function withTotalPagination(pageSize: number) {
  return {
    pageSize,
    showTotal: (total: number) => `共${total}条，${Math.ceil(total / pageSize)}页`
  };
}

function buildAnomalyReport(plan: RenewalPlan): string {
  const detail = analyzeAnomalies(plan);
  const lines: string[] = [];

  lines.push('【阻断】');
  if (detail.blockers.length) {
    detail.blockers.forEach((x) => lines.push(`- ${x}`));
  } else {
    lines.push('- 无');
  }

  lines.push('');
  lines.push('【警告】');
  if (detail.warnings.length) {
    detail.warnings.forEach((x) => lines.push(`- ${x}`));
  } else {
    lines.push('- 无');
  }

  lines.push('');
  lines.push(`【结论】阻断 ${detail.blockers.length} 项，警告 ${detail.warnings.length} 项`);
  return lines.join('\n');
}

function analyzeAnomalies(plan: RenewalPlan): { blockers: string[]; warnings: string[] } {
  const blockers: string[] = [];
  const warnings: string[] = [];

  if ((plan.unmatched_config_count || 0) > 0) {
    blockers.push(`未匹配套餐配置: ${plan.unmatched_config_count} 个（${(plan.unmatched_config_types || []).join('、')}）`);
  }

  const computeSelected = plan.sections?.find((s) => s.bucket === 'compute')?.selected_cores || 0;
  if ((plan.required_compute_cores || 0) > computeSelected) {
    blockers.push(`计算型缺口未补齐（要求 ${plan.required_compute_cores || 0}，已选 ${computeSelected}）`);
  }

  const warm = plan.sections?.find((s) => s.bucket === 'warm_storage');
  if ((warm?.required_storage_tb || 0) > (warm?.selected_storage_tb || 0)) {
    blockers.push(`温存储缺口未补齐（要求 ${warm?.required_storage_tb || 0}TB，已选 ${warm?.selected_storage_tb || 0}TB）`);
  }

  const hot = plan.sections?.find((s) => s.bucket === 'hot_storage');
  if ((hot?.required_storage_tb || 0) > (hot?.selected_storage_tb || 0)) {
    blockers.push(`热存储缺口未补齐（要求 ${hot?.required_storage_tb || 0}TB，已选 ${hot?.selected_storage_tb || 0}TB）`);
  }

  if (!plan.items?.length) {
    warnings.push('方案列表为空，请检查输入与过滤条件');
  }

  return { blockers, warnings };
}

function buildPlanSummary(plan: RenewalPlan) {
  const computeRate = calcRate(plan.required_compute_cores || 0, plan.sections?.find((s) => s.bucket === 'compute')?.selected_cores || 0);
  const warm = plan.sections?.find((s) => s.bucket === 'warm_storage');
  const hot = plan.sections?.find((s) => s.bucket === 'hot_storage');
  const warmRate = calcRate(warm?.required_storage_tb || 0, warm?.selected_storage_tb || 0);
  const hotRate = calcRate(hot?.required_storage_tb || 0, hot?.selected_storage_tb || 0);
  return { computeRate, warmRate, hotRate };
}

function calcRate(required: number, selected: number): number {
  if (!required || required <= 0) {
    return 100;
  }
  return Math.min(999.9, (selected / required) * 100);
}

function formatInt(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { maximumFractionDigits: 0 });
}

function formatFloat(v?: number) {
  return Number(v || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
