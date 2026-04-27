<template>
  <div class="task-item" :class="task.status">
    <div class="task-body">
      <div class="task-title">{{ task.title }}</div>
      <span class="badge" :class="task.status">{{ statusLabels[task.status] }}</span>
    </div>
    <div class="task-actions">
      <button
        v-for="s in nextStatuses"
        :key="s"
        class="btn-ghost"
        @click="change(s)"
        :disabled="loading"
      >{{ actionLabels[s] }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { changeStatus } from '../api.ts'
import type { Status, Task } from '../types.ts'

const props = defineProps<{ task: Task }>()
const emit = defineEmits<{ updated: [task: Task] }>()
const loading = ref(false)

const transitions: Record<Status, Status[]> = {
  pending:     ['in_progress'],
  in_progress: ['done', 'pending'],
  done:        [],
}

const statusLabels: Record<Status, string> = {
  pending:     'Pending',
  in_progress: 'In Progress',
  done:        'Done',
}

const actionLabels: Record<Status, string> = {
  pending:     'Reset',
  in_progress: 'Start',
  done:        'Complete',
}

const nextStatuses = computed(() => transitions[props.task.status])

async function change(status: Status): Promise<void> {
  loading.value = true
  try {
    const updated = await changeStatus(props.task.id, status)
    emit('updated', updated)
  } finally {
    loading.value = false
  }
}
</script>
