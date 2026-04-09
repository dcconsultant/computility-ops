import { useState } from 'react';
import { Button, Card, InputNumber, message, Space, Typography } from 'antd';
import { useNavigate } from 'react-router-dom';
import { createPlan } from '../api';
import { ensureApiOk, parseApiError } from '../error';

const { Text } = Typography;

export default function PlanPage() {
  const [targetCores, setTargetCores] = useState<number>(1200);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function onCreatePlan() {
    if (!targetCores || targetCores <= 0) {
      message.warning('请输入有效目标核数');
      return;
    }
    setLoading(true);
    try {
      const resp = ensureApiOk(await createPlan(targetCores));
      message.success(`方案已生成：${resp.data.plan_id}`);
      navigate(`/result/${resp.data.plan_id}`);
    } catch (e) {
      message.error(parseApiError(e, '生成失败'));
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card title="生成续保方案">
      <Space direction="vertical" size="middle">
        <Space>
          <InputNumber
            min={1}
            value={targetCores}
            onChange={(v) => setTargetCores(v ?? 0)}
            addonBefore="目标核数"
          />
          <Button type="primary" loading={loading} onClick={onCreatePlan}>
            生成方案
          </Button>
        </Space>
        <Text type="secondary">系统按（PSA × 架构标准化系数）排名；特殊名单加白强制续保、加黑强制不续保。</Text>
      </Space>
    </Card>
  );
}
