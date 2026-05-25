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
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <Spin tip="SSO 登录处理中..." size="large" />
    </div>
  );
}

export default SsoCallback;
