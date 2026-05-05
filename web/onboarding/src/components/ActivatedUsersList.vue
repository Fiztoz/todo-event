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
        <template v-if="editingId !== u.id">
          <div class="user-name">{{ u.name }}</div>
          <div class="user-email">{{ u.email }}</div>
          <div class="user-meta">
            <span>credit {{ u.credit_score }}</span>
            <span>·</span>
            <span>{{ formatDate(u.created_at) }}</span>
          </div>
          <div v-if="u.bio" class="user-bio">{{ u.bio }}</div>
          <button class="btn-link" @click="startEdit(u)">Edit</button>
        </template>

        <template v-else>
          <div class="user-edit">
            <input v-model="draft!.name" placeholder="Name" />
            <input v-model="draft!.email" placeholder="Email" type="email" />
            <div v-if="editError" class="form-error">{{ editError }}</div>
            <div class="user-actions">
              <button class="btn-primary" :disabled="saving" @click="save(u)">
                {{ saving ? 'Saving…' : 'Save' }}
              </button>
              <button class="btn-secondary" :disabled="saving" @click="cancelEdit">Cancel</button>
            </div>
          </div>
        </template>
      </li>
    </ul>

    <button class="btn-secondary" @click="load" :disabled="loading">Refresh</button>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listActivated, updateContact } from '../api.ts'
import type { User } from '../types.ts'

const users = ref<User[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const editingId = ref<string | null>(null)
const draft = ref<{ name: string; email: string } | null>(null)
const editError = ref<string | null>(null)
const saving = ref(false)

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

function startEdit(u: User) {
  editingId.value = u.id
  draft.value = { name: u.name, email: u.email }
  editError.value = null
}

function cancelEdit() {
  editingId.value = null
  draft.value = null
  editError.value = null
}

async function save(u: User) {
  if (!draft.value) return
  saving.value = true
  editError.value = null
  try {
    const updated = await updateContact(u.id, draft.value.name, draft.value.email)
    const i = users.value.findIndex(x => x.id === u.id)
    if (i >= 0) users.value[i] = updated
    cancelEdit()
  } catch (e: any) {
    editError.value = e?.message ?? 'Failed to save'
  } finally {
    saving.value = false
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

onMounted(load)
</script>
