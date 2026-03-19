<script setup lang="ts">
import { Database, LayoutDashboard, RefreshCw, LogOut } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { apiJson, clearToken } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

const route = useRoute()
const router = useRouter()

const emit = defineEmits<{
  refreshed: []
}>()

async function refreshSnapshots() {
  await apiJson<{ count: number }>('/api/snapshots/refresh', { method: 'POST' })
  emit('refreshed')
}

async function logout() {
  clearToken()
  await router.replace('/login')
}
</script>

<template>
  <Card class="w-full border-border/70 bg-card/95 shadow-sm backdrop-blur-sm lg:w-72">
    <CardHeader class="space-y-2">
      <CardTitle class="text-xl">Prostic</CardTitle>
      <CardDescription>Cached restic metadata for overview and snapshot browsing.</CardDescription>
    </CardHeader>
    <CardContent class="space-y-3">
      <Button
        class="w-full justify-start"
        :variant="route.name === 'overview' ? 'default' : 'outline'"
        @click="router.push('/')"
      >
        <LayoutDashboard class="mr-2 size-4" />
        Overview
      </Button>
      <Button
        class="w-full justify-start"
        :variant="route.name === 'snapshots' ? 'default' : 'outline'"
        @click="router.push('/snapshots')"
      >
        <Database class="mr-2 size-4" />
        Snapshots
      </Button>
      <Button class="w-full justify-start" variant="outline" @click="refreshSnapshots">
        <RefreshCw class="mr-2 size-4" />
        Refresh Cache
      </Button>
      <Button class="w-full justify-start" variant="outline" @click="logout">
        <LogOut class="mr-2 size-4" />
        Logout
      </Button>
    </CardContent>
  </Card>
</template>
