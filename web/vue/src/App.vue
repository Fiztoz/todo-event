<template>
  <template v-if="!authenticated">
    <LoginPage @logged-in="onLoggedIn" />
  </template>

  <template v-else>
    <nav class="nav">
      <span class="nav-logo">Todoe</span>
      <button class="btn-logout" @click="onLogout">Sign out</button>
    </nav>

    <main class="page">
      <div class="page-header">
        <h1>My Tasks</h1>
        <p>Track and manage your work</p>
      </div>

      <CreateTask @created="onCreated" />

      <div v-if="loading" class="state-msg">Loading tasks…</div>
      <div v-else-if="error" class="state-msg" style="color: var(--red)">{{ error }}</div>
      <template v-else>
        <div class="section-label">{{ tasks.length }} task{{ tasks.length !== 1 ? 's' : '' }}</div>
        <div v-if="tasks.length === 0" class="state-msg">No tasks yet. Add one above.</div>
        <ul v-else class="task-list">
          <li v-for="task in tasks" :key="task.id">
            <TaskItem :task="task" @updated="onUpdated" @renamed="onRenamed" />
          </li>
        </ul>
      </template>
    </main>
  </template>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import CreateTask from './components/CreateTask.vue'
import TaskItem from './components/TaskItem.vue'
import LoginPage from './components/LoginPage.vue'
import { listTasks, logout, getToken, clearToken, AuthError } from './api.ts'
import type { Task } from './types.ts'

const authenticated = ref(false)
const tasks = ref<Task[]>([])
const loading = ref(false)
const error = ref('')

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    tasks.value = await listTasks()
  } catch (e) {
    if (e instanceof AuthError) {
      clearToken()
      authenticated.value = false
    } else {
      error.value = (e as Error).message
    }
  } finally {
    loading.value = false
  }
}

function onLoggedIn(): void {
  authenticated.value = true
  load()
}

async function onLogout(): Promise<void> {
  await logout()
  authenticated.value = false
  tasks.value = []
}

function onCreated(task: Task): void {
  tasks.value.unshift(task)
}

function onUpdated(updated: Task): void {
  const i = tasks.value.findIndex(t => t.id === updated.id)
  if (i !== -1) tasks.value[i] = updated
}

function onRenamed(task: Task): void {
  const i = tasks.value.findIndex(t => t.id === task.id)
  if (i !== -1) tasks.value[i] = task
}

onMounted(async () => {
  if (getToken()) {
    authenticated.value = true
    await load()
  }
})
</script>

<style>
.btn-logout {
  margin-left: auto;
  height: 34px;
  padding: 0 14px;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.82rem;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}
.btn-logout:hover {
  border-color: var(--red);
  color: var(--red);
}
</style>
