import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import * as authApi from '@/api/auth'
import type { User, LoginCredentials, SignupCredentials } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const initialized = ref(false)

  const isAuthenticated = computed(() => !!user.value)

  let initPromise: Promise<void> | null = null

  function initialize(): Promise<void> {
    if (!initPromise) {
      initPromise = authApi.getCurrentUser()
        .then((userData) => {
          user.value = userData
        })
        .catch(() => {
          user.value = null
        })
        .finally(() => {
          initialized.value = true
        })
    }
    return initPromise
  }

  async function login(credentials: LoginCredentials): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      await authApi.login(credentials)
      user.value = await authApi.getCurrentUser()
      return true
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.data?.error) {
        error.value = err.response.data.error
      } else {
        error.value = 'Login failed'
      }
      return false
    } finally {
      loading.value = false
    }
  }

  async function signup(credentials: SignupCredentials): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      await authApi.signup(credentials)
      user.value = await authApi.getCurrentUser()
      return true
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.data?.error) {
        error.value = err.response.data.error
      } else {
        error.value = 'Signup failed'
      }
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await authApi.logout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      user.value = null
      loading.value = false
      initialized.value = false
      initPromise = null
    }
  }

  return {
    user,
    loading,
    error,
    initialized,
    isAuthenticated,
    initialize,
    login,
    signup,
    logout
  }
})
