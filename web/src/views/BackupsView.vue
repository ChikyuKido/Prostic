<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RefreshCw, Trash2 } from 'lucide-vue-next'

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

interface BackupGroup {
  backupID: string
  latestTime: string
  snapshotCount: number
  vmSummary: string
  allInConfig: boolean
  snapshots: Snapshot[]
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
const deleteBackupLoading = ref(false)
const error = ref('')
const backups = ref<BackupGroup[]>([])
const taskStatus = ref<TaskStatus>({ running: false })
const deleteBackupDialogOpen = ref(false)
const deleteBackupIDValue = ref('')
const deleteBackupCandidates = ref<Snapshot[]>([])
let statusTimer: number | null = null

const deleteBackupTitle = computed(() => {
  if (deleteBackupCandidates.value.length === 0) {
    return 'Nothing to delete'
  }

  return `Delete backup ${deleteBackupIDValue.value}`
})

function groupBackups(snapshots: Snapshot[]): BackupGroup[] {
  const groups = new Map<string, BackupGroup>()

  for (const snapshot of snapshots) {
    if (!snapshot.backupID) {
      continue
    }

    const existing = groups.get(snapshot.backupID)
    if (!existing) {
      groups.set(snapshot.backupID, {
        backupID: snapshot.backupID,
        latestTime: snapshot.time,
        snapshotCount: 1,
        vmSummary: snapshot.name || 'Unknown',
        allInConfig: snapshot.existsInConfig,
        snapshots: [snapshot],
      })
      continue
    }

    existing.snapshotCount += 1
    existing.snapshots.push(snapshot)
    existing.allInConfig = existing.allInConfig && snapshot.existsInConfig
    if (new Date(snapshot.time).getTime() > new Date(existing.latestTime).getTime()) {
      existing.latestTime = snapshot.time
    }
  }

  for (const group of groups.values()) {
    const names = [...new Set(group.snapshots.map((snapshot) => snapshot.name || 'Unknown'))]
    group.vmSummary = names.join(', ')
  }

  return [...groups.values()].sort((a, b) => new Date(b.latestTime).getTime() - new Date(a.latestTime).getTime())
}

async function loadBackups() {
  loading.value = true
  error.value = ''

  try {
    const response = await apiJson<{ snapshots: Snapshot[] }>('/api/snapshots')
    backups.value = groupBackups(response.snapshots)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load backups'
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
      await loadBackups()
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
    await loadBackups()
    await loadTaskStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to refresh cache'
  } finally {
    refreshing.value = false
  }
}

async function previewDeleteBackup(backup: BackupGroup) {
  if (taskStatus.value.running || !backup.backupID) {
    return
  }

  deleteBackupLoading.value = true
  error.value = ''
  deleteBackupIDValue.value = backup.backupID

  try {
    const response = await apiJson<{ snapshots: Snapshot[] }>('/api/tasks/delete-backup-id', {
      method: 'POST',
      body: JSON.stringify({ backupID: backup.backupID }),
    })
    deleteBackupCandidates.value = response.snapshots
    deleteBackupDialogOpen.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to prepare delete backup task'
  } finally {
    deleteBackupLoading.value = false
  }
}

async function confirmDeleteBackup() {
  deleteBackupLoading.value = true
  error.value = ''

  try {
    await apiJson<{ task: Task }>('/api/tasks/delete-backup-id?confirm=true', {
      method: 'POST',
      body: JSON.stringify({
        backupID: deleteBackupIDValue.value,
        snapshots: deleteBackupCandidates.value.map((snapshot) => ({
          snapshotID: snapshot.snapshotID,
          time: snapshot.time,
          backupID: snapshot.backupID,
          vmid: snapshot.vmid,
          name: snapshot.name,
          vmType: snapshot.vmType,
          snapshotType: snapshot.snapshotType,
          srcFile: snapshot.srcFile,
        })),
      }),
    })
    deleteBackupDialogOpen.value = false
    deleteBackupCandidates.value = []
    deleteBackupIDValue.value = ''
    await loadTaskStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to start delete backup task'
  } finally {
    deleteBackupLoading.value = false
  }
}

onMounted(() => {
  void loadBackups()
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
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h1 class="text-3xl font-semibold tracking-tight">Backup Sets</h1>
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
        <CardTitle>Backup Sets</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-sm text-muted-foreground">Loading backups...</div>
        <div v-else-if="backups.length === 0" class="text-sm text-muted-foreground">No backups.</div>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Backup ID</TableHead>
              <TableHead>VMs</TableHead>
              <TableHead>Snapshots</TableHead>
              <TableHead>In Config</TableHead>
              <TableHead class="w-[120px]">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="backup in backups" :key="backup.backupID">
              <TableCell class="whitespace-nowrap text-muted-foreground">
                {{ new Date(backup.latestTime).toLocaleString() }}
              </TableCell>
              <TableCell class="font-mono text-xs">{{ backup.backupID }}</TableCell>
              <TableCell class="max-w-md truncate">{{ backup.vmSummary }}</TableCell>
              <TableCell>{{ backup.snapshotCount }}</TableCell>
              <TableCell>
                <span :class="backup.allInConfig ? 'text-foreground' : 'text-destructive'">
                  {{ backup.allInConfig ? 'Yes' : 'No' }}
                </span>
              </TableCell>
              <TableCell>
                <Button
                  size="sm"
                  variant="destructive"
                  :disabled="taskStatus.running || deleteBackupLoading"
                  @click="previewDeleteBackup(backup)"
                >
                  <Trash2 class="size-4" />
                  Delete
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog v-model:open="deleteBackupDialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ deleteBackupTitle }}</DialogTitle>
          <DialogDescription>
            The snapshot cache is refreshed first. Confirming will delete exactly the listed snapshots for this backup ID.
          </DialogDescription>
        </DialogHeader>

        <div v-if="deleteBackupCandidates.length === 0" class="text-sm text-muted-foreground">
          No snapshots were found for this backup ID.
        </div>

        <div v-else class="max-h-[420px] overflow-auto rounded-lg border border-border/70">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>VM</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Source</TableHead>
                <TableHead>Snapshot</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="snapshot in deleteBackupCandidates" :key="snapshot.snapshotID">
                <TableCell>
                  <div class="font-medium">{{ snapshot.name || 'Unknown' }}</div>
                  <div class="text-xs text-muted-foreground">
                    {{ snapshot.vmType || '-' }} {{ snapshot.vmid ?? '-' }}
                  </div>
                </TableCell>
                <TableCell class="capitalize">{{ snapshot.snapshotType || '-' }}</TableCell>
                <TableCell class="max-w-xs truncate font-mono text-xs text-muted-foreground">
                  {{ snapshot.srcFile || '-' }}
                </TableCell>
                <TableCell class="font-mono text-xs">{{ snapshot.snapshotID }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>

        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button
            :disabled="taskStatus.running || deleteBackupLoading || deleteBackupCandidates.length === 0"
            variant="destructive"
            @click="confirmDeleteBackup"
          >
            {{ deleteBackupLoading ? 'Starting...' : 'Confirm Delete' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
