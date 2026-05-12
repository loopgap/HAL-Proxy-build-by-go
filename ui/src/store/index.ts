import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface User {
  id: string
  name: string
  email: string
  role: string
}

export type AuthMode = 'local_trusted' | 'bearer' | 'api_key'

interface AuthState {
  user: User | null
  token: string | null
  mode: AuthMode | null
  isAuthenticated: boolean
  login: (user: User, token: string, mode?: Extract<AuthMode, 'bearer' | 'api_key'>) => void
  useLocalTrusted: () => void
  logout: () => void
  updateUser: (user: Partial<User>) => void
}

export const readStoredAuth = (): { user: User | null; token: string | null; mode: AuthMode | null } => {
  try {
    const stored = sessionStorage.getItem('auth_token')
    if (stored) {
      const parsed = JSON.parse(stored)
      const mode = (parsed.mode ?? (parsed.token ? 'bearer' : null)) as AuthMode | null
      return { user: parsed.user ?? null, token: parsed.token ?? null, mode }
    }
  } catch {
  }
  return { user: null, token: null, mode: null }
}

const localTrustedUser: User = {
  id: 'local-agent',
  name: 'Local Trusted',
  email: 'local@bridgeos.local',
  role: 'service',
}

const initialAuth = readStoredAuth()

export const useAuthStore = create<AuthState>()(
  (set) => ({
    user: initialAuth.user,
    token: initialAuth.token,
    mode: initialAuth.mode,
    isAuthenticated: initialAuth.mode === 'local_trusted' || initialAuth.token !== null,
    login: (user, token, mode = 'bearer') => {
      const data = JSON.stringify({ user, token, mode })
      try {
        sessionStorage.setItem('auth_token', data)
      } catch {
        console.warn('Failed to persist auth token - session storage may be unavailable')
      }
      set({ user, token, mode, isAuthenticated: token !== null })
    },
    useLocalTrusted: () => {
      try {
        sessionStorage.setItem('auth_token', JSON.stringify({ user: localTrustedUser, token: null, mode: 'local_trusted' }))
      } catch {
        console.warn('Failed to persist auth mode - session storage may be unavailable')
      }
      set({ user: localTrustedUser, token: null, mode: 'local_trusted', isAuthenticated: true })
    },
    logout: () => {
      try {
        sessionStorage.removeItem('auth_token')
      } catch {
      }
      set({ user: null, token: null, mode: null, isAuthenticated: false })
    },
    updateUser: (updates) => set((state) => ({ user: state.user ? { ...state.user, ...updates } : null })),
  })
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
