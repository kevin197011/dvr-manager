import { useState } from 'react';
import { Form, Input, Button, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import './Login.css';

function Login() {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuthStore();

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const result = await login(values.username, values.password);
      if (result.success) {
        message.success('登录成功');
        navigate('/');
      } else {
        message.error(result.message || '登录失败');
      }
    } catch (error) {
      message.error('登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-background">
        {/* 视频流网格背景 */}
        <div className="login-video-grid"></div>
        
        {/* 视频波形效果（模拟视频信号波形） */}
        <div className="login-video-waveform">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="login-waveform-bar" style={{
              left: `${10 + i * 18}%`,
              animationDelay: `${i * 0.15}s`,
            }}></div>
          ))}
        </div>
        
        {/* 视频流数据包效果 */}
        <div className="login-video-packets">
          {[...Array(8)].map((_, i) => (
            <div key={i} className="login-packet" style={{
              left: `${Math.random() * 100}%`,
              top: `${Math.random() * 100}%`,
              animationDelay: `${Math.random() * 3}s`,
              animationDuration: `${4 + Math.random() * 3}s`,
            }}></div>
          ))}
        </div>
        
        {/* 视频播放进度指示 */}
        <div className="login-video-progress">
          <div className="login-progress-bar"></div>
        </div>
        
        {/* 视频信号连接点 */}
        <div className="login-video-nodes">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="login-node" style={{
              left: `${15 + (i % 3) * 35}%`,
              top: `${20 + Math.floor(i / 3) * 50}%`,
              animationDelay: `${i * 0.3}s`,
            }}>
              <div className="login-node-pulse"></div>
            </div>
          ))}
        </div>
        
        {/* 视频流扫描线 */}
        <div className="login-video-scanline"></div>
      </div>
      <div className="login-wrapper">
        <div className="login-header">
          <h1 className="login-title">DVR 点播系统</h1>
        </div>
        
        <div className="login-form-wrapper">
          <Form
            name="login"
            onFinish={onFinish}
            autoComplete="off"
            size="large"
            layout="vertical"
            className="login-form"
          >
            <Form.Item
              name="username"
              label="用户名"
              rules={[{ required: true, message: '请输入用户名' }]}
              className="login-form-item"
            >
              <Input
                prefix={<UserOutlined className="login-input-icon" />}
                placeholder="请输入用户名"
                className="login-input"
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请输入密码' }]}
              className="login-form-item"
            >
              <Input.Password
                prefix={<LockOutlined className="login-input-icon" />}
                placeholder="请输入密码"
                className="login-input"
              />
            </Form.Item>

            <Form.Item className="login-form-item-submit">
              <Button 
                type="primary" 
                htmlType="submit" 
                block 
                loading={loading}
                className="login-submit-button"
              >
                {loading ? '登录中...' : '登录'}
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>
      <div className="login-footer">
        系统运维部驱动
      </div>
    </div>
  );
}

export default Login;
