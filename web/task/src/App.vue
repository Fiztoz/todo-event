<template>
  <nav class="nav">
    <span class="nav-logo">Tasks</span>
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
import { onMounted, ref } from 'vue'
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

const statuses: TaskStatus[] = ['pending', 'in_progress', 'done']

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

onMounted(load)
</script>
