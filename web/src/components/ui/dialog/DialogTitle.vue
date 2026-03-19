<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { DialogTitle as DialogTitlePrimitive, type DialogTitleProps } from 'reka-ui'

import { cn } from '@/lib/utils'

interface Props extends DialogTitleProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()
</script>

<template>
  <DialogTitlePrimitive v-bind="props" :class="cn('text-lg font-semibold text-foreground', props.class)">
    <slot />
  </DialogTitlePrimitive>
</template>
