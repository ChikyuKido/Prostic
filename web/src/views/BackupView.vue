<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Play, RefreshCw, Save } from 'lucide-vue-next'

import { apiJson } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

interface BackupStatus {
  running: boolean
  runnerBusy: boolean
  runnerKind: string
  runnerPurpose: string
  backupRunID?: number
  backupID: string
  trigger: string
  startedAt: string | null
  totalItems: number
  completedItems: number
  currentVMID?: number | null
  currentVMName: string
  currentItemType: string
  currentSrcFile: string
  currentDestFile: string
  currentBytesDone: number
  currentBytesTotal: number
  currentItemStarted: string | null
  lastMessage: string
  cronExpression: string
}

interface BackupRun {
  id: number
  backupID: string
  trigger: string
  status: string
  logs: string
  totalItems: number
  completedItems: number
  startedAt: string
  finishedAt: string | null
}

const loading = ref(true)
const startLoading = ref(false)
const saveLoading = ref(false)
const error = ref('')
const status = ref<BackupStatus | null>(null)
const runs = ref<BackupRun[]>([])
const cronExpression = ref('')
const logsRun = ref<BackupRun | null>(null)
let statusTimer: number | null = null

const overallPercent = computed(() => {
  if (!status.value || status.value.totalItems <= 0) {
    return 0
  }

  return Math.round((status.value.completedItems / status.value.totalItems) * 100)
})

const currentPercent = computed(() => {
  if (!status.value || status.value.currentBytesTotal <= 0) {
    return 0
  }

  return Math.max(0, Math.min(100, Math.round((status.value.currentBytesDone / status.value.currentBytesTotal) * 100)))
})

const etaLabel = computed(() => {
  if (!status.value || !status.value.currentItemStarted || status.value.currentBytesDone <= 0 || status.value.currentBytesTotal <= 0) {
    return '-'
  }

  const elapsedSeconds = (Date.now() - new Date(status.value.currentItemStarted).getTime()) / 1000
  if (elapsedSeconds <= 0) {
    return '-'
  }

  const rate = status.value.currentBytesDone / elapsedSeconds
  if (rate <= 0) {
    return '-'
  }

  const remainingSeconds = Math.max(0, Math.round((status.value.currentBytesTotal - status.value.currentBytesDone) / rate))
  const minutes = Math.floor(remainingSeconds / 60)
  const seconds = remainingSeconds % 60
  return `${minutes}m ${seconds}s`
})

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

function statusClass(value: string) {
  if (value === 'success') {
    return 'text-emerald-600'
  }
  if (value === 'failed') {
    return 'text-destructive'
  }
  return 'text-amber-600'
}

async function loadStatus() {
  try {
    const wasRunning = status.value?.running ?? false
    const nextStatus = await apiJson<BackupStatus>('/api/backup/status')
    status.value = nextStatus
    cronExpression.value = nextStatus.cronExpression || ''

    if (wasRunning && !nextStatus.running) {
      await loadRuns()
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load backup status'
  }
}

async function loadRuns() {
  loading.value = true
  error.value = ''

  try {
    const response = await apiJson<{ runs: BackupRun[] }>('/api/backup/runs')
    runs.value = response.runs
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load backup runs'
  } finally {
    loading.value = false
  }
}

async function startBackup() {
  if (!status.value || status.value.runnerBusy) {
    return
  }

  startLoading.value = true
  error.value = ''

  try {
    await apiJson('/api/backup/start', { method: 'POST' })
    await loadStatus()
    await loadRuns()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to start backup'
  } finally {
    startLoading.value = false
  }
}

async function saveCron() {
  saveLoading.value = true
  error.value = ''

  try {
    await apiJson('/api/backup/settings', {
      method: 'PUT',
      body: JSON.stringify({ cronExpression: cronExpression.value }),
    })
    await loadStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to save cron expression'
  } finally {
    saveLoading.value = false
  }
}

async function refreshAll() {
  await loadStatus()
  await loadRuns()
}

onMounted(() => {
  void refreshAll()
  statusTimer = window.setInterval(() => {
    void loadStatus()
  }, 2000)
})

onUnmounted(() => {
  if (statusTimer !== null) {
    window.clearInterval(statusTimer)
  }
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h1 class="text-3xl font-semibold tracking-tight">Backup</h1>
      <div class="flex flex-col gap-2 sm:flex-row">
        <Button variant="outline" @click="refreshAll">
          <RefreshCw class="mr-2 size-4" />
          Refresh
        </Button>
        <Button :disabled="startLoading || status?.runnerBusy" @click="startBackup">
          <Play class="mr-2 size-4" />
          {{ startLoading ? 'Starting...' : 'Start Backup' }}
        </Button>
      </div>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
      <Card class="border-border/70 bg-card/95">
        <CardHeader>
          <CardTitle>Live Progress</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div v-if="status?.running" class="space-y-4">
            <div class="space-y-2">
              <div class="flex items-center justify-between text-sm">
                <span class="text-muted-foreground">Overall</span>
                <span>{{ status.completedItems }} / {{ status.totalItems }}</span>
              </div>
              <div class="h-3 rounded-full bg-muted">
                <div class="h-3 rounded-full bg-primary transition-all" :style="{ width: `${overallPercent}%` }" />
              </div>
            </div>

            <div class="space-y-2">
              <div class="flex items-center justify-between text-sm">
                <span class="text-muted-foreground">Current Item</span>
                <span>{{ currentPercent }}%</span>
              </div>
              <div class="h-3 rounded-full bg-muted">
                <div class="h-3 rounded-full bg-chart-2 transition-all" :style="{ width: `${currentPercent}%` }" />
              </div>
            </div>

            <div class="grid gap-2 text-sm text-muted-foreground md:grid-cols-2">
              <div><span class="text-foreground">Backup ID:</span> {{ status.backupID || '-' }}</div>
              <div><span class="text-foreground">Trigger:</span> {{ status.trigger || '-' }}</div>
              <div><span class="text-foreground">VM:</span> {{ status.currentVMName || '-' }}</div>
              <div><span class="text-foreground">Type:</span> {{ status.currentItemType || '-' }}</div>
              <div class="md:col-span-2"><span class="text-foreground">Source:</span> {{ status.currentSrcFile || '-' }}</div>
              <div><span class="text-foreground">Done:</span> {{ formatBytes(status.currentBytesDone) }}</div>
              <div><span class="text-foreground">Total:</span> {{ formatBytes(status.currentBytesTotal) }}</div>
              <div><span class="text-foreground">ETA:</span> {{ etaLabel }}</div>
              <div><span class="text-foreground">Started:</span> {{ status.startedAt ? new Date(status.startedAt).toLocaleString() : '-' }}</div>
            </div>

            <div v-if="status.lastMessage" class="rounded-lg border border-border/70 bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
              {{ status.lastMessage }}
            </div>
          </div>

          <div v-else class="space-y-2 text-sm text-muted-foreground">
            <div v-if="status?.runnerBusy">
              Another job is currently running: {{ status.runnerKind || 'job' }} {{ status.runnerPurpose || '' }}
            </div>
            <div v-else>No backup is currently running.</div>
          </div>
        </CardContent>
      </Card>

      <Card class="border-border/70 bg-card/95">
        <CardHeader>
          <CardTitle>Schedule</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <div class="text-sm text-muted-foreground">Cron Expression</div>
            <Input v-model="cronExpression" placeholder="0 2 * * *" />
          </div>
          <div class="text-sm text-muted-foreground">
            Empty disables scheduling. The expression is evaluated in the server's local timezone.
          </div>
          <Button :disabled="saveLoading" @click="saveCron">
            <Save class="mr-2 size-4" />
            {{ saveLoading ? 'Saving...' : 'Save Schedule' }}
          </Button>
        </CardContent>
      </Card>
    </div>

    <Card class="border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle>Backup History</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-sm text-muted-foreground">Loading backup runs...</div>
        <div v-else-if="runs.length === 0" class="text-sm text-muted-foreground">No backup runs yet.</div>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Started</TableHead>
              <TableHead>Trigger</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Backup ID</TableHead>
              <TableHead>Items</TableHead>
              <TableHead>Finished</TableHead>
              <TableHead class="w-[120px]">Logs</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="run in runs" :key="run.id">
              <TableCell class="whitespace-nowrap text-muted-foreground">
                {{ new Date(run.startedAt).toLocaleString() }}
              </TableCell>
              <TableCell class="capitalize">{{ run.trigger }}</TableCell>
              <TableCell>
                <span class="font-medium capitalize" :class="statusClass(run.status)">{{ run.status }}</span>
              </TableCell>
              <TableCell class="font-mono text-xs">{{ run.backupID || '-' }}</TableCell>
              <TableCell>{{ run.completedItems }} / {{ run.totalItems }}</TableCell>
              <TableCell class="whitespace-nowrap text-muted-foreground">
                {{ run.finishedAt ? new Date(run.finishedAt).toLocaleString() : '-' }}
              </TableCell>
              <TableCell>
                <Button variant="outline" size="sm" @click="logsRun = run">View</Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog :open="logsRun !== null" @update:open="(open) => { if (!open) logsRun = null }">
      <DialogContent class="max-w-4xl">
        <DialogHeader>
          <DialogTitle>Backup Run Logs</DialogTitle>
          <DialogDescription>
            {{ logsRun ? new Date(logsRun.startedAt).toLocaleString() : '' }}
          </DialogDescription>
        </DialogHeader>
        <div class="max-h-[480px] overflow-auto rounded-lg border border-border/70 bg-muted/30 p-4">
          <pre class="whitespace-pre-wrap break-words text-xs text-foreground">{{ logsRun?.logs || 'No logs.' }}</pre>
        </div>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">Close</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
