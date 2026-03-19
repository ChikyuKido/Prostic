<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RefreshCw } from 'lucide-vue-next'

import { apiJson } from '@/lib/api'
import RepoStatsChart from '@/components/RepoStatsChart.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface Overview {
  totalSnapshots: number
  totalBackups: number
  totalVMs: number
  diskSnapshots: number
  configSnapshots: number
  latestSnapshot: string | null
  totalSize: number
  totalUncompressedSize: number
  compressionRatio: number
  compressionSpaceSaving: number
  totalBlobCount: number
  repoSnapshotsCount: number
  lastRefreshedAt: string | null
  history: {
    timestamp: string
    totalSize: number
    totalUncompressedSize: number
  }[]
}

const loading = ref(true)
const error = ref('')
const refreshing = ref(false)
const overview = ref<Overview | null>(null)

function formatBytes(value: number) {
  if (!value) {
    return '0 B'
  }

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = value
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }

  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

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
    await apiJson<{ snapshotCount: number }>('/api/refresh', { method: 'POST' })
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
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h1 class="text-3xl font-semibold tracking-tight">Overview</h1>
      <Button variant="outline" :disabled="refreshing" @click="refreshCache">
        <RefreshCw class="mr-2 size-4" />
        {{ refreshing ? 'Refreshing...' : 'Refresh' }}
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
          <div class="text-sm text-muted-foreground">Backups</div>
          <CardTitle class="text-3xl">{{ overview?.totalBackups ?? 0 }}</CardTitle>
        </CardHeader>
      </Card>
      <Card class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <div class="text-sm text-muted-foreground">Repo Size</div>
          <CardTitle class="text-3xl">{{ formatBytes(overview?.totalSize ?? 0) }}</CardTitle>
        </CardHeader>
      </Card>
      <Card class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <div class="text-sm text-muted-foreground">Raw Size</div>
          <CardTitle class="text-3xl">{{ formatBytes(overview?.totalUncompressedSize ?? 0) }}</CardTitle>
        </CardHeader>
      </Card>
      <Card class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <div class="text-sm text-muted-foreground">Compression</div>
          <CardTitle class="text-3xl">{{ (overview?.compressionRatio ?? 0).toFixed(2) }}x</CardTitle>
        </CardHeader>
      </Card>
    </div>

    <Card class="border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle>Repo History</CardTitle>
      </CardHeader>
      <CardContent>
        <RepoStatsChart :history="overview?.history ?? []" />
      </CardContent>
    </Card>

    <Card class="border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle>Latest</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid gap-2 text-sm text-muted-foreground md:grid-cols-2 xl:grid-cols-4">
          <div>
            <span class="text-foreground">Snapshots:</span>
            {{ overview?.totalSnapshots ?? 0 }}
          </div>
          <div>
            <span class="text-foreground">VMs:</span>
            {{ overview?.totalVMs ?? 0 }}
          </div>
          <div>
            <span class="text-foreground">Disks:</span>
            {{ overview?.diskSnapshots ?? 0 }}
          </div>
          <div>
            <span class="text-foreground">Configs:</span>
            {{ overview?.configSnapshots ?? 0 }}
          </div>
          <div>
            <span class="text-foreground">Repo snapshots:</span>
            {{ overview?.repoSnapshotsCount ?? 0 }}
          </div>
          <div>
            <span class="text-foreground">Blobs:</span>
            {{ overview?.totalBlobCount ?? 0 }}
          </div>
          <div>
            <span class="text-foreground">Space saved:</span>
            {{ ((overview?.compressionSpaceSaving ?? 0) * 100).toFixed(1) }}%
          </div>
          <div>
            <span class="text-foreground">Latest snapshot:</span>
            {{ overview?.latestSnapshot ? new Date(overview.latestSnapshot).toLocaleString() : 'None' }}
          </div>
          <div class="md:col-span-2 xl:col-span-4">
            <span class="text-foreground">Refreshed:</span>
            {{ overview?.lastRefreshedAt ? new Date(overview.lastRefreshedAt).toLocaleString() : 'Never' }}
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
