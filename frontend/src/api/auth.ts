import apiClient from './client'
import type { User, LoginCredentials, SignupCredentials, AuthResponse } from '@/types'

export function login(credentials: LoginCredentials): Promise<AuthResponse> {
  return apiClient.post('/auth/login', credentials).then(res => res.data)
}

export function signup(credentials: SignupCredentials): Promise<AuthResponse> {
  return apiClient.post('/auth/signup', credentials).then(res => res.data)
}

export function logout(): Promise<void> {
  return apiClient.post('/auth/logout').then(res => res.data)
}

export function getCurrentUser(): Promise<User> {
  return apiClient.get('/user/me').then(res => res.data)
}
