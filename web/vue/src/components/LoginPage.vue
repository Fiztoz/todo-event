<template>
  <div class="login-wrap">
    <div class="login-card">
      <div class="login-header">
        <span class="login-logo">Todoe</span>
        <p>Sign in to manage your tasks</p>
      </div>

      <form class="login-form" @submit.prevent="submit">
        <div class="field">
          <label for="email">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="you@example.com"
            autocomplete="email"
            required
          />
        </div>

        <div class="field">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="••••••••"
            autocomplete="current-password"
            required
          />
        </div>

        <div v-if="error" class="form-error">{{ error }}</div>

        <button class="btn-primary" type="submit" :disabled="loading">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { login } from '../api.ts'

const emit = defineEmits<{ (e: 'logged-in'): void }>()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    await login(email.value, password.value)
    emit('logged-in')
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  padding: 24px;
}

.login-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  padding: 40px;
  width: 100%;
  max-width: 400px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-logo {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--blue);
  letter-spacing: -0.4px;
  display: block;
  margin-bottom: 8px;
}

.login-header p {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text);
}

.field input {
  height: 44px;
  padding: 0 14px;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.95rem;
  font-family: inherit;
  color: var(--text);
  background: var(--bg);
  outline: none;
  transition: border-color 0.15s;
}

.field input::placeholder { color: var(--text-muted); }
.field input:focus { border-color: var(--blue); background: #fff; }

.btn-primary {
  width: 100%;
  height: 44px;
  margin-top: 4px;
}
</style>
