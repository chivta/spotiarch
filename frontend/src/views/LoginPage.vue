<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-card">
        <h1 class="logo">Spotiarch</h1>
        <h2 class="page-title">Login to your account</h2>

        <form @submit.prevent="handleLogin" class="login-form">
          <div class="form-group">
            <label for="email">Email</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              placeholder="your@email.com"
              required
            />
          </div>

          <div class="form-group">
            <label for="password">Password</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              placeholder="Enter your password"
              required
            />
          </div>

          <div v-if="authStore.error" class="error-message">
            {{ authStore.error }}
          </div>

          <button
            type="submit"
            class="submit-button"
            :disabled="authStore.loading"
          >
            {{ authStore.loading ? 'Logging in...' : 'Login' }}
          </button>
        </form>

        <div class="page-footer">
          <p>Don't have an account? <router-link to="/signup">Sign up</router-link></p>
          <router-link to="/" class="back-link">← Back to home</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const form = reactive({
  email: '',
  password: ''
})

async function handleLogin() {
  const success = await authStore.login({
    email: form.email,
    password: form.password
  })

  if (success) {
    const redirect = route.query.redirect as string | undefined
    router.push(redirect || { name: 'dashboard' })
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-container {
  width: 100%;
  max-width: 440px;
}

.login-card {
  background: white;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.logo {
  text-align: center;
  color: #667eea;
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 8px;
}

.page-title {
  text-align: center;
  color: #666;
  font-size: 16px;
  font-weight: 400;
  margin-bottom: 30px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.form-group input {
  padding: 12px 16px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  transition: border-color 0.3s ease;
}

.form-group input:focus {
  border-color: #667eea;
}

.error-message {
  padding: 12px;
  background-color: #fee;
  color: #c33;
  border-radius: 6px;
  font-size: 14px;
}

.submit-button {
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 16px;
  font-weight: 600;
  border-radius: 6px;
  margin-top: 10px;
}

.submit-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.submit-button:disabled:hover {
  transform: none;
  box-shadow: none;
}

.page-footer {
  text-align: center;
  margin-top: 24px;
}

.page-footer p {
  color: #666;
  font-size: 14px;
  margin-bottom: 12px;
}

.page-footer p a {
  color: #667eea;
  font-weight: 500;
}

.back-link {
  color: #999;
  font-size: 13px;
}
</style>
