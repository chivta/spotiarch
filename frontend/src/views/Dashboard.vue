<template>
  <div class="dashboard">
    <header class="dashboard-header">
      <div class="header-content">
        <h1 class="logo">Spotiarch</h1>
        <div class="user-section">
          <span class="user-name">{{ authStore.user?.name || 'User' }}</span>
          <button class="logout-button" @click="handleLogout">Logout</button>
        </div>
      </div>
    </header>

    <main class="dashboard-main">
      <div class="container">
        <div class="welcome-section">
          <h2>Welcome back, {{ authStore.user?.name }}!</h2>
          <p>Here's your music overview</p>
        </div>

        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-icon">📊</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.totalItems }}</div>
              <div class="stat-label">Total Items</div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">✓</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.activeItems }}</div>
              <div class="stat-label">Active</div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">✓</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.completedItems }}</div>
              <div class="stat-label">Completed</div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">⏳</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.pendingItems }}</div>
              <div class="stat-label">Pending</div>
            </div>
          </div>
        </div>

        <div class="items-section">
          <h3>Recent Items</h3>
          <div class="items-list">
            <div
              v-for="item in items"
              :key="item.id"
              class="item-card"
            >
              <div class="item-header">
                <h4>{{ item.title }}</h4>
                <span :class="['status-badge', item.status]">
                  {{ item.status }}
                </span>
              </div>
              <p class="item-description">{{ item.description }}</p>
              <div class="item-footer">
                <span class="item-date">{{ formatDate(item.createdAt) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { DashboardStats, DashboardItem } from '@/types'

const router = useRouter()
const authStore = useAuthStore()

const stats = ref<DashboardStats>({
  totalItems: 42,
  activeItems: 15,
  completedItems: 20,
  pendingItems: 7
})

const items = ref<DashboardItem[]>([
  {
    id: '1',
    title: 'Summer Vibes Playlist',
    description: 'A collection of chill summer tracks perfect for relaxing by the beach.',
    status: 'active',
    createdAt: '2026-01-15T10:00:00Z'
  },
  {
    id: '2',
    title: 'Workout Mix 2026',
    description: 'High-energy tracks to keep you motivated during your workout sessions.',
    status: 'completed',
    createdAt: '2026-01-10T14:30:00Z'
  },
  {
    id: '3',
    title: 'Jazz Classics',
    description: 'Timeless jazz standards from the golden era of jazz music.',
    status: 'active',
    createdAt: '2026-01-05T09:15:00Z'
  },
  {
    id: '4',
    title: 'Lo-Fi Study Session',
    description: 'Ambient lo-fi beats perfect for studying or focused work.',
    status: 'pending',
    createdAt: '2026-01-01T16:45:00Z'
  },
  {
    id: '5',
    title: 'Electronic Dreams',
    description: 'Modern electronic music from emerging artists around the world.',
    status: 'active',
    createdAt: '2025-12-28T11:20:00Z'
  },
  {
    id: '6',
    title: 'Indie Rock Collection',
    description: 'The best indie rock tracks from the past decade.',
    status: 'completed',
    createdAt: '2025-12-20T08:00:00Z'
  }
])

async function handleLogout() {
  await authStore.logout()
  router.push({ name: 'landing' })
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}
</script>

<style scoped>
.dashboard {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.dashboard-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-size: 24px;
  font-weight: bold;
}

.user-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-name {
  font-size: 16px;
  font-weight: 500;
}

.logout-button {
  background-color: rgba(255, 255, 255, 0.2);
  color: white;
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.3s ease;
}

.logout-button:hover {
  background-color: rgba(255, 255, 255, 0.3);
  transform: none;
}

.dashboard-main {
  padding: 40px 20px;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
}

.welcome-section {
  margin-bottom: 40px;
}

.welcome-section h2 {
  font-size: 32px;
  color: #2c3e50;
  margin-bottom: 8px;
}

.welcome-section p {
  font-size: 18px;
  color: #666;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.stat-icon {
  font-size: 36px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #2c3e50;
}

.stat-label {
  font-size: 14px;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.items-section h3 {
  font-size: 24px;
  color: #2c3e50;
  margin-bottom: 20px;
}

.items-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.item-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.item-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  margin-bottom: 12px;
}

.item-header h4 {
  font-size: 18px;
  color: #2c3e50;
  flex: 1;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.active {
  background-color: #e3f2fd;
  color: #1976d2;
}

.status-badge.completed {
  background-color: #e8f5e9;
  color: #388e3c;
}

.status-badge.pending {
  background-color: #fff3e0;
  color: #f57c00;
}

.item-description {
  font-size: 14px;
  color: #666;
  line-height: 1.5;
  margin-bottom: 16px;
}

.item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.item-date {
  font-size: 12px;
  color: #999;
}
</style>
