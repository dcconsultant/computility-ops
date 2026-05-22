import { useEffect, useState } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { Alert, Button, Divider, Drawer, Form, Input, InputNumber, Layout, List, Menu, Space, Typography, message } from 'antd';
import { ReloadOutlined, SettingOutlined } from '@ant-design/icons';
import { APP_VERSION } from './version';
import { listImportErrors, testMySQLConnection } from './api';
import { ensureApiOk, parseApiError } from './error';
import type { ImportErrorInsight } from './types';

const { Header, Content } = Layout;
const { Title, Text } = Typography;

export default function AppLayout() {
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testOK, setTestOK] = useState<string>('');
  const [loadingErrors, setLoadingErrors] = useState(false);
  const [importErrors, setImportErrors] = useState<ImportErrorInsight[]>([]);
  const [form] = Form.useForm();

  const section = new URLSearchParams(location.search).get('section') || '';
  const key = location.pathname.startsWith('/result') || location.pathname.startsWith('/plan/')
    ? '/plan'
    : location.pathname.startsWith('/failure/')
      ? '/failure'
      : location.pathname.startsWith('/reconfig/')
        ? '/reconfig'
        : location.pathname.startsWith('/meta-models')
          ? '/meta-models'
          : location.pathname.startsWith('/import') && section === 'value-score'
            ? '/value-score'
            : location.pathname.startsWith('/import') && section === 'resource-analysis'
              ? '/resource-analysis'
              : location.pathname.startsWith('/import') && section === 'test-zone'
                ? '/test-zone'
                : location.pathname;
  const isFailureDashboard = location.pathname === '/failure/dashboard';
  const isMetaModels = location.pathname.startsWith('/meta-models');

  useEffect(() => {
    if (open) {
      loadImportErrors();
    }
  }, [open]);

  async function onTestMySQL() {
    try {
      const values = await form.validateFields();
      setTesting(true);
      const res = ensureApiOk(await testMySQLConnection(values));
      setTestOK(`连接成功，延迟 ${res.data.latency_ms}ms`);
      message.success('MySQL 连接成功');
    } catch (e) {
      setTestOK('');
      message.error(parseApiError(e, 'MySQL 连接失败'));
    } finally {
      setTesting(false);
    }
  }

  async function loadImportErrors() {
    setLoadingErrors(true);
    try {
      const resp = ensureApiOk(await listImportErrors(20));
      setImportErrors(resp.data.list || []);
    } catch (e) {
      message.error(parseApiError(e, '加载导入异常失败'));
    } finally {
      setLoadingErrors(false);
    }
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
        <Title level={4} style={{ color: '#fff', margin: 0 }}>
          Computility Ops <Text style={{ color: '#ddd', fontSize: 14 }}>{APP_VERSION}</Text>
        </Title>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={[key]}
          items={[
            { key: '/meta-models', label: <Link to="/meta-models">元数据</Link> },
            { key: '/value-score', label: <Link to="/import?section=value-score&tab=value_score_setup">价值分</Link> },
            { key: '/resource-planning', label: <Link to="/resource-planning">资源规划</Link> },
            {
              key: '/ops-management',
              label: '运维管理',
              children: [
                { key: '/plan', label: <Link to="/plan">续保管理</Link> },
                { key: '/failure', label: <Link to="/failure">故障率</Link> },
                { key: '/reconfig', label: <Link to="/reconfig">改配管理</Link> }
              ]
            },
            {
              key: '/business-terms',
              label: '商务条款',
              children: [
                { key: '/contracts', label: <Link to="/contracts">合同</Link> },
                { key: '/suppliers', label: <Link to="/suppliers">供应商</Link> }
              ]
            },
            { key: '/resource-analysis', label: <Link to="/import?section=resource-analysis&tab=assets">资源分析</Link> },
            { key: '/test-zone', label: <Link to="/import?section=test-zone&tab=servers">测试专区</Link> }
          ]}
          style={{ flex: 1, minWidth: 0 }}
        />

        <Button
          type="text"
          size="small"
          icon={<SettingOutlined style={{ color: 'rgba(255,255,255,0.6)' }} />}
          style={{ opacity: 0.7 }}
          onClick={() => setOpen(true)}
        />
      </Header>
      <Content style={{ padding: isFailureDashboard ? 0 : (isMetaModels ? 12 : 24), maxWidth: (isFailureDashboard || isMetaModels) ? 'none' : 1400, width: '100%', margin: isMetaModels ? 0 : '0 auto' }}>
        <Outlet />
      </Content>

      <Drawer
        title="系统配置"
        placement="right"
        width={460}
        onClose={() => setOpen(false)}
        open={open}
      >
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Text type="secondary">用于测试 MySQL 连接可用性；并提供导入异常自动分析。</Text>
          {testOK ? <Alert type="success" message={testOK} showIcon /> : null}

          <Form
            form={form}
            layout="vertical"
            initialValues={{
              host: '127.0.0.1',
              port: 3306,
              params: 'parseTime=true&loc=Local&charset=utf8mb4'
            }}
          >
            <Form.Item label="Host" name="host" rules={[{ required: true, message: '请输入 host' }]}>
              <Input placeholder="127.0.0.1" />
            </Form.Item>
            <Form.Item label="Port" name="port" rules={[{ required: true, message: '请输入 port' }]}>
              <InputNumber min={1} max={65535} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="User" name="user" rules={[{ required: true, message: '请输入 user' }]}>
              <Input placeholder="root" />
            </Form.Item>
            <Form.Item label="Password" name="password">
              <Input.Password placeholder="******" />
            </Form.Item>
            <Form.Item label="Database" name="database" rules={[{ required: true, message: '请输入 database' }]}>
              <Input placeholder="computility_ops" />
            </Form.Item>
            <Form.Item label="Params" name="params">
              <Input placeholder="parseTime=true&loc=Local&charset=utf8mb4" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" loading={testing} onClick={onTestMySQL}>
                测试 MySQL 连接
              </Button>
            </Form.Item>
          </Form>

          <Divider style={{ margin: '4px 0' }} />

          <Space direction="vertical" size={4} style={{ width: '100%' }}>
            <Text strong>帮助文档</Text>
            <a href="https://github.com/dcconsultant/computility-ops/blob/main/docs/fault_rate_advanced_feature_plan_v2.md" target="_blank" rel="noreferrer">
              服务器故障率特性高级功能方案（V2）
            </a>
            <a href="https://github.com/dcconsultant/computility-ops/blob/main/docs/fault_rate_metric_spec.md" target="_blank" rel="noreferrer">
              故障率口径说明（现有）
            </a>
          </Space>

          <Divider style={{ margin: '4px 0' }} />

          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <Text strong>导入异常分析（最近20条）</Text>
            <Button icon={<ReloadOutlined />} loading={loadingErrors} onClick={loadImportErrors}>刷新</Button>
          </Space>
          <List
            size="small"
            bordered
            loading={loadingErrors}
            locale={{ emptyText: '暂无导入异常记录' }}
            dataSource={importErrors}
            renderItem={(item) => (
              <List.Item>
                <Space direction="vertical" size={2} style={{ width: '100%' }}>
                  <Text style={{ fontSize: 12 }}>{item.time} | {item.action}</Text>
                  <Text style={{ fontSize: 12 }} type="secondary">请求ID: {item.request_id || '-'}</Text>
                  <Text style={{ fontSize: 12 }} type="danger">原因: {item.reason}</Text>
                  <Text style={{ fontSize: 12 }}>建议: {item.hint}</Text>
                </Space>
              </List.Item>
            )}
          />
        </Space>
      </Drawer>
    </Layout>
  );
}
