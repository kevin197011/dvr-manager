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
      <div className="login-background" aria-hidden="true">
        {/* 点播底图：渐变 + 网格 */}
        <div className="login-video-grid"></div>

        {/* 胶片帧条：模拟视频帧流动（双份实现无缝循环） */}
        <div className="login-film-strip">
          {[...Array(16)].map((_, i) => (
            <div key={i} className="login-film-frame" />
          ))}
          {[...Array(16)].map((_, i) => (
            <div key={`dup-${i}`} className="login-film-frame" />
          ))}
        </div>

        {/* 波形：模拟音视频电平 */}
        <div className="login-video-waveform">
          {[...Array(7)].map((_, i) => (
            <div key={i} className="login-waveform-bar" style={{ animationDelay: `${i * 0.1}s` }} />
          ))}
        </div>

        {/* 流式数据点（点播数据包） */}
        <div className="login-video-packets">
          {[...Array(12)].map((_, i) => (
            <div
              key={i}
              className="login-packet"
              style={{
                left: `${5 + (i % 5) * 22}%`,
                animationDelay: `${i * 0.5}s`,
                animationDuration: `${5 + (i % 4)}s`,
              }}
            />
          ))}
        </div>

        {/* 播放进度条 */}
        <div className="login-video-progress">
          <div className="login-progress-bar" />
        </div>

        {/* 节点：DVR/流媒体节点 */}
        <div className="login-video-nodes">
          {[0, 1, 2, 3, 4, 5].map((i) => (
            <div
              key={i}
              className="login-node"
              style={{
                left: `${12 + (i % 3) * 38}%`,
                top: `${18 + Math.floor(i / 3) * 55}%`,
                animationDelay: `${i * 0.25}s`,
              }}
            >
              <div className="login-node-pulse" />
            </div>
          ))}
        </div>
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
