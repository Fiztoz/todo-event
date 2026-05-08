<template>
  <div class="step-card">
    <div class="card-title">Create your account</div>
    <div class="card-desc">Start by telling us who you are.</div>

    <div class="form-field">
      <label for="name">Name</label>
      <input id="name" v-model="name" type="text" placeholder="Your full name" @keydown.enter="submit" />
    </div>
    <div class="form-field">
      <label for="email">Email</label>
      <input id="email" v-model="email" type="email" placeholder="you@example.com" @keydown.enter="submit" />
    </div>

    <button class="btn-primary" :disabled="!name.trim() || !email.trim() || loading" @click="submit">
      {{ loading ? 'Registering…' : 'Register' }}
    </button>
    <div v-if="error" class="form-error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { register } from '../api.ts'
import type { User } from '../types.ts'

const emit = defineEmits<{ done: [user: User] }>()

const name = ref('')
const email = ref('')
const loading = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  if (!name.value.trim() || !email.value.trim()) return
  loading.value = true
  error.value = ''
  try {
    const user = await register(name.value.trim(), email.value.trim())
    emit('done', user)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
</script>
