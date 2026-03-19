<script setup lang="ts">
import { computed } from 'vue'

import { ChartContainer, type ChartConfig } from '@/components/ui/chart'

interface HistoryPoint {
  timestamp: string
  totalSize: number
  totalUncompressedSize: number
}

interface Props {
  history: HistoryPoint[]
}

const props = defineProps<Props>()

const chartConfig = {
  repoSize: {
    label: 'Repo size',
    color: 'var(--chart-1)',
  },
  rawSize: {
    label: 'Raw size',
    color: 'var(--chart-2)',
  },
} satisfies ChartConfig

const width = 760
const height = 280
const padding = { top: 16, right: 16, bottom: 32, left: 16 }
const innerWidth = width - padding.left - padding.right
const innerHeight = height - padding.top - padding.bottom

const chartPoints = computed(() =>
  props.history.map((point) => ({
    ...point,
    timestampDate: new Date(point.timestamp),
  })),
)

const maxValue = computed(() => {
  const max = Math.max(
    ...chartPoints.value.flatMap((point) => [point.totalSize, point.totalUncompressedSize]),
    0,
  )

  return max > 0 ? max : 1
})

function xForIndex(index: number) {
  if (chartPoints.value.length <= 1) {
    return padding.left + innerWidth / 2
  }

  return padding.left + (index / (chartPoints.value.length - 1)) * innerWidth
}

function yForValue(value: number) {
  return padding.top + innerHeight - (value / maxValue.value) * innerHeight
}

const repoLine = computed(() =>
  chartPoints.value.map((point, index) => `${xForIndex(index)},${yForValue(point.totalSize)}`).join(' '),
)

const rawLine = computed(() =>
  chartPoints.value.map((point, index) => `${xForIndex(index)},${yForValue(point.totalUncompressedSize)}`).join(' '),
)

const gridLines = computed(() => [0, 0.25, 0.5, 0.75, 1].map((step) => padding.top + innerHeight * step))

const xTicks = computed(() => {
  if (chartPoints.value.length <= 3) {
    return chartPoints.value.map((point, index) => ({
      x: xForIndex(index),
      label: point.timestampDate.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
      }),
    }))
  }

  const indexes = [0, Math.floor((chartPoints.value.length - 1) / 2), chartPoints.value.length - 1]
  return indexes.map((index) => ({
    x: xForIndex(index),
    label: chartPoints.value[index].timestampDate.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
    }),
  }))
})
</script>

<template>
  <ChartContainer :config="chartConfig" class="min-h-[320px]">
    <div class="flex items-center gap-4 pb-4 text-sm">
      <div class="flex items-center gap-2">
        <span class="size-2.5 rounded-full bg-[var(--color-repoSize)]" />
        <span class="text-muted-foreground">{{ chartConfig.repoSize.label }}</span>
      </div>
      <div class="flex items-center gap-2">
        <span class="size-2.5 rounded-full bg-[var(--color-rawSize)]" />
        <span class="text-muted-foreground">{{ chartConfig.rawSize.label }}</span>
      </div>
    </div>

    <div v-if="history.length === 0" class="flex min-h-[240px] items-center justify-center text-sm text-muted-foreground">
      No historic stats yet.
    </div>

    <svg v-else :viewBox="`0 0 ${width} ${height}`" class="h-[280px] w-full overflow-visible">
      <g>
        <line
          v-for="line in gridLines"
          :key="line"
          :x1="padding.left"
          :x2="width - padding.right"
          :y1="line"
          :y2="line"
          stroke="var(--border)"
          stroke-width="1"
        />
      </g>

      <polyline
        fill="none"
        :points="rawLine"
        stroke="var(--color-rawSize)"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="3"
      />
      <polyline
        fill="none"
        :points="repoLine"
        stroke="var(--color-repoSize)"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="3"
      />

      <g v-for="point in xTicks" :key="point.x">
        <text
          :x="point.x"
          :y="height - 8"
          text-anchor="middle"
          class="fill-muted-foreground text-[11px]"
        >
          {{ point.label }}
        </text>
      </g>
    </svg>
  </ChartContainer>
</template>
