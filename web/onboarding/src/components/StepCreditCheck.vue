<template>
  <div class="step-card">
    <div class="card-title">Credit check</div>
    <div class="card-desc">Sit tight — we're scoring your credit application automatically.</div>

    <div class="credit-waiting">
      <div class="spinner"></div>
      <div style="color: var(--text-muted); font-size: 0.875rem;">Checking your credit…</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { getUser } from '../api.ts'
import type { User } from '../types.ts'

const props = defineProps<{ user: User }>()
const emit = defineEmits<{ done: [user: User] }>()

let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  timer = setInterval(async () => {
    try {
      const updated = await getUser(props.user.id)
      if (updated.status !== 'email_verified') {
        clearInterval(timer!)
        emit('done', updated)
      }
    } catch {
      // keep polling on transient errors
    }
  }, 3000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
