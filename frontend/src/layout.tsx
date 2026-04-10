import { Link, Outlet, useLocation } from 'react-router-dom';
import { Layout, Menu, Typography } from 'antd';
import { APP_VERSION } from './version';

const { Header, Content } = Layout;
const { Title, Text } = Typography;

export default function AppLayout() {
  const location = useLocation();
  const key = location.pathname.startsWith('/result') || location.pathname.startsWith('/plan/')
    ? '/plan'
    : location.pathname.startsWith('/failure/')
      ? '/failure'
      : location.pathname;
  const isFailureDashboard = location.pathname === '/failure/dashboard';

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
      </Header>
      <Content style={{ padding: isFailureDashboard ? 0 : 24, maxWidth: isFailureDashboard ? 'none' : 1400, width: '100%', margin: '0 auto' }}>
        <Outlet />
      </Content>
    </Layout>
  );
}
