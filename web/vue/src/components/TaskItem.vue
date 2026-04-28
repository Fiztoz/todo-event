<template>
  <div class="task-item" :class="task.status">
    <div class="task-body">
      <div
        v-if="!editing"
        class="task-title"
        @click="startEdit"
        style="cursor: pointer"
      >{{ task.title }}</div>
      <input
        v-else
        ref="titleInput"
        class="task-title-input"
        v-model="editTitle"
        @keydown.enter="saveEdit"
        @keydown.escape="cancelEdit"
        @blur="cancelEdit"
      />
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
import { ref, computed, nextTick } from 'vue'
import { changeStatus } from '../api.ts'
import type { Status, Task } from '../types.ts'

const props = defineProps<{ task: Task }>()
const emit = defineEmits<{ updated: [task: Task]; renamed: [task: Task] }>()
const loading = ref(false)
const editing = ref(false)
const editTitle = ref('')
const titleInput = ref<HTMLInputElement | null>(null)

function startEdit(): void {
  editTitle.value = props.task.title
  editing.value = true
  nextTick(() => titleInput.value?.select())
}

function saveEdit(): void {
  const trimmed = editTitle.value.trim()
  if (trimmed && trimmed !== props.task.title) {
    emit('renamed', { ...props.task, title: trimmed })
  }
  editing.value = false
}

function cancelEdit(): void {
  editing.value = false
}

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
