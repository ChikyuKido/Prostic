const TOKEN_KEY = 'prostic.jwt'
const NEEDS_PASSWORD_CHANGE_KEY = 'prostic.needs-password-change'

export function getToken(): string | null {
  return window.localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  window.localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  window.localStorage.removeItem(TOKEN_KEY)
  clearNeedsPasswordChange()
}

export function hasToken(): boolean {
  return getToken() !== null
}

export function getNeedsPasswordChange(): boolean {
  return window.localStorage.getItem(NEEDS_PASSWORD_CHANGE_KEY) === 'true'
}

export function setNeedsPasswordChange(value: boolean) {
  window.localStorage.setItem(NEEDS_PASSWORD_CHANGE_KEY, String(value))
}

export function clearNeedsPasswordChange() {
  window.localStorage.removeItem(NEEDS_PASSWORD_CHANGE_KEY)
}

function redirectToLogin() {
  if (window.location.pathname !== '/login') {
    window.location.assign('/login')
  }
}

export async function apiFetch(input: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers)
  const token = getToken()

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(input, {
    ...init,
    headers,
  })

  if (response.status === 401) {
    clearToken()
    redirectToLogin()
  }

  return response
}

export async function apiJson<T>(input: string, init: RequestInit = {}): Promise<T> {
  const response = await apiFetch(input, init)

  if (!response.ok) {
    const payload = await response.json().catch(() => null) as { error?: string } | null
    throw new Error(payload?.error ?? `Request failed with status ${response.status}`)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}
