import { Navigate } from 'react-router-dom';
import { Spin } from 'antd';
import { useAuthStore } from '../store/authStore';
import { useEffect, useState } from 'react';

function ProtectedRoute({ children }) {
  const token = useAuthStore((s) => s.token);
  const checkAuth = useAuthStore((s) => s.checkAuth);
  const [loading, setLoading] = useState(true);
  const [authed, setAuthed] = useState(false);

  useEffect(() => {
    const verify = async () => {
      const ok = await checkAuth();
      setAuthed(ok || !!useAuthStore.getState().token);
      setLoading(false);
    };
    verify();
  }, [checkAuth]);

  if (loading) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (!authed && !token) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

export default ProtectedRoute;
