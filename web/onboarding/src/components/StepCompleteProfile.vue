<template>
  <div class="step-card">
    <div class="card-title">Complete your profile</div>
    <div class="card-desc">One last step — tell us a bit about yourself.</div>

    <div class="credit-result approved">
      Credit approved — score <span class="credit-score">{{ user.credit_score }}</span>
    </div>

    <div class="form-field">
      <label for="bio">Bio</label>
      <textarea id="bio" v-model="bio" placeholder="A short bio about yourself…"></textarea>
    </div>

    <button class="btn-primary" :disabled="loading" @click="submit">
      {{ loading ? 'Saving…' : 'Complete profile' }}
    </button>
    <div v-if="error" class="form-error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { completeProfile } from '../api.ts'
import type { User } from '../types.ts'

const props = defineProps<{ user: User }>()
const emit = defineEmits<{ done: [user: User] }>()

const bio = ref('')
const loading = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const updated = await completeProfile(props.user.id, bio.value)
    emit('done', updated)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
</script>
