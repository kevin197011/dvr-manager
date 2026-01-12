import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { authService } from '../services/authService';

const useAuthStore = create(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      
      login: async (username, password) => {
        try {
          const response = await authService.login(username, password);
          set({
            token: response.token,
            user: response.user,
            isAuthenticated: true,
          });
          return { success: true };
        } catch (error) {
          return { success: false, message: error.message };
        }
      },
      
      logout: () => {
        set({
          token: null,
          user: null,
          isAuthenticated: false,
        });
      },
      
      checkAuth: async () => {
        const token = useAuthStore.getState().token;
        if (!token) {
          set({ isAuthenticated: false });
          return false;
        }
        
        try {
          const user = await authService.verifyToken(token);
          set({ user, isAuthenticated: true });
          return true;
        } catch (error) {
          set({ token: null, user: null, isAuthenticated: false });
          return false;
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ token: state.token, user: state.user }),
    }
  )
);

export { useAuthStore };
