import { useState } from 'react';
import { Button, Card, DatePicker, Input, InputNumber, message, Space, Typography } from 'antd';
import dayjs from 'dayjs';
import { useNavigate } from 'react-router-dom';
import { createPlan } from '../api';
import { ensureApiOk, parseApiError } from '../error';

const { Text } = Typography;

export default function PlanPage() {
  const [targetDate, setTargetDate] = useState(dayjs().format('YYYY-MM-DD'));
  const [excludeEnvs, setExcludeEnvs] = useState('开发,测试');
  const [excludePSAs, setExcludePSAs] = useState('');
  const [targetCores, setTargetCores] = useState<number>(1200);
  const [warmTargetStorageTB, setWarmTargetStorageTB] = useState<number>(0);
  const [hotTargetStorageTB, setHotTargetStorageTB] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

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
      const excluded = excludeEnvs
        .split(/[，,]/)
        .map((x) => x.trim())
        .filter(Boolean);
      const excludedPSAList = excludePSAs
        .split(/[，,]/)
        .map((x) => x.trim())
        .filter(Boolean);
      const resp = ensureApiOk(await createPlan({
        target_date: targetDate,
        excluded_environments: excluded,
        excluded_psas: excludedPSAList,
        target_cores: targetCores,
        warm_target_storage_tb: warmTargetStorageTB,
        hot_target_storage_tb: hotTargetStorageTB
      }));
      message.success(`方案已生成：${resp.data.plan_id}`);
      navigate(`/result/${resp.data.plan_id}`);
    } catch (e) {
      message.error(parseApiError(e, '生成失败'));
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card title="生成续保方案（四栏目）">
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
          <Input
            style={{ width: 320 }}
            value={excludeEnvs}
            onChange={(e) => setExcludeEnvs(e.target.value)}
            placeholder="开发,测试"
          />
          <Text type="secondary">多个环境用逗号分隔</Text>
        </Space>

        <Space wrap>
          <Text>排除PSA</Text>
          <Input
            style={{ width: 320 }}
            value={excludePSAs}
            onChange={(e) => setExcludePSAs(e.target.value)}
            placeholder="例如：A,B,C 或 10,20"
          />
          <Text type="secondary">排除的PSA不参与总量和续保清单统计</Text>
        </Space>

        <Space wrap>
          <InputNumber
            min={1}
            value={targetCores}
            onChange={(v) => setTargetCores(v ?? 0)}
            addonBefore="计算型目标核数"
          />
          <InputNumber
            min={0}
            step={1}
            value={warmTargetStorageTB}
            onChange={(v) => setWarmTargetStorageTB(v ?? 0)}
            addonBefore="温存储目标容量(TB)"
          />
          <InputNumber
            min={0}
            step={1}
            value={hotTargetStorageTB}
            onChange={(v) => setHotTargetStorageTB(v ?? 0)}
            addonBefore="热存储目标容量(TB)"
          />
        </Space>

        <Button type="primary" loading={loading} onClick={onCreatePlan}>
          生成方案
        </Button>

        <Text type="secondary">
          分类依据：主机套餐配置的“场景大类”；GPU 全部续保。温/热存储按容量规划，
          评分按基础分直接除以套餐故障率（AFR_avg）。
        </Text>
      </Space>
    </Card>
  );
}
