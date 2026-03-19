<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { computed } from 'vue'

import { cn } from '@/lib/utils'

type ChartConfig = Record<string, { label: string; color: string }>

interface Props {
  config: ChartConfig
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()

const styleVars = computed(() =>
  Object.fromEntries(
    Object.entries(props.config).map(([key, value]) => [`--color-${key}`, value.color]),
  ),
)
</script>

<template>
  <div data-slot="chart" :class="cn('min-h-[220px] w-full', props.class)" :style="styleVars">
    <slot />
  </div>
</template>
