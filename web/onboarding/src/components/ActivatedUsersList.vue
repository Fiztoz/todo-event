<template>
  <div class="step-card">
    <div class="card-title">Activated users</div>
    <div class="card-desc">Everyone who has completed onboarding.</div>

    <div v-if="loading" class="credit-waiting">
      <div class="spinner"></div>
      <div style="font-size: 0.85rem; color: var(--text-muted);">Loading…</div>
    </div>

    <div v-else-if="error" class="form-error">{{ error }}</div>

    <div v-else-if="users.length === 0" class="empty">
      No activated users yet. Walk one through the onboarding flow on the other tab.
    </div>

    <ul v-else class="user-list">
      <li v-for="u in users" :key="u.id" class="user-row">
        <div class="user-name">{{ u.name }}</div>
        <div class="user-email">{{ u.email }}</div>
        <div class="user-meta">
          <span>credit {{ u.credit_score }}</span>
          <span>·</span>
          <span>{{ formatDate(u.created_at) }}</span>
        </div>
        <div v-if="u.bio" class="user-bio">{{ u.bio }}</div>
      </li>
    </ul>

    <button class="btn-secondary" @click="load" :disabled="loading">Refresh</button>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listActivated } from '../api.ts'
import type { User } from '../types.ts'

const users = ref<User[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

async function load() {
  loading.value = true
  error.value = null
  try {
    users.value = await listActivated()
  } catch (e: any) {
    error.value = e?.message ?? 'Failed to load'
  } finally {
    loading.value = false
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

onMounted(load)
</script>
