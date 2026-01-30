<template>
  <div class="signup-page">
    <div class="signup-container">
      <div class="signup-card">
        <h1 class="logo">Spotiarch</h1>
        <h2 class="page-title">Create your account</h2>

        <form @submit.prevent="handleSignup" class="signup-form">
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
              placeholder="Create a password"
              required
              minlength="8"
            />
          </div>

          <div class="form-group">
            <label for="confirmPassword">Confirm Password</label>
            <input
              id="confirmPassword"
              v-model="form.confirmPassword"
              type="password"
              placeholder="Repeat your password"
              required
            />
          </div>

          <div v-if="formError || authStore.error" class="error-message">
            {{ formError || authStore.error }}
          </div>

          <button
            type="submit"
            class="submit-button"
            :disabled="authStore.loading"
          >
            {{ authStore.loading ? 'Creating account...' : 'Sign Up' }}
          </button>
        </form>

        <div class="page-footer">
          <p>Already have an account? <router-link to="/login">Login</router-link></p>
          <router-link to="/" class="back-link">← Back to home</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const formError = ref<string | null>(null)

const form = reactive({
  email: '',
  password: '',
  confirmPassword: ''
})

async function handleSignup() {
  formError.value = null

  if (form.password !== form.confirmPassword) {
    formError.value = 'Passwords do not match'
    return
  }

  const success = await authStore.signup({
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
.signup-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
  padding: 20px;
}

.signup-container {
  width: 100%;
  max-width: 440px;
}

.signup-card {
  background: white;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.logo {
  text-align: center;
  color: #764ba2;
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

.signup-form {
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
  border-color: #764ba2;
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
  background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
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
  color: #764ba2;
  font-weight: 500;
}

.back-link {
  color: #999;
  font-size: 13px;
}
</style>
