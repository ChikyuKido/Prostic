import { inject, provide, ref, watch, type InjectionKey, type Ref } from 'vue'

type SidebarContext = {
  open: Ref<boolean>
  toggle: () => void
  setOpen: (value: boolean) => void
}

const SIDEBAR_CONTEXT_KEY: InjectionKey<SidebarContext> = Symbol('sidebar')

type SidebarProviderOptions = {
  defaultOpen?: boolean
  storageKey?: string
}

export function provideSidebar(options: SidebarProviderOptions = {}) {
  const { defaultOpen = true, storageKey } = options
  const initialValue = storageKey
    ? window.localStorage.getItem(storageKey) !== 'false'
    : defaultOpen

  const open = ref(initialValue)

  const setOpen = (value: boolean) => {
    open.value = value
  }

  const toggle = () => {
    open.value = !open.value
  }

  if (storageKey) {
    watch(
      open,
      (value) => {
        window.localStorage.setItem(storageKey, String(value))
      },
      { immediate: true },
    )
  }

  provide(SIDEBAR_CONTEXT_KEY, {
    open,
    toggle,
    setOpen,
  })

  return {
    open,
    toggle,
    setOpen,
  }
}

export function useSidebar() {
  const context = inject(SIDEBAR_CONTEXT_KEY)
  if (!context) {
    throw new Error('useSidebar must be used within SidebarProvider')
  }

  return context
}
