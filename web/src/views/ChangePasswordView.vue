<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { apiJson, setNeedsPasswordChange } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

const router = useRouter()
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')

const submitDisabled = computed(() => loading.value || !password.value || !confirmPassword.value)

async function submit() {
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  if (!password.value) {
    error.value = 'Password is required'
    return
  }

  loading.value = true
  error.value = ''

  try {
    await apiJson<void>('/api/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ password: password.value }),
    })
    setNeedsPasswordChange(false)
    await router.replace('/')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to change password'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card class="w-full max-w-md border-border/60 bg-card/95 backdrop-blur-sm">
    <CardHeader class="space-y-2">
      <CardTitle class="text-2xl">Change Default Password</CardTitle>
      <CardDescription>
        The default `admin` password is still active. Set a new password before using the UI.
      </CardDescription>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" @submit.prevent="submit">
        <div class="space-y-2">
          <div class="text-sm font-medium text-foreground">New password</div>
          <Input
            v-model="password"
            type="password"
            autocomplete="new-password"
            placeholder="Enter new password"
          />
        </div>
        <div class="space-y-2">
          <div class="text-sm font-medium text-foreground">Confirm password</div>
          <Input
            v-model="confirmPassword"
            type="password"
            autocomplete="new-password"
            placeholder="Repeat new password"
          />
        </div>
        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
        <Button class="w-full" :disabled="submitDisabled" type="submit">
          {{ loading ? 'Saving...' : 'Update password' }}
        </Button>
      </form>
    </CardContent>
    <CardFooter class="text-sm text-muted-foreground">
      This clears the forced password change flag in the database.
    </CardFooter>
  </Card>
</template>
