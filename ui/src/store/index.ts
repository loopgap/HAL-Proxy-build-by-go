import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface User {
  id: string
  name: string
  email: string
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  login: (user: User, token: string) => void
  logout: () => void
  updateUser: (user: Partial<User>) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      login: (user, token) => {
        localStorage.setItem('auth_token', token)
        set({ user, token, isAuthenticated: true })
      },
      logout: () => {
        localStorage.removeItem('auth_token')
        set({ user: null, token: null, isAuthenticated: false })
      },
      updateUser: (updates) => set((state) => ({ user: state.user ? { ...state.user, ...updates } : null })),
    }),
    { name: 'auth-storage', storage: createJSONStorage(() => localStorage), partialize: (state) => ({ user: state.user, token: state.token, isAuthenticated: state.isAuthenticated }) }
  )
)

interface UIState {
  sidebarOpen: boolean
  theme: 'light' | 'dark' | 'system'
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setTheme: (theme: 'light' | 'dark' | 'system') => void
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({ sidebarOpen: true, theme: 'system', toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })), setSidebarOpen: (open) => set({ sidebarOpen: open }), setTheme: (theme) => set({ theme }) }),
    { name: 'ui-storage', storage: createJSONStorage(() => localStorage) }
  )
)

interface Notification {
  id: string
  type: 'success' | 'error' | 'info' | 'warning'
  message: string
}

interface NotificationState {
  notifications: Notification[]
  addNotification: (notification: Omit<Notification, 'id'>) => void
  removeNotification: (id: string) => void
  clearAll: () => void
}

export const useNotificationStore = create<NotificationState>()((set) => ({
  notifications: [],
  addNotification: (notification) => set((state) => ({ notifications: [...state.notifications, { ...notification, id: crypto.randomUUID() }] })),
  removeNotification: (id) => set((state) => ({ notifications: state.notifications.filter((n) => n.id !== id) })),
  clearAll: () => set({ notifications: [] }),
}))

interface FilterState {
  caseFilters: { status: string[]; riskClass: string[]; search: string }
  approvalFilters: { status: string[]; riskClass: string[] }
  dateRange: { start: string | null; end: string | null }
  setCaseFilters: (filters: Partial<FilterState['caseFilters']>) => void
  setApprovalFilters: (filters: Partial<FilterState['approvalFilters']>) => void
  setDateRange: (range: { start: string | null; end: string | null }) => void
  resetFilters: () => void
}

const initialFilters = {
  caseFilters: { status: [], riskClass: [], search: '' },
  approvalFilters: { status: [], riskClass: [] },
  dateRange: { start: null, end: null },
}

export const useFilterStore = create<FilterState>()((set) => ({
  ...initialFilters,
  setCaseFilters: (filters) => set((state) => ({ caseFilters: { ...state.caseFilters, ...filters } })),
  setApprovalFilters: (filters) => set((state) => ({ approvalFilters: { ...state.approvalFilters, ...filters } })),
  setDateRange: (range) => set({ dateRange: range }),
  resetFilters: () => set(initialFilters),
}))

export default { useAuthStore, useUIStore, useNotificationStore, useFilterStore }