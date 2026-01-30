export interface User {
  id: string
  email: string
  name: string
  createdAt: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface SignupCredentials {
  email: string
  password: string
}

export interface AuthResponse {
  user: User
}

export interface DashboardStats {
  totalItems: number
  activeItems: number
  completedItems: number
  pendingItems: number
}

export interface DashboardItem {
  id: string
  title: string
  description: string
  status: 'active' | 'completed' | 'pending'
  createdAt: string
}
