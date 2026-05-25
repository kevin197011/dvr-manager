import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器：添加 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth-storage');
    if (token) {
      try {
        const authData = JSON.parse(token);
        if (authData.state?.token) {
          config.headers.Authorization = `Bearer ${authData.state.token}`;
        }
      } catch (e) {
        // 忽略解析错误
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器：处理错误
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      // 未授权，清除 token
      localStorage.removeItem('auth-storage');
      window.location.href = '/login';
    }
    // 保留错误响应数据，以便前端可以访问错误信息
    if (error.response?.data) {
      return Promise.reject({
        ...error,
        response: {
          ...error.response,
          data: error.response.data,
        },
      });
    }
    return Promise.reject(error);
  }
);

export const authService = {
  login: async (username, password) => {
    const response = await api.post('/auth/login', { username, password });
    return response;
  },
  
  verifyToken: async (token) => {
    const response = await api.get('/auth/me', {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.user;
  },
  
  logout: async () => {
    await api.post('/auth/logout');
  },

  changePassword: async (oldPassword, newPassword) => {
    return await api.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
    });
  },
};

export const dvrService = {
  play: async (recordId) => {
    return await api.post('/play', { record_id: recordId });
  },
  
  batchPlay: async (recordIds) => {
    return await api.post('/play', { record_ids: recordIds });
  },
  
  getConfig: async () => {
    return await api.get('/config');
  },
};

export const adminService = {
  getConfig: async () => {
    return await api.get('/admin/config');
  },
  
  updateConfig: async (config) => {
    return await api.post('/admin/config', config);
  },
  
  getDVRServers: async () => {
    return await api.get('/admin/dvr-servers');
  },
  
  updateDVRServers: async (servers) => {
    return await api.post('/admin/dvr-servers', { servers });
  },
  
  reloadConfig: async () => {
    return await api.post('/admin/reload');
  },

  /**
   * 获取审计日志（管理员）
   * @param {Object} params - from, to (RFC3339), action, username, page, page_size
   */
  getAuditLogs: async (params = {}) => {
    return await api.get('/admin/audit', { params });
  },

  /** 列出所有用户 */
  listUsers: async () => {
    return await api.get('/admin/users');
  },

  /** 创建用户 */
  createUser: async ({ username, password, role }) => {
    return await api.post('/admin/users', { username, password, role });
  },

  /** 修改用户角色 */
  updateUserRole: async (id, role) => {
    return await api.put(`/admin/users/${id}/role`, { role });
  },

  /** 管理员重置用户密码 */
  resetUserPassword: async (id, newPassword) => {
    return await api.post(`/admin/users/${id}/reset-password`, {
      new_password: newPassword,
    });
  },

  /** 删除用户 */
  deleteUser: async (id) => {
    return await api.delete(`/admin/users/${id}`);
  },
};

export default api;
