import axios from 'axios';
import { useAuthStore } from '../store/authStore';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout(false);
      if (!window.location.pathname.startsWith('/login')) {
        window.location.href = '/login';
      }
    }
    if (error.response?.data) {
      return Promise.reject({
        ...error,
        response: { ...error.response, data: error.response.data },
      });
    }
    return Promise.reject(error);
  }
);

export const authService = {
  login: async (username, password) => api.post('/auth/login', { username, password }),

  verifyToken: async (token) => {
    const response = await api.get('/auth/me', {
      headers: { Authorization: `Bearer ${token}` },
    });
    return response.user;
  },

  logout: async () => api.post('/auth/logout'),

  changePassword: async (oldPassword, newPassword) =>
    api.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
    }),

  listSSOProviders: async () => api.get('/auth/sso/providers'),

  ssoLoginURL: (provider) => {
    const base = API_BASE_URL.replace(/\/$/, '');
    return `${base}/auth/sso/${provider.type}/${provider.id}/login`;
  },
};

export const dvrService = {
  play: async (recordId, options = {}) =>
    api.post('/play', { record_id: recordId }, { signal: options.signal }),

  batchPlay: async (recordIds, options = {}) =>
    api.post('/play', { record_ids: recordIds }, { signal: options.signal }),

  getConfig: async () => api.get('/config'),
};

export const adminService = {
  getConfig: async () => api.get('/admin/config'),
  updateConfig: async (config) => api.post('/admin/config', config),
  getDVRServers: async () => api.get('/admin/dvr-servers'),
  updateDVRServers: async (servers) => api.post('/admin/dvr-servers', { servers }),
  reloadConfig: async () => api.post('/admin/reload'),
  getAuditLogs: async (params = {}) => api.get('/admin/audit', { params }),
  getDashboardStats: async (params = {}) => api.get('/admin/dashboard/stats', { params }),
  listUsers: async () => api.get('/admin/users'),
  createUser: async ({ username, password, role }) =>
    api.post('/admin/users', { username, password, role }),
  updateUserRole: async (id, role) => api.put(`/admin/users/${id}/role`, { role }),
  resetUserPassword: async (id, newPassword) =>
    api.post(`/admin/users/${id}/reset-password`, { new_password: newPassword }),
  deleteUser: async (id) => api.delete(`/admin/users/${id}`),
  listSSOProvidersAdmin: async () => api.get('/admin/sso/providers'),
  createSSOProvider: async (payload) => api.post('/admin/sso/providers', payload),
  updateSSOProvider: async (id, payload) => api.put(`/admin/sso/providers/${id}`, payload),
  toggleSSOProvider: async (id) => api.post(`/admin/sso/providers/${id}/toggle`),
  deleteSSOProvider: async (id) => api.delete(`/admin/sso/providers/${id}`),
};

export default api;
