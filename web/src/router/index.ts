import { createRouter, createWebHistory } from 'vue-router'

import { getNeedsPasswordChange, hasToken } from '@/lib/api'
import AppLayoutView from '@/views/AppLayoutView.vue'
import ChangePasswordView from '@/views/ChangePasswordView.vue'
import LoginView from '@/views/LoginView.vue'
import OverviewView from '@/views/OverviewView.vue'
import SnapshotsView from '@/views/SnapshotsView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
    {
      path: '/change-password',
      name: 'change-password',
      component: ChangePasswordView,
    },
    {
      path: '/',
      component: AppLayoutView,
      children: [
        {
          path: '',
          name: 'overview',
          component: OverviewView,
        },
        {
          path: 'snapshots',
          name: 'snapshots',
          component: SnapshotsView,
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const loggedIn = hasToken()
  const needsPasswordChange = getNeedsPasswordChange()

  if (!loggedIn && to.name !== 'login') {
    return { name: 'login' }
  }

  if (loggedIn && to.name === 'login') {
    return { name: needsPasswordChange ? 'change-password' : 'overview' }
  }

  if (loggedIn && needsPasswordChange && to.name !== 'change-password') {
    return { name: 'change-password' }
  }

  if (loggedIn && !needsPasswordChange && to.name === 'change-password') {
    return { name: 'overview' }
  }

  return true
})

export default router
