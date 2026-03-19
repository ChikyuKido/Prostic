<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { apiJson, setNeedsPasswordChange, setToken } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

const router = useRouter()
const loading = ref(false)
const password = ref('')
const error = ref('')

async function submit() {
  if (!password.value) {
    error.value = 'Password is required'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await apiJson<{ token: string; needsPasswordChange: boolean }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ password: password.value }),
    })

    setToken(response.token)
    setNeedsPasswordChange(response.needsPasswordChange)
    await router.replace(response.needsPasswordChange ? '/change-password' : '/')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Request failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card class="w-full max-w-md border-border/60 bg-card/95 backdrop-blur-sm">
    <CardHeader class="space-y-2">
      <CardTitle class="text-2xl">Prostic Admin</CardTitle>
      <CardDescription>Enter the admin password to access the web UI.</CardDescription>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" @submit.prevent="submit">
        <div class="space-y-2">
          <div class="text-sm font-medium text-foreground">Password</div>
          <Input
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="Enter password"
          />
        </div>
        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
        <Button class="w-full" :disabled="loading" type="submit">
          {{ loading ? 'Please wait...' : 'Login' }}
        </Button>
      </form>
    </CardContent>
    <CardFooter class="text-sm text-muted-foreground">
      First login uses the default password `admin` until it is changed.
    </CardFooter>
  </Card>
</template>
