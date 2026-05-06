<template>
  <nav class="nav">
    <span class="nav-logo">Tasks</span>
    <button
      class="theme-toggle"
      @click="cycleTheme"
      :aria-label="`Current theme: ${theme}. Click to change.`"
    >
      {{ themeLabel }}
    </button>
  </nav>

  <main class="page">
    <div class="card">
      <div class="card-title">Create task</div>
      <form class="create-form" @submit.prevent="onCreate">
        <input
          v-model="newTitle"
          placeholder="What needs to be done?"
          :disabled="creating"
        />
        <button type="submit" :disabled="creating || !newTitle.trim()">
          {{ creating ? 'Adding…' : 'Add' }}
        </button>
      </form>
      <div v-if="createError" class="error">{{ createError }}</div>
    </div>

    <div class="card">
      <div class="card-title">All tasks</div>

      <div v-if="loading" class="spinner"></div>
      <div v-else-if="loadError" class="error">{{ loadError }}</div>
      <div v-else-if="tasks.length === 0" class="empty">
        No tasks yet — add one above.
      </div>

      <ul v-else class="task-list">
        <li
          v-for="t in tasks"
          :key="t.id"
          class="task"
          :class="{ done: t.status === 'done' }"
        >
          <div class="task-row" @click="toggleExpand(t.id)">
            <div class="task-title">{{ t.title }}</div>
            <span class="status-pill" :class="t.status">{{ formatStatus(t.status) }}</span>
          </div>

          <div v-if="expandedId === t.id" class="task-detail">
            <dl>
              <dt>id</dt>     <dd>{{ t.id }}</dd>
              <dt>created</dt><dd>{{ formatDate(t.created_at) }}</dd>
            </dl>

            <div class="status-actions">
              <button
                v-for="s in statuses"
                :key="s"
                class="status-btn"
                :class="{ active: t.status === s }"
                :disabled="updatingId === t.id || t.status === s"
                @click="onChangeStatus(t, s)"
              >{{ formatStatus(s) }}</button>
            </div>
            <div v-if="updateError && updatingId === t.id" class="error">{{ updateError }}</div>
          </div>
        </li>
      </ul>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { listTasks, createTask, changeStatus } from './api.ts'
import type { Task, TaskStatus } from './types.ts'

const tasks = ref<Task[]>([])
const loading = ref(false)
const loadError = ref<string | null>(null)

const newTitle = ref('')
const creating = ref(false)
const createError = ref<string | null>(null)

const expandedId = ref<string | null>(null)
const updatingId = ref<string | null>(null)
const updateError = ref<string | null>(null)

type Theme = 'light' | 'dark' | 'high-contrast'
const theme = ref<Theme>('dark')

const statuses: TaskStatus[] = ['pending', 'in_progress', 'done']

const themeLabels: Record<Theme, string> = {
  light: 'Light',
  dark: 'Dark',
  'high-contrast': 'HC'
}

const themeLabel = computed(() => themeLabels[theme.value])

async function load() {
  loading.value = true
  loadError.value = null
  try {
    tasks.value = await listTasks()
  } catch (e: any) {
    loadError.value = e?.message ?? 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function onCreate() {
  const title = newTitle.value.trim()
  if (!title) return
  creating.value = true
  createError.value = null
  try {
    const created = await createTask(title)
    tasks.value.unshift(created)
    newTitle.value = ''
  } catch (e: any) {
    createError.value = e?.message ?? 'Failed to create'
  } finally {
    creating.value = false
  }
}

function toggleExpand(id: string) {
  expandedId.value = expandedId.value === id ? null : id
  updateError.value = null
}

async function onChangeStatus(task: Task, status: TaskStatus) {
  if (task.status === status) return
  updatingId.value = task.id
  updateError.value = null
  try {
    const updated = await changeStatus(task.id, status)
    const i = tasks.value.findIndex(x => x.id === task.id)
    if (i >= 0) tasks.value[i] = updated
  } catch (e: any) {
    updateError.value = e?.message ?? 'Failed to update status'
  } finally {
    updatingId.value = null
  }
}

function formatStatus(s: TaskStatus): string {
  return s.replace('_', ' ')
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

function cycleTheme() {
  const themes: Theme[] = ['light', 'dark', 'high-contrast']
  const currentIndex = themes.indexOf(theme.value)
  const nextTheme = themes[(currentIndex + 1) % themes.length]
  theme.value = nextTheme
  applyTheme(nextTheme)
}

function applyTheme(t: Theme) {
  document.body.classList.remove('light', 'high-contrast')
  localStorage.setItem('theme', t)
  if (t === 'light') {
    document.body.classList.add('light')
  } else if (t === 'high-contrast') {
    document.body.classList.add('high-contrast')
  }
}

onMounted(() => {
  const saved = localStorage.getItem('theme') as Theme | null
  if (saved && ['light', 'dark', 'high-contrast'].includes(saved)) {
    theme.value = saved
    applyTheme(saved)
  }
  load()
})
</script>
