import { useState } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { Alert, Button, Drawer, Form, Input, InputNumber, Layout, Menu, Space, Typography, message } from 'antd';
import { SettingOutlined } from '@ant-design/icons';
import { APP_VERSION } from './version';
import { testMySQLConnection } from './api';
import { ensureApiOk, parseApiError } from './error';

const { Header, Content } = Layout;
const { Title, Text } = Typography;

export default function AppLayout() {
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testOK, setTestOK] = useState<string>('');
  const [form] = Form.useForm();

  const key = location.pathname.startsWith('/result') || location.pathname.startsWith('/plan/')
    ? '/plan'
    : location.pathname.startsWith('/failure/')
      ? '/failure'
      : location.pathname;
  const isFailureDashboard = location.pathname === '/failure/dashboard';

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
        width={420}
        onClose={() => setOpen(false)}
        open={open}
      >
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Text type="secondary">用于测试 MySQL 连接可用性，不会保存密码到后端。</Text>
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
        </Space>
      </Drawer>
    </Layout>
  );
}
