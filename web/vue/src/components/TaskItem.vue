<template>
  <div class="task-item" :class="task.status">
    <div class="task-body">
      <div class="task-title">{{ task.title }}</div>
      <span class="badge" :class="task.status">{{ label }}</span>
    </div>
    <div class="task-actions">
      <button
        v-for="s in nextStatuses"
        :key="s"
        class="btn-ghost"
        @click="change(s)"
        :disabled="loading"
      >{{ actionLabel(s) }}</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { changeStatus } from '../api.js'

const props = defineProps({ task: Object })
const emit = defineEmits(['updated'])
const loading = ref(false)

const transitions = {
  pending:     ['in_progress'],
  in_progress: ['done', 'pending'],
  done:        [],
}

const statusLabels = {
  pending:     'Pending',
  in_progress: 'In Progress',
  done:        'Done',
}

const actionLabels = {
  pending:     'Reset',
  in_progress: 'Start',
  done:        'Complete',
}

const label = computed(() => statusLabels[props.task.status] ?? props.task.status)
const nextStatuses = computed(() => transitions[props.task.status] ?? [])
const actionLabel = s => actionLabels[s] ?? s

async function change(status) {
  loading.value = true
  try {
    const updated = await changeStatus(props.task.id, status)
    emit('updated', updated)
  } finally {
    loading.value = false
  }
}
</script>
