<script setup lang="ts">
import type { PrimitiveProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { computed } from 'vue'
import { Primitive } from 'reka-ui'

import { cn } from '@/lib/utils'
import { useSidebar } from './sidebar'

interface Props extends PrimitiveProps {
  class?: HTMLAttributes['class']
  isActive?: boolean
  size?: 'default' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  as: 'button',
  size: 'default',
  isActive: false,
})

const { open } = useSidebar()

const buttonClass = computed(() =>
  cn(
    'text-sidebar-foreground ring-sidebar-ring hover:bg-sidebar-accent hover:text-sidebar-accent-foreground data-[active=true]:bg-sidebar-accent data-[active=true]:text-sidebar-accent-foreground flex w-full items-center gap-2 overflow-hidden rounded-lg px-2 py-2 text-left text-sm outline-none transition-colors focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50 [&>svg]:size-4 [&>svg]:shrink-0',
    props.size === 'lg' && 'min-h-12',
    !open.value && 'justify-center px-0',
    props.class,
  ),
)
</script>

<template>
  <Primitive
    data-slot="sidebar-menu-button"
    :as="as"
    :as-child="asChild"
    :data-active="isActive"
    :class="buttonClass"
  >
    <slot />
  </Primitive>
</template>
