import { useEffect, useState } from 'react';
import { Alert, Button, Card, Col, Form, Input, InputNumber, Row, Space, Statistic, Typography, message } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { calculateResourcePlanning, getResourcePlanningConfig, saveResourcePlanningConfig } from '../api';
import { ensureApiOk, parseApiError } from '../error';
import type { ResourcePlanningResponse } from '../types';

const { Title, Text } = Typography;

const thousandFormatter = (value?: string | number) => {
  const s = String(value ?? '');
  if (!s) return '';
  const neg = s.startsWith('-') ? '-' : '';
  const body = neg ? s.slice(1) : s;
  const [intPart, decPart] = body.split('.');
  const formattedInt = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  return `${neg}${decPart !== undefined ? `${formattedInt}.${decPart}` : formattedInt}`;
};

const thousandParser = (value?: string) => (value || '').replace(/,/g, '');

const numberProps = {
  style: { width: '100%' },
  formatter: thousandFormatter,
  parser: thousandParser
} as const;

export default function ResourcePlanningPage() {
  const [loading, setLoading] = useState(false);
  const [configLoading, setConfigLoading] = useState(false);
  const [result, setResult] = useState<ResourcePlanningResponse | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    (async () => {
      setConfigLoading(true);
      try {
        const resp = await getResourcePlanningConfig();
        const data = ensureApiOk(resp).data;
        if (data?.found && data.state?.config) {
          form.setFieldsValue(data.state.config);
        }
      } catch {
        // ignore
      } finally {
        setConfigLoading(false);
      }
    })();
  }, [form]);

  async function onSubmit(values: any) {
    setLoading(true);
    try {
      const payload: any = { ...values };
      if (payload.reconfig_done_server_count == null) delete payload.reconfig_done_server_count;
      if (payload.reconfig_done_logical_cores == null) delete payload.reconfig_done_logical_cores;
      if (payload.reconfig_done_cost_cny == null) delete payload.reconfig_done_cost_cny;
      await saveResourcePlanningConfig(payload);
      const resp = await calculateResourcePlanning(payload);
      setResult(ensureApiOk(resp).data);
      message.success('资源规划计算完成');
    } catch (e) {
      message.error(parseApiError(e, '资源规划计算失败'));
    } finally {
      setLoading(false);
    }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card>
        <Title level={4} style={{ marginTop: 0 }}>资源规划（3.1-3.8）</Title>
        <Text type="secondary">按最新需求文档执行：规划配置、改配利旧、准系统采购利旧、新机采购、续保、自维修、处置、结果分析。</Text>
      </Card>

      <Card title="3.1 规划配置 + 3.2/3.3 输入">
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            admit_value_score: 0,
            compute_demand_cores: 100000,
            warm_storage_demand_tb: 0,
            hot_storage_demand_tb: 0,
            cabinet_and_other_cost_cny: 0,
            annual_depreciation_cny: 0,
            disposal_psas: '/server-decommission',
            non_business_psas: '/online-product',
            quasi_purchase_server_count: 0,
            quasi_purchase_logical_cores: 0,
            quasi_purchase_cost_cny: 0
          }}
          onFinish={onSubmit}
        >
          <Row gutter={16}>
            <Col span={6}><Form.Item name="admit_value_score" label="准入套餐价值分"><InputNumber {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="compute_demand_cores" label="计算需求核" rules={[{ required: true }]}><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="warm_storage_demand_tb" label="温存储需求(TB)"><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="hot_storage_demand_tb" label="热存储需求(TB)"><InputNumber min={0} {...numberProps} /></Form.Item></Col>

            <Col span={6}><Form.Item name="cabinet_and_other_cost_cny" label="机柜及其他费用(CNY)"><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="annual_depreciation_cny" label="年度折旧(CNY)"><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="disposal_psas" label="处置PSA(逗号分隔)" rules={[{ required: true }]}><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="non_business_psas" label="非业务PSA(逗号分隔)" rules={[{ required: true }]}><Input /></Form.Item></Col>

            <Col span={6}><Form.Item name="reconfig_done_server_count" label="已改配成功服务器(可选)"><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="reconfig_done_logical_cores" label="已改配成功逻辑核(可选)"><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="reconfig_done_cost_cny" label="已改配费用(CNY,可选)"><InputNumber min={0} {...numberProps} /></Form.Item></Col>

            <Col span={6}><Form.Item name="quasi_purchase_server_count" label="准系统服务器数" rules={[{ required: true }]}><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="quasi_purchase_logical_cores" label="准系统逻辑核" rules={[{ required: true }]}><InputNumber min={0} {...numberProps} /></Form.Item></Col>
            <Col span={6}><Form.Item name="quasi_purchase_cost_cny" label="准系统费用(CNY)" rules={[{ required: true }]}><InputNumber min={0} {...numberProps} /></Form.Item></Col>
          </Row>
          <Space>
            <Button type="primary" htmlType="submit" loading={loading}>生成资源规划</Button>
            <Button icon={<ReloadOutlined />} onClick={() => form.resetFields()} loading={configLoading}>重置</Button>
            <Button onClick={async () => {
              try {
                const values = form.getFieldsValue();
                await saveResourcePlanningConfig(values);
                message.success('规划配置已保存');
              } catch (e) {
                message.error(parseApiError(e, '保存规划配置失败'));
              }
            }}>保存配置</Button>
          </Space>
        </Form>
      </Card>

      {result ? (
        <>
          <Alert
            type="success"
            showIcon
            message="计算完成"
            description={`生成时间：${new Date(result.generated_at).toLocaleString()}，新机套餐：${result.new_purchase_plan.package_config_type}`}
          />

          <Row gutter={16}>
            <Col span={8}><Card title="3.2 改配利旧"><Statistic title="成功服务器" value={result.reconfig_plan.server_count} /><Statistic title="成功逻辑核" value={result.reconfig_plan.logical_cores} /><Statistic title="费用(CNY)" value={result.reconfig_plan.cost_cny} /></Card></Col>
            <Col span={8}><Card title="3.3 准系统采购利旧"><Statistic title="服务器数" value={result.quasi_purchase_plan.server_count} /><Statistic title="逻辑核" value={result.quasi_purchase_plan.logical_cores} /><Statistic title="费用(CNY)" value={result.quasi_purchase_plan.cost_cny} /></Card></Col>
            <Col span={8}><Card title="3.5 续保"><Statistic title="设备数量" value={result.renewal_plan.device_count} /><Statistic title="计算覆盖(核)" value={result.renewal_plan.covered_compute_cores} /><Statistic title="温存储覆盖(TB)" value={result.renewal_plan.covered_warm_storage_tb} /><Statistic title="热存储覆盖(TB)" value={result.renewal_plan.covered_hot_storage_tb} /><Statistic title="GPU覆盖(卡)" value={result.renewal_plan.covered_gpu_cards} /><Statistic title="预算占用(CNY)" value={result.renewal_plan.budget_cny} /></Card></Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Card title="3.4 新机采购">
                <Statistic title="推荐套餐" value={result.new_purchase_plan.package_config_type} />
                <Statistic title="发布年份" value={result.new_purchase_plan.package_release_year} />
                <Statistic title="建议采购数量" value={result.new_purchase_plan.server_count} />
                <Statistic title="覆盖算力(核)" value={result.new_purchase_plan.covered_logical_cores} />
                <Statistic title="采购金额(CNY)" value={result.new_purchase_plan.purchase_amount_cny} />
                <Statistic title="年度预算(CNY)" value={result.new_purchase_plan.annual_budget_cny} />
              </Card>
            </Col>
            <Col span={12}>
              <Card title="3.6 / 3.7">
                <Statistic title="自维修设备数量" value={result.self_repair_plan.device_count} />
                <Statistic title="自维修覆盖算力(核)" value={result.self_repair_plan.covered_cores} />
                <Statistic title="处置设备数量" value={result.disposal_plan.device_count} />
                <Statistic title="处置计算覆盖(核)" value={result.disposal_plan.covered_compute_cores} />
                <Statistic title="处置温存储覆盖(TB)" value={result.disposal_plan.covered_warm_storage_tb} />
                <Statistic title="处置热存储覆盖(TB)" value={result.disposal_plan.covered_hot_storage_tb} />
                <Statistic title="处置GPU覆盖(卡)" value={result.disposal_plan.covered_gpu_cards} />
                <Statistic title="处置套餐未匹配数" value={result.disposal_plan.unmatched_package_count} />
                <Statistic title="PSA命中设备数" value={result.disposal_plan.matched_psa_server_count} />
                <Text type="secondary">命中PSA：{(result.disposal_plan.normalized_psas || []).join(', ') || '-'}</Text>
              </Card>
            </Col>
          </Row>

          <Card title="3.8 结果分析">
            <Row gutter={16}>
              <Col span={8}>
                <Card type="inner" title="费用构成(CNY)">
                  <Statistic title="改配利旧" value={result.result_analysis.amount.reconfig_cost_cny} />
                  <Statistic title="准系统采购利旧" value={result.result_analysis.amount.quasi_purchase_cost_cny} />
                  <Statistic title="新机采购" value={result.result_analysis.amount.new_purchase_cost_cny} />
                  <Statistic title="续保" value={result.result_analysis.amount.renewal_cost_cny} />
                  <Statistic title="机柜及其他" value={result.result_analysis.amount.cabinet_other_cost_cny} />
                  <Statistic title="合计" value={result.result_analysis.amount.total_cost_cny} />
                </Card>
              </Col>
              <Col span={8}>
                <Card type="inner" title="成本构成(CNY)">
                  <Statistic title="改配利旧" value={result.result_analysis.cost.reconfig_cost_cny} />
                  <Statistic title="准系统采购利旧" value={result.result_analysis.cost.quasi_purchase_cost_cny} />
                  <Statistic title="新机采购" value={result.result_analysis.cost.new_purchase_cost_cny} />
                  <Statistic title="续保" value={result.result_analysis.cost.renewal_cost_cny} />
                  <Statistic title="折旧" value={result.result_analysis.cost.depreciation_cost_cny} />
                  <Statistic title="机柜及其他" value={result.result_analysis.cost.cabinet_other_cost_cny} />
                  <Statistic title="合计" value={result.result_analysis.cost.total_cost_cny} />
                </Card>
              </Col>
              <Col span={8}>
                <Card type="inner" title="算力-计算(核)">
                  <Statistic title="改配利旧" value={result.result_analysis.compute_capacity.reconfig_cores} />
                  <Statistic title="准系统采购利旧" value={result.result_analysis.compute_capacity.quasi_purchase_cores} />
                  <Statistic title="新机采购" value={result.result_analysis.compute_capacity.new_purchase_cores} />
                  <Statistic title="存量继续使用" value={result.result_analysis.compute_capacity.stock_continue_cores} />
                  <Statistic title="合计" value={result.result_analysis.compute_capacity.total_cores} />
                </Card>
              </Col>
            </Row>
            <Row gutter={16} style={{ marginTop: 16 }}>
              <Col span={12}>
                <Card type="inner" title="算力-温存储(TB)">
                  <Statistic title="改配利旧" value={result.result_analysis.warm_storage_capacity.reconfig_tb} />
                  <Statistic title="准系统采购利旧" value={result.result_analysis.warm_storage_capacity.quasi_purchase_tb} />
                  <Statistic title="新机采购" value={result.result_analysis.warm_storage_capacity.new_purchase_tb} />
                  <Statistic title="存量继续使用" value={result.result_analysis.warm_storage_capacity.stock_continue_tb} />
                  <Statistic title="合计" value={result.result_analysis.warm_storage_capacity.total_tb} />
                </Card>
              </Col>
              <Col span={12}>
                <Card type="inner" title="算力-热存储(TB)">
                  <Statistic title="改配利旧" value={result.result_analysis.hot_storage_capacity.reconfig_tb} />
                  <Statistic title="准系统采购利旧" value={result.result_analysis.hot_storage_capacity.quasi_purchase_tb} />
                  <Statistic title="新机采购" value={result.result_analysis.hot_storage_capacity.new_purchase_tb} />
                  <Statistic title="存量继续使用" value={result.result_analysis.hot_storage_capacity.stock_continue_tb} />
                  <Statistic title="合计" value={result.result_analysis.hot_storage_capacity.total_tb} />
                </Card>
              </Col>
            </Row>
          </Card>
        </>
      ) : null}
    </Space>
  );
}
