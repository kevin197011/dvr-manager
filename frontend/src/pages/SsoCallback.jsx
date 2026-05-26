import { useEffect, useState } from 'react';
import { Result, Spin, Button } from 'antd';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';

function SsoCallback() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { hydrate } = useAuthStore();
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    const err = params.get('error');
    if (err) {
      setErrorMsg(err);
      return;
    }
    const token = params.get('token');
    const username = params.get('username');
    const role = params.get('role');
    if (!token || !username) {
      setErrorMsg('SSO 回调缺少 token 或 username');
      return;
    }
    hydrate({ token, user: { username, role: role || 'user' } });
    // 用 replace 跳到首页，并让 React 在下一帧再执行，确保当前组件先卸载
    navigate('/', { replace: true });
  }, [params, hydrate, navigate]);

  if (errorMsg) {
    return (
      <Result
        status="error"
        title="SSO 登录失败"
        subTitle={errorMsg}
        extra={
          <Button type="primary" onClick={() => navigate('/login', { replace: true })}>
            返回登录页
          </Button>
        }
      />
    );
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 16,
      }}
    >
      <Spin size="large" />
      <span style={{ color: '#64748B' }}>SSO 登录处理中...</span>
    </div>
  );
}

export default SsoCallback;
