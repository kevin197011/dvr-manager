import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { authService } from '../services/authService';
import { getApiErrorMessage } from '../utils/format';

const useAuthStore = create(
  persist(
    (set, get) => ({
      token: null,
      user: null,

      login: async (username, password) => {
        try {
          const response = await authService.login(username, password);
          set({
            token: response.token,
            user: response.user,
          });
          return { success: true };
        } catch (error) {
          return { success: false, message: getApiErrorMessage(error, '登录失败') };
        }
      },

      logout: async (callServer = true) => {
        if (callServer) {
          try {
            await authService.logout();
          } catch {
            // ignore
          }
        }
        set({ token: null, user: null });
      },

      hydrate: ({ token, user }) => {
        set({ token, user });
      },

      checkAuth: async () => {
        const token = get().token;
        if (!token) {
          return false;
        }

        try {
          const user = await authService.verifyToken(token);
          set({ user });
          return true;
        } catch (error) {
          const status = error?.response?.status;
          if (status === 401 || status === 403) {
            set({ token: null, user: null });
            return false;
          }
          // 网络/5xx：保留 token，避免误登出
          return !!get().user;
        }
      },

      isAuthenticated: () => !!get().token,
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ token: state.token, user: state.user }),
    }
  )
);

export { useAuthStore };
