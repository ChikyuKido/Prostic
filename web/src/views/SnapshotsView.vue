<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RefreshCw } from 'lucide-vue-next'

import { apiJson } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

interface Snapshot {
  id: number
  snapshotID: string
  time: string
  backupID: string
  vmid: number | null
  name: string
  vmType: string
  snapshotType: string
  backupDate: string
  srcFile: string
  destFile: string
  hostname: string
}

const loading = ref(true)
const refreshing = ref(false)
const error = ref('')
const snapshots = ref<Snapshot[]>([])

async function loadSnapshots() {
  loading.value = true
  error.value = ''

  try {
    const response = await apiJson<{ snapshots: Snapshot[] }>('/api/snapshots')
    snapshots.value = response.snapshots
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load snapshots'
  } finally {
    loading.value = false
  }
}

async function refreshCache() {
  refreshing.value = true
  error.value = ''

  try {
    await apiJson<{ count: number }>('/api/snapshots/refresh', { method: 'POST' })
    await loadSnapshots()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to refresh cache'
  } finally {
    refreshing.value = false
  }
}

onMounted(() => {
  void loadSnapshots()
})

defineExpose({
  refresh: loadSnapshots,
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-3xl font-semibold tracking-tight">Snapshots</h1>
        <p class="text-sm text-muted-foreground">All cached restic snapshots from the latest refresh.</p>
      </div>
      <Button variant="outline" :disabled="refreshing" @click="refreshCache">
        <RefreshCw class="mr-2 size-4" />
        {{ refreshing ? 'Refreshing...' : 'Refresh cache' }}
      </Button>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card class="border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle>Snapshot Cache</CardTitle>
        <CardDescription>{{ snapshots.length }} row(s) currently cached.</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-sm text-muted-foreground">Loading snapshots...</div>
        <div v-else-if="snapshots.length === 0" class="text-sm text-muted-foreground">
          No cached snapshots yet. Run a refresh first.
        </div>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Snapshot</TableHead>
              <TableHead>Backup</TableHead>
              <TableHead>VM</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Source</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="snapshot in snapshots" :key="snapshot.snapshotID">
              <TableCell class="whitespace-nowrap text-muted-foreground">
                {{ new Date(snapshot.time).toLocaleString() }}
              </TableCell>
              <TableCell class="font-mono text-xs">{{ snapshot.snapshotID }}</TableCell>
              <TableCell class="font-mono text-xs">{{ snapshot.backupID || '-' }}</TableCell>
              <TableCell>
                <div class="font-medium">{{ snapshot.name || 'Unknown' }}</div>
                <div class="text-xs text-muted-foreground">
                  {{ snapshot.vmType || '-' }} {{ snapshot.vmid ?? '-' }}
                </div>
              </TableCell>
              <TableCell class="capitalize">{{ snapshot.snapshotType || '-' }}</TableCell>
              <TableCell class="max-w-md truncate text-xs text-muted-foreground">
                {{ snapshot.srcFile || snapshot.destFile || snapshot.hostname || '-' }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
