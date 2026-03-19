<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RefreshCw } from 'lucide-vue-next'

import { apiJson } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface Overview {
  totalSnapshots: number
  totalBackups: number
  totalVMs: number
  diskSnapshots: number
  configSnapshots: number
  latestSnapshot: string | null
}

const loading = ref(true)
const error = ref('')
const refreshing = ref(false)
const overview = ref<Overview | null>(null)

async function loadOverview() {
  loading.value = true
  error.value = ''

  try {
    overview.value = await apiJson<Overview>('/api/overview')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load overview'
  } finally {
    loading.value = false
  }
}

async function refreshCache() {
  refreshing.value = true
  error.value = ''

  try {
    await apiJson<{ count: number }>('/api/snapshots/refresh', { method: 'POST' })
    await loadOverview()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to refresh cache'
  } finally {
    refreshing.value = false
  }
}

onMounted(() => {
  void loadOverview()
})

defineExpose({
  refresh: loadOverview,
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-3xl font-semibold tracking-tight">Overview</h1>
        <p class="text-sm text-muted-foreground">Cached snapshot metadata from the last restic refresh.</p>
      </div>
      <Button variant="outline" :disabled="refreshing" @click="refreshCache">
        <RefreshCw class="mr-2 size-4" />
        {{ refreshing ? 'Refreshing...' : 'Refresh cache' }}
      </Button>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <Card v-for="card in 4" :key="card" class="border-border/70 bg-card/95">
        <CardContent class="p-6 text-sm text-muted-foreground">Loading...</CardContent>
      </Card>
    </div>

    <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <Card class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <CardDescription>Total snapshots</CardDescription>
          <CardTitle class="text-3xl">{{ overview?.totalSnapshots ?? 0 }}</CardTitle>
        </CardHeader>
      </Card>
      <Card class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <CardDescription>Backup runs</CardDescription>
          <CardTitle class="text-3xl">{{ overview?.totalBackups ?? 0 }}</CardTitle>
        </CardHeader>
      </Card>
      <Card class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <CardDescription>VMs / LXCs</CardDescription>
          <CardTitle class="text-3xl">{{ overview?.totalVMs ?? 0 }}</CardTitle>
        </CardHeader>
      </Card>
      <Card class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <CardDescription>Disk snapshots</CardDescription>
          <CardTitle class="text-3xl">{{ overview?.diskSnapshots ?? 0 }}</CardTitle>
        </CardHeader>
      </Card>
    </div>

    <Card class="border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle>Latest Snapshot</CardTitle>
        <CardDescription>Most recent snapshot currently in the cache.</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="text-sm text-foreground">
          {{ overview?.latestSnapshot ? new Date(overview.latestSnapshot).toLocaleString() : 'No cached snapshots yet.' }}
        </div>
        <p class="mt-2 text-sm text-muted-foreground">
          Config snapshots: {{ overview?.configSnapshots ?? 0 }}
        </p>
      </CardContent>
    </Card>
  </div>
</template>
