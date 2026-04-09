import { Link, Outlet, useLocation } from 'react-router-dom';
import { Layout, Menu, Typography } from 'antd';

const { Header, Content } = Layout;
const { Title } = Typography;

export default function AppLayout() {
  const location = useLocation();
  const key = location.pathname.startsWith('/result') ? '/result' : location.pathname;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
        <Title level={4} style={{ color: '#fff', margin: 0 }}>
          Computility Ops
        </Title>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={[key]}
          items={[
            { key: '/import', label: <Link to="/import">导入清单</Link> },
            { key: '/plan', label: <Link to="/plan">生成方案</Link> },
            { key: '/result', label: <Link to="/result">结果查询</Link> }
          ]}
          style={{ flex: 1, minWidth: 0 }}
        />
      </Header>
      <Content style={{ padding: 24, maxWidth: 1400, width: '100%', margin: '0 auto' }}>
        <Outlet />
      </Content>
    </Layout>
  );
}
