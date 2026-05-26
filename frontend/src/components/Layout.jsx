import { useState } from 'react';
import {
  Layout as AntLayout,
  Menu,
  Avatar,
  Dropdown,
  Space,
  Switch,
  Modal,
  Form,
  Input,
  message,
} from 'antd';
import { Outlet, useNavigate, useLocation, Link } from 'react-router-dom';
import {
  SettingOutlined,
  LogoutOutlined,
  UserOutlined,
  VideoCameraOutlined,
  BulbOutlined,
  AuditOutlined,
  TeamOutlined,
  KeyOutlined,
  CloudOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../store/authStore';
import { useThemeStore } from '../store/themeStore';
import { authService } from '../services/authService';
import './Layout.css';

const { Header, Sider, Content, Footer } = AntLayout;

function Layout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const { theme, toggleTheme } = useThemeStore();
  const [pwdOpen, setPwdOpen] = useState(false);
  const [pwdLoading, setPwdLoading] = useState(false);
  const [pwdForm] = Form.useForm();

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
          {
            key: '/admin/users',
            icon: <TeamOutlined />,
            label: '用户管理',
          },
          {
            key: '/admin/sso',
            icon: <CloudOutlined />,
            label: 'SSO 配置',
          },
          {
            key: '/admin/audit',
            icon: <AuditOutlined />,
            label: '审计查询',
          },
        ]
      : []),
  ];

  const userMenuItems = [
    {
      key: 'change-password',
      icon: <KeyOutlined />,
      label: '修改密码',
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
    } else if (key === 'change-password') {
      pwdForm.resetFields();
      setPwdOpen(true);
    } else {
      navigate(key);
    }
  };

  const onChangePassword = async () => {
    try {
      const values = await pwdForm.validateFields();
      if (values.new_password !== values.confirm_password) {
        message.error('两次输入的新密码不一致');
        return;
      }
      setPwdLoading(true);
      const res = await authService.changePassword(values.old_password, values.new_password);
      if (res?.success) {
        message.success('密码修改成功，请重新登录');
        setPwdOpen(false);
        logout();
        navigate('/login');
      } else {
        message.error(res?.message || '修改失败');
      }
    } catch (err) {
      if (err?.errorFields) return;
      message.error(err?.response?.data?.message || '修改失败');
    } finally {
      setPwdLoading(false);
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
                trigger={['click']}
              >
                <a
                  onClick={(e) => e.preventDefault()}
                  className="user-info"
                  style={{ display: 'inline-flex', alignItems: 'center', cursor: 'pointer', color: 'inherit' }}
                >
                  <Avatar icon={<UserOutlined />} />
                  <span style={{ marginLeft: 8 }}>
                    {user?.username || 'User'}
                    {user?.role === 'admin' && (
                      <span style={{ marginLeft: 8, color: 'var(--ant-color-primary)', fontSize: 12 }}>
                        (管理员)
                      </span>
                    )}
                  </span>
                </a>
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

      <Modal
        title="修改密码"
        open={pwdOpen}
        onOk={onChangePassword}
        onCancel={() => setPwdOpen(false)}
        confirmLoading={pwdLoading}
        okText="确定"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={pwdForm} layout="vertical" autoComplete="off">
          <Form.Item
            name="old_password"
            label="原密码"
            rules={[{ required: true, message: '请输入原密码' }]}
          >
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码长度至少 6 位' },
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            label="确认新密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请再次输入新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的新密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
        </Form>
      </Modal>
    </AntLayout>
  );
}

export default Layout;
