<template>
  <nav class="nav">
    <span class="nav-logo">Todoe</span>
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
          <TaskItem :task="task" @updated="onUpdated" />
        </li>
      </ul>
    </template>
  </main>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import CreateTask from './components/CreateTask.vue'
import TaskItem from './components/TaskItem.vue'
import { listTasks } from './api.js'

const tasks = ref([])
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    tasks.value = await listTasks()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function onCreated(task) {
  tasks.value.unshift(task)
}

function onUpdated(updated) {
  const i = tasks.value.findIndex(t => t.id === updated.id)
  if (i !== -1) tasks.value[i] = updated
}

onMounted(load)
</script>
