<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ClipboardList, RefreshCw, Trash2 } from 'lucide-vue-next'

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

interface PruneCandidate {
  snapshotID: string
  time: string
  backupID: string
  vmid: number | null
  name: string
  vmType: string
  snapshotType: string
  srcFile: string
}

const loading = ref(true)
const refreshing = ref(false)
const pruneLoading = ref(false)
const error = ref('')
const tasks = ref<Task[]>([])
const logsTask = ref<Task | null>(null)
const pruneDialogOpen = ref(false)
const pruneCandidates = ref<PruneCandidate[]>([])
const taskStatus = ref<TaskStatus>({ running: false })
let statusTimer: number | null = null

const pruneTitle = computed(() => {
  if (pruneCandidates.value.length === 0) {
    return 'Nothing to prune'
  }

  return `Delete ${pruneCandidates.value.length} snapshots`
})

async function loadTasks() {
  loading.value = true
  error.value = ''

  try {
    const response = await apiJson<{ tasks: Task[] }>('/api/tasks')
    tasks.value = response.tasks
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load tasks'
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
      await loadTasks()
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load task status'
  }
}

async function refreshTasks() {
  refreshing.value = true
  try {
    await loadTasks()
    await loadTaskStatus()
  } finally {
    refreshing.value = false
  }
}

async function previewPruneNotInConfig() {
  if (taskStatus.value.running) {
    return
  }

  pruneLoading.value = true
  error.value = ''

  try {
    const response = await apiJson<{ snapshots: PruneCandidate[] }>('/api/tasks/prune-not-in-config', {
      method: 'POST',
    })
    pruneCandidates.value = response.snapshots
    pruneDialogOpen.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to prepare prune task'
  } finally {
    pruneLoading.value = false
  }
}

async function confirmPruneNotInConfig() {
  pruneLoading.value = true
  error.value = ''

  try {
    await apiJson<{ task: Task }>('/api/tasks/prune-not-in-config?confirm=true', {
      method: 'POST',
      body: JSON.stringify({ snapshots: pruneCandidates.value }),
    })
    pruneDialogOpen.value = false
    pruneCandidates.value = []
    await loadTaskStatus()
    await loadTasks()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to run prune task'
  } finally {
    pruneLoading.value = false
  }
}

function taskLabel(purpose: string) {
  if (purpose === 'prune_not_in_config') {
    return 'Prune Not In Config'
  }
  if (purpose === 'delete_snapshot') {
    return 'Delete Snapshot'
  }
  if (purpose === 'delete_backup_id') {
    return 'Delete Backup ID'
  }

  return purpose
}

function statusClass(status: string) {
  if (status === 'success') {
    return 'text-emerald-600'
  }
  if (status === 'failed') {
    return 'text-destructive'
  }
  return 'text-amber-600'
}

onMounted(() => {
  void loadTasks()
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
      <h1 class="text-3xl font-semibold tracking-tight">Tasks</h1>
      <div class="flex flex-col gap-2 sm:flex-row">
        <Button variant="outline" :disabled="refreshing || loading" @click="refreshTasks">
          <RefreshCw class="mr-2 size-4" />
          {{ refreshing ? 'Refreshing...' : 'Refresh' }}
        </Button>
        <Button :disabled="pruneLoading || taskStatus.running" @click="previewPruneNotInConfig">
          <Trash2 class="mr-2 size-4" />
          {{ pruneLoading ? 'Preparing...' : 'Prune Not In Config' }}
        </Button>
      </div>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <p v-if="taskStatus.running" class="text-sm text-muted-foreground">
      Running: {{ taskStatus.purpose ? taskLabel(taskStatus.purpose) : (taskStatus.kind || 'job') }}
      <span v-if="taskStatus.startedAt"> since {{ new Date(taskStatus.startedAt).toLocaleString() }}</span>
    </p>

    <Card class="border-border/70 bg-card/95">
      <CardHeader>
        <CardTitle>Task History</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-sm text-muted-foreground">Loading tasks...</div>
        <div v-else-if="tasks.length === 0" class="text-sm text-muted-foreground">No tasks yet.</div>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Started</TableHead>
              <TableHead>Purpose</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Finished</TableHead>
              <TableHead class="w-[120px]">Logs</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="task in tasks" :key="task.id">
              <TableCell class="whitespace-nowrap text-muted-foreground">
                {{ new Date(task.startedAt).toLocaleString() }}
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-2">
                  <ClipboardList class="size-4 text-muted-foreground" />
                  <span>{{ taskLabel(task.purpose) }}</span>
                </div>
              </TableCell>
              <TableCell>
                <span class="font-medium capitalize" :class="statusClass(task.status)">
                  {{ task.status }}
                </span>
              </TableCell>
              <TableCell class="whitespace-nowrap text-muted-foreground">
                {{ task.finishedAt ? new Date(task.finishedAt).toLocaleString() : '-' }}
              </TableCell>
              <TableCell>
                <Button variant="outline" size="sm" @click="logsTask = task">View</Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog v-model:open="pruneDialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ pruneTitle }}</DialogTitle>
          <DialogDescription>
            The current snapshot cache is refreshed first. Confirming will delete exactly the listed snapshots.
          </DialogDescription>
        </DialogHeader>

        <div v-if="pruneCandidates.length === 0" class="text-sm text-muted-foreground">
          No snapshots outside the current config were found.
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
              <TableRow v-for="snapshot in pruneCandidates" :key="snapshot.snapshotID">
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
            :disabled="pruneLoading || taskStatus.running || pruneCandidates.length === 0"
            variant="destructive"
            @click="confirmPruneNotInConfig"
          >
            {{ pruneLoading ? 'Deleting...' : 'Confirm Delete' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="logsTask !== null" @update:open="(open) => { if (!open) logsTask = null }">
      <DialogContent class="max-w-4xl">
        <DialogHeader>
          <DialogTitle>{{ logsTask ? taskLabel(logsTask.purpose) : 'Task Logs' }}</DialogTitle>
          <DialogDescription>
            {{ logsTask ? new Date(logsTask.startedAt).toLocaleString() : '' }}
          </DialogDescription>
        </DialogHeader>
        <div class="max-h-[480px] overflow-auto rounded-lg border border-border/70 bg-muted/30 p-4">
          <pre class="whitespace-pre-wrap break-words text-xs text-foreground">{{ logsTask?.logs || 'No logs.' }}</pre>
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
