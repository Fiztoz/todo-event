<template>
  <div class="create-card">
    <form class="create-form" @submit.prevent="submit">
      <input
        v-model="title"
        placeholder="What needs to be done?"
        :disabled="loading"
        autocomplete="off"
      />
      <button class="btn-primary" type="submit" :disabled="!title.trim() || loading">
        Add task
      </button>
    </form>
    <p v-if="error" class="form-error">{{ error }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { createTask } from '../api.js'

const emit = defineEmits(['created'])
const title = ref('')
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const task = await createTask(title.value.trim())
    title.value = ''
    emit('created', task)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>
