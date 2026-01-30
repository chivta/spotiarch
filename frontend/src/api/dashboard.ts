import apiClient from './client'
import type { DashboardStats, DashboardItem } from '@/types'

export function getDashboardStats(): Promise<DashboardStats> {
  return apiClient.get('/dashboard/stats').then(res => res.data)
}

export function getDashboardItems(): Promise<DashboardItem[]> {
  return apiClient.get('/dashboard/items').then(res => res.data)
}
