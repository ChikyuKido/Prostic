<script setup lang="ts">
import { Archive, ClipboardList, Database, HardDriveUpload, LayoutDashboard, LogOut, Settings2 } from 'lucide-vue-next'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { clearToken } from '@/lib/api'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '@/components/ui/sidebar'
import { useSidebar } from '@/components/ui/sidebar/sidebar'

const route = useRoute()
const router = useRouter()
const { open } = useSidebar()

async function logout() {
  clearToken()
  await router.replace('/login')
}

const items = [
  {
    title: 'Overview',
    to: '/',
    icon: LayoutDashboard,
    name: 'overview',
  },
  {
    title: 'Backup',
    to: '/backup',
    icon: HardDriveUpload,
    name: 'backup',
  },
  {
    title: 'Snapshots',
    to: '/snapshots',
    icon: Database,
    name: 'snapshots',
  },
  {
    title: 'Backups',
    to: '/backups',
    icon: Archive,
    name: 'backups',
  },
  {
    title: 'Config',
    to: '/config',
    icon: Settings2,
    name: 'config',
  },
  {
    title: 'Tasks',
    to: '/tasks',
    icon: ClipboardList,
    name: 'tasks',
  },
]
</script>

<template>
  <Sidebar>
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg">
            <div class="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
              <Database class="size-4" />
            </div>
            <div v-if="open" class="grid flex-1 text-left text-sm leading-tight">
              <span class="truncate font-semibold">Prostic</span>
              <span class="truncate text-xs text-sidebar-foreground/60">Backups</span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel v-if="open">App</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in items" :key="item.title">
              <SidebarMenuButton as-child :is-active="route.name === item.name">
                <RouterLink :to="item.to" :title="item.title">
                  <component :is="item.icon" />
                  <span v-if="open" class="truncate">{{ item.title }}</span>
                </RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
    <SidebarFooter class="mt-auto border-t border-sidebar-border/70">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton title="Logout" @click="logout">
            <LogOut />
            <span v-if="open" class="truncate">Logout</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
    <SidebarRail />
  </Sidebar>
</template>
