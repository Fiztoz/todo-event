<template>
  <div class="step-card">
    <div class="card-title">Prove you're human</div>
    <div class="card-desc">Solve the math problem to continue.</div>

    <div v-if="!challenge && !loadError" class="info-row">Loading challenge…</div>

    <template v-if="challenge">
      <div class="form-field">
        <label for="answer">{{ challenge.question }} = ?</label>
        <input
          id="answer"
          v-model.number="answer"
          type="number"
          inputmode="numeric"
          placeholder="Your answer"
          @keydown.enter="submit"
        />
      </div>

      <button class="btn-primary" :disabled="!hasAnswer || loading" @click="submit">
        {{ loading ? 'Verifying…' : 'Verify' }}
      </button>
    </template>

    <button v-if="loadError" class="btn-primary" @click="loadChallenge">Retry</button>

    <div v-if="error" class="form-error">{{ error }}</div>
    <div v-if="loadError" class="form-error">{{ loadError }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { issueCaptcha, verifyCaptcha } from '../api.ts'
import type { Captcha } from '../types.ts'

const emit = defineEmits<{ done: [] }>()

const challenge = ref<Captcha | null>(null)
const answer = ref<number | string>('')
const loading = ref(false)
const error = ref('')
const loadError = ref('')

const hasAnswer = computed(() => answer.value !== '' && answer.value !== null)

async function loadChallenge(): Promise<void> {
  loadError.value = ''
  error.value = ''
  answer.value = ''
  try {
    challenge.value = await issueCaptcha()
  } catch (e) {
    challenge.value = null
    loadError.value = (e as Error).message
  }
}

onMounted(loadChallenge)

async function submit(): Promise<void> {
  if (!challenge.value || !hasAnswer.value) return
  loading.value = true
  error.value = ''
  try {
    await verifyCaptcha(challenge.value.id, Number(answer.value))
    emit('done')
  } catch (e) {
    error.value = (e as Error).message
    answer.value = ''
    if (/expired|already/i.test(error.value)) {
      await loadChallenge()
    }
  } finally {
    loading.value = false
  }
}
</script>
