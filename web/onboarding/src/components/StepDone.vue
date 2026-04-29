<template>
  <div class="step-card done-card">
    <div class="done-icon">🎉</div>
    <h2>Welcome aboard, {{ user.name }}!</h2>
    <p>Your account is fully set up. You're ready to go.</p>
  </div>
</template>

<script setup lang="ts">
import type { User } from '../types.ts'
defineProps<{ user: User }>()
</script>
