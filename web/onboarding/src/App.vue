<template>
  <nav class="nav">
    <span class="nav-logo">Todoe</span>
  </nav>

  <main class="page">
    <div class="step-bar">
      <div v-for="step in steps" :key="step.n" class="step-bar-item" :class="stepBarClass(step.n)">
        <div class="step-dot" :class="stepDotClass(step.n)">
          <svg v-if="stepDotClass(step.n) === 'completed'" width="12" height="12" viewBox="0 0 12 12" fill="none">
            <path d="M2 6l3 3 5-5" stroke="#fff" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <span v-else>{{ step.n }}</span>
        </div>
        <div class="step-label" :class="stepDotClass(step.n)">{{ step.label }}</div>
      </div>
    </div>

    <StepRegister v-if="!user" @done="user = $event" />

    <StepCaptcha v-else-if="user.status === 'registered' && !captchaPassed" @done="captchaPassed = true" />

    <StepVerifyEmail v-else-if="user.status === 'registered'" :user="user" @done="user = $event" />

    <StepCreditCheck v-else-if="user.status === 'email_verified'" :user="user" @done="user = $event" />

    <template v-else-if="user.status === 'credit_denied'">
      <div class="step-card">
        <div class="card-title">Application declined</div>
        <div class="card-desc">Unfortunately we can't proceed with your application.</div>
        <div class="credit-result denied">
          Credit denied — score <span class="credit-score">{{ user.credit_score }}</span>
        </div>
        <p style="font-size: 0.875rem; color: var(--text-muted);">
          Profile completion requires an approved credit check.
        </p>
      </div>
    </template>

    <StepCompleteProfile
      v-else-if="user.status === 'credit_approved'"
      :user="user"
      @done="user = $event"
    />

    <StepDone v-else-if="user.status === 'onboarding_complete'" :user="user" />
  </main>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import StepRegister from './components/StepRegister.vue'
import StepCaptcha from './components/StepCaptcha.vue'
import StepVerifyEmail from './components/StepVerifyEmail.vue'
import StepCreditCheck from './components/StepCreditCheck.vue'
import StepCompleteProfile from './components/StepCompleteProfile.vue'
import StepDone from './components/StepDone.vue'
import type { User, UserStatus } from './types.ts'

const user = ref<User | null>(null)
const captchaPassed = ref(false)

const steps = [
  { n: 1, label: 'Register' },
  { n: 2, label: 'Captcha' },
  { n: 3, label: 'Verify' },
  { n: 4, label: 'Credit' },
  { n: 5, label: 'Profile' },
]

const activeStep = computed((): number => {
  if (!user.value) return 1
  if (user.value.status === 'registered' && !captchaPassed.value) return 2
  const map: Record<UserStatus, number> = {
    registered:          3,
    email_verified:      4,
    credit_approved:     5,
    credit_denied:       4,
    onboarding_complete: 6,
  }
  return map[user.value.status]
})

function stepDotClass(n: number): string {
  if (n < activeStep.value) return 'completed'
  if (n === activeStep.value) return 'active'
  return ''
}

function stepBarClass(n: number): string {
  return n < activeStep.value ? 'completed' : ''
}
</script>
