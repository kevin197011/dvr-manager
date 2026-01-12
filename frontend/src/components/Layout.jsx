import { useState } from 'react';
import { Layout as AntLayout, Menu, Avatar, Dropdown, Space, Switch } from 'antd';
import { Outlet, useNavigate, useLocation, Link } from 'react-router-dom';
import {
  HomeOutlined,
  SettingOutlined,
  LogoutOutlined,
  UserOutlined,
  VideoCameraOutlined,
  BulbOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../store/authStore';
import { useThemeStore } from '../store/themeStore';
import './Layout.css';

const { Header, Sider, Content, Footer } = AntLayout;

function Layout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const { theme, toggleTheme } = useThemeStore();

  // 根据用户角色显示菜单
  const menuItems = [
    {
      key: '/',
      icon: <VideoCameraOutlined />,
      label: '录像查询',
    },
    // 只有管理员可以看到系统管理菜单
    ...(user?.role === 'admin'
      ? [
          {
            key: '/admin',
            icon: <SettingOutlined />,
            label: '系统管理',
          },
        ]
      : []),
  ];

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ];

  const handleMenuClick = ({ key }) => {
    if (key === 'logout') {
      logout();
      navigate('/login');
    } else if (key === 'profile') {
      // TODO: 个人信息页面
    } else {
      navigate(key);
    }
  };

  return (
    <AntLayout className="app-layout">
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme={theme}
      >
        <div className="logo">
          <Link 
            to="/"
            style={{ 
              color: 'inherit', 
              textDecoration: 'none',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '100%',
              height: '100%',
            }}
          >
            {collapsed ? 'DVR' : 'DVR 点播系统'}
          </Link>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <AntLayout>
        <Header className="app-header">
          <div className="header-right">
            <Space size="middle">
              <Space>
                <BulbOutlined />
                <span style={{ fontSize: 14 }}>主题</span>
                <Switch
                  checked={theme === 'dark'}
                  onChange={toggleTheme}
                  checkedChildren="暗"
                  unCheckedChildren="亮"
                />
              </Space>
              <Dropdown
                menu={{
                  items: userMenuItems,
                  onClick: handleMenuClick,
                }}
                placement="bottomRight"
              >
                <Space className="user-info" style={{ cursor: 'pointer' }}>
                  <Avatar icon={<UserOutlined />} />
                  <span>
                    {user?.username || 'User'}
                    {user?.role === 'admin' && (
                      <span style={{ marginLeft: 8, color: 'var(--ant-color-primary)', fontSize: 12 }}>
                        (管理员)
                      </span>
                    )}
                  </span>
                </Space>
              </Dropdown>
            </Space>
          </div>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
        <Footer className="app-footer">
          <div>系统运行部驱动</div>
        </Footer>
      </AntLayout>
    </AntLayout>
  );
}

export default Layout;
