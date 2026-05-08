<template>
  <div class="step-card">
    <div class="card-title">Verify your email</div>
    <div class="card-desc">Check the onboarding service logs for your verification token.</div>

    <div class="info-row">
      Check <code>cmd/onboarding</code> logs — the token was printed when you registered.
    </div>

    <div class="form-field">
      <label for="token">Verification token</label>
      <input id="token" v-model="token" type="text" placeholder="Paste token here" @keydown.enter="submit" />
    </div>

    <button class="btn-primary" :disabled="!token.trim() || loading" @click="submit">
      {{ loading ? 'Verifying…' : 'Verify email' }}
    </button>
    <div v-if="error" class="form-error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { verifyEmail } from '../api.ts'
import type { User } from '../types.ts'

const props = defineProps<{ user: User }>()
const emit = defineEmits<{ done: [user: User] }>()

const token = ref('')
const loading = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  if (!token.value.trim()) return
  loading.value = true
  error.value = ''
  try {
    const updated = await verifyEmail(props.user.id, token.value.trim())
    emit('done', updated)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
</script>
