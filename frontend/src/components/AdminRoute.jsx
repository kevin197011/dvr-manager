import { Navigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';

function AdminRoute({ children }) {
  const { user } = useAuthStore();

  // 只有管理员可以访问
  if (user?.role !== 'admin') {
    return <Navigate to="/" replace />;
  }

  return children;
}

export default AdminRoute;
