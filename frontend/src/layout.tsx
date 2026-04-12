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

  const key = location.pathname.startsWith('/result') || location.pathname.startsWith('/plan/')
    ? '/plan'
    : location.pathname.startsWith('/failure/')
      ? '/failure'
      : location.pathname;
  const isFailureDashboard = location.pathname === '/failure/dashboard';

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
            { key: '/import', label: <Link to="/import">配置管理</Link> },
            { key: '/plan', label: <Link to="/plan">续保管理</Link> },
            { key: '/failure', label: <Link to="/failure">故障率分析</Link> }
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
      <Content style={{ padding: isFailureDashboard ? 0 : 24, maxWidth: isFailureDashboard ? 'none' : 1400, width: '100%', margin: '0 auto' }}>
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
