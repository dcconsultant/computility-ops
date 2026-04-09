import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  DatePicker,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
  message
} from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { createPlan, deletePlan, exportPlan, getPlan, listPlans } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { RenewalPlan } from '../types';

const { Text, Paragraph } = Typography;

export default function PlanPage() {
  const [targetDate, setTargetDate] = useState(dayjs().format('YYYY-MM-DD'));
  const [excludeEnvs, setExcludeEnvs] = useState('开发,测试');
  const [excludePSAs, setExcludePSAs] = useState('');
  const [targetCores, setTargetCores] = useState<number>(1200);
  const [warmTargetStorageTB, setWarmTargetStorageTB] = useState<number>(0);
  const [hotTargetStorageTB, setHotTargetStorageTB] = useState<number>(0);
  const [loading, setLoading] = useState(false);

  const [plans, setPlans] = useState<RenewalPlan[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [viewPlan, setViewPlan] = useState<RenewalPlan | null>(null);
  const [viewOpen, setViewOpen] = useState(false);

  async function reloadPlans() {
    setListLoading(true);
    try {
      const resp = ensureApiOk(await listPlans());
      setPlans(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载历史方案失败'));
    } finally {
      setListLoading(false);
    }
  }

  useEffect(() => {
    reloadPlans();
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
      setViewPlan(resp.data);
      setViewOpen(true);
    } catch (e) {
      message.error(parseApiError(e, '生成失败'));
    } finally {
      setLoading(false);
    }
  }

  async function onView(planId: string) {
    try {
      const resp = ensureApiOk(await getPlan(planId));
      setViewPlan(resp.data);
      setViewOpen(true);
    } catch (e) {
      message.error(parseApiError(e, '查询方案失败'));
    }
  }

  async function onDelete(planId: string) {
    try {
      ensureApiOk(await deletePlan(planId));
      message.success('方案已删除');
      await reloadPlans();
      if (viewPlan?.plan_id === planId) {
        setViewOpen(false);
        setViewPlan(null);
      }
    } catch (e) {
      message.error(parseApiError(e, '删除失败'));
    }
  }

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
    { title: '计算目标核数', dataIndex: 'target_cores', width: 120 },
    { title: '温存储空间需求(TB)', dataIndex: 'warm_target_storage_tb', width: 160 },
    { title: '热存储空间需求(TB)', dataIndex: 'hot_target_storage_tb', width: 160 },
    {
      title: '异常',
      width: 70,
      render: (_: unknown, r: RenewalPlan) => {
        const hasIssue = (r.unmatched_config_count || 0) > 0;
        const report = buildAnomalyReport(r);
        return (
          <Tooltip title={report}>
            <Button
              type={hasIssue ? 'primary' : 'default'}
              danger={hasIssue}
              icon={<ExclamationCircleOutlined />}
              size="small"
              onClick={() => Modal.info({ title: `方案 ${r.plan_id} 异常报告`, width: 760, content: <pre style={{ whiteSpace: 'pre-wrap' }}>{report}</pre> })}
            />
          </Tooltip>
        );
      }
    },
    {
      title: '操作',
      fixed: 'right' as const,
      width: 220,
      render: (_: unknown, r: RenewalPlan) => (
        <Space>
          <Button size="small" onClick={() => onView(r.plan_id)}>查看</Button>
          <Button size="small" onClick={() => exportPlan(r.plan_id, 'xlsx')}>下载</Button>
          <Popconfirm title="确认删除该方案？" onConfirm={() => onDelete(r.plan_id)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const planIssues = useMemo(() => (viewPlan ? buildAnomalyReport(viewPlan) : ''), [viewPlan]);

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

      <Card title="续保管理 - 历史方案列表" extra={<Button onClick={reloadPlans} loading={listLoading}>刷新</Button>}>
        <Table
          rowKey="plan_id"
          loading={listLoading}
          dataSource={plans}
          columns={columns}
          scroll={{ x: 1500 }}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        open={viewOpen}
        title={viewPlan ? `方案详情 ${viewPlan.plan_id}` : '方案详情'}
        footer={null}
        width={1000}
        onCancel={() => setViewOpen(false)}
      >
        {viewPlan && (
          <Space direction="vertical" style={{ width: '100%' }}>
            {(viewPlan.unmatched_config_count || 0) > 0 && (
              <Alert
                type="warning"
                showIcon
                message={`检测到未匹配套餐配置 ${viewPlan.unmatched_config_count} 个`}
                description={(viewPlan.unmatched_config_types || []).join('、')}
              />
            )}
            <Space wrap>
              <Tag>目标时间: {viewPlan.target_date || '-'}</Tag>
              <Tag>计算目标: {viewPlan.target_cores}</Tag>
              <Tag>温存储目标: {viewPlan.warm_target_storage_tb || 0}TB</Tag>
              <Tag>热存储目标: {viewPlan.hot_target_storage_tb || 0}TB</Tag>
              <Tag color="blue">入选台数: {viewPlan.selected_count}</Tag>
            </Space>
            <Paragraph copyable>{planIssues}</Paragraph>
            <Table
              rowKey="sn"
              dataSource={viewPlan.items}
              pagination={{ pageSize: 10 }}
              columns={[
                { title: '排名', dataIndex: 'rank', width: 70 },
                { title: '栏目', dataIndex: 'bucket', width: 100 },
                { title: 'SN', dataIndex: 'sn', width: 150 },
                { title: '服务器型号', dataIndex: 'model', width: 150 },
                { title: '配置类型', dataIndex: 'config_type', width: 150 },
                { title: '最终分', dataIndex: 'final_score', width: 100 }
              ]}
            />
          </Space>
        )}
      </Modal>
    </Space>
  );
}

function buildAnomalyReport(plan: RenewalPlan): string {
  const lines: string[] = [];
  if ((plan.unmatched_config_count || 0) > 0) {
    lines.push(`- 未匹配套餐配置: ${plan.unmatched_config_count} 个`);
    lines.push(`  清单: ${(plan.unmatched_config_types || []).join('、')}`);
  }
  if ((plan.required_compute_cores || 0) > (plan.sections?.find((s) => s.bucket === 'compute')?.selected_cores || 0)) {
    lines.push('- 计算型缺口未补齐');
  }
  const warm = plan.sections?.find((s) => s.bucket === 'warm_storage');
  if ((warm?.required_storage_tb || 0) > (warm?.selected_storage_tb || 0)) {
    lines.push('- 温存储缺口未补齐');
  }
  const hot = plan.sections?.find((s) => s.bucket === 'hot_storage');
  if ((hot?.required_storage_tb || 0) > (hot?.selected_storage_tb || 0)) {
    lines.push('- 热存储缺口未补齐');
  }
  if (!lines.length) {
    lines.push('- 未发现明显异常');
  }
  return lines.join('\n');
}
