<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { apiJson } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface ConfigVM {
  id: number
  name: string
  type: string
  isVM: boolean
  configFile: string
  disks: string[]
}

interface ConfigResponse {
  vms: ConfigVM[]
}

const loading = ref(true)
const error = ref('')
const config = ref<ConfigResponse | null>(null)

async function loadConfig() {
  loading.value = true
  error.value = ''

  try {
    config.value = await apiJson<ConfigResponse>('/api/config')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load config'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadConfig()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-semibold tracking-tight">Config</h1>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <div v-if="loading" class="grid gap-4 lg:grid-cols-2">
      <Card v-for="card in 2" :key="card" class="border-border/70 bg-card/95">
        <CardContent class="p-6 text-sm text-muted-foreground">Loading...</CardContent>
      </Card>
    </div>

    <div v-else-if="!config || config.vms.length === 0" class="text-sm text-muted-foreground">
      No configured VMs or LXCs.
    </div>

    <div v-else class="grid gap-4 xl:grid-cols-2">
      <Card v-for="vm in config.vms" :key="vm.id" class="border-border/70 bg-card/95">
        <CardHeader class="space-y-1">
          <CardTitle class="flex items-center justify-between gap-4">
            <span class="truncate">{{ vm.name }}</span>
            <span class="text-sm font-normal text-muted-foreground">
              {{ vm.type }} {{ vm.id }}
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-1">
            <div class="text-xs uppercase tracking-[0.16em] text-muted-foreground">Config</div>
            <div class="font-mono text-xs text-foreground">{{ vm.configFile }}</div>
          </div>
          <div class="space-y-2">
            <div class="text-xs uppercase tracking-[0.16em] text-muted-foreground">Disks</div>
            <div v-if="vm.disks.length === 0" class="text-sm text-muted-foreground">No disks configured.</div>
            <div v-else class="space-y-2">
              <div
                v-for="disk in vm.disks"
                :key="disk"
                class="rounded-lg border border-border/70 bg-background/70 px-3 py-2 font-mono text-xs text-foreground"
              >
                {{ disk }}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
