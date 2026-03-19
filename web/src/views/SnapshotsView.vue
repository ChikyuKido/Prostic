<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { RefreshCw, Trash2 } from 'lucide-vue-next'

import { apiJson } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
  existsInConfig: boolean
}

interface Task {
  id: number
  purpose: string
  status: string
  logs: string
  startedAt: string
  finishedAt: string | null
}

interface TaskStatus {
  running: boolean
  kind?: string
  purpose?: string
  startedAt?: string | null
}

const loading = ref(true)
const refreshing = ref(false)
const deleteSnapshotID = ref<string | null>(null)
const error = ref('')
const snapshots = ref<Snapshot[]>([])
const taskStatus = ref<TaskStatus>({ running: false })
let statusTimer: number | null = null

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

async function loadTaskStatus() {
  try {
    const wasRunning = taskStatus.value.running
    const nextStatus = await apiJson<TaskStatus>('/api/tasks/status')
    taskStatus.value = nextStatus

    if (wasRunning && !nextStatus.running) {
      await loadSnapshots()
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load task status'
  }
}

async function refreshCache() {
  refreshing.value = true
  error.value = ''

  try {
    await apiJson<{ snapshotCount: number }>('/api/refresh', { method: 'POST' })
    await loadSnapshots()
    await loadTaskStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to refresh cache'
  } finally {
    refreshing.value = false
  }
}

async function startDeleteSnapshot(snapshot: Snapshot) {
  if (taskStatus.value.running) {
    return
  }
  if (!window.confirm(`Delete snapshot ${snapshot.snapshotID}?`)) {
    return
  }

  deleteSnapshotID.value = snapshot.snapshotID
  error.value = ''

  try {
    await apiJson<{ task: Task }>('/api/tasks/delete-snapshot', {
      method: 'POST',
      body: JSON.stringify({
        snapshot: {
          snapshotID: snapshot.snapshotID,
          time: snapshot.time,
          backupID: snapshot.backupID,
          vmid: snapshot.vmid,
          name: snapshot.name,
          vmType: snapshot.vmType,
          snapshotType: snapshot.snapshotType,
          srcFile: snapshot.srcFile,
        },
      }),
    })
    await loadTaskStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to start delete snapshot task'
  } finally {
    deleteSnapshotID.value = null
  }
}

onMounted(() => {
  void loadSnapshots()
  void loadTaskStatus()
  statusTimer = window.setInterval(() => {
    void loadTaskStatus()
  }, 3000)
})

onUnmounted(() => {
  if (statusTimer !== null) {
    window.clearInterval(statusTimer)
  }
})

defineExpose({
  refresh: loadSnapshots,
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h1 class="text-3xl font-semibold tracking-tight">Snapshots</h1>
      <Button variant="outline" :disabled="refreshing || taskStatus.running" @click="refreshCache">
        <RefreshCw class="mr-2 size-4" />
        {{ refreshing ? 'Refreshing...' : 'Refresh' }}
      </Button>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <p v-if="taskStatus.running" class="text-sm text-muted-foreground">
      Running {{ taskStatus.kind || 'job' }} since {{ taskStatus.startedAt ? new Date(taskStatus.startedAt).toLocaleString() : 'now' }}.
    </p>

    <Card class="border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle>Snapshot Cache</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-sm text-muted-foreground">Loading snapshots...</div>
        <div v-else-if="snapshots.length === 0" class="text-sm text-muted-foreground">
          No snapshots.
        </div>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Snapshot</TableHead>
              <TableHead>Backup</TableHead>
              <TableHead>VM</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>In Config</TableHead>
              <TableHead>Source</TableHead>
              <TableHead class="w-[120px]">Action</TableHead>
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
              <TableCell>
                <span :class="snapshot.existsInConfig ? 'text-foreground' : 'text-destructive'">
                  {{ snapshot.existsInConfig ? 'Yes' : 'No' }}
                </span>
              </TableCell>
              <TableCell class="max-w-md truncate text-xs text-muted-foreground">
                {{ snapshot.srcFile || snapshot.destFile || snapshot.hostname || '-' }}
              </TableCell>
              <TableCell>
                <Button
                  size="sm"
                  variant="destructive"
                  :disabled="taskStatus.running || deleteSnapshotID === snapshot.snapshotID"
                  @click="startDeleteSnapshot(snapshot)"
                >
                  <Trash2 class="size-4" />
                  {{ deleteSnapshotID === snapshot.snapshotID ? 'Starting...' : 'Delete' }}
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
