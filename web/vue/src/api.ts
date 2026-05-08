import type { Session, Status, Task } from './types.ts'

const BASE = '/api'
const TOKEN_KEY = 'todoe_token'

export class AuthError extends Error {}

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? ''
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

function authHeaders(): Record<string, string> {
  const token = getToken()
  return token ? { Authorization: token } : {}
}

async function taskFetch(input: string, init?: RequestInit): Promise<Response> {
  const res = await fetch(input, {
    ...init,
    headers: { ...authHeaders(), ...(init?.headers as Record<string, string> ?? {}) },
  })
  if (res.status === 401) throw new AuthError('unauthorized')
  return res
}

export async function login(email: string, password: string): Promise<Session> {
  const res = await fetch(`${BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) throw new Error('invalid email or password')
  const session: Session = await res.json()
  setToken(session.token)
  return session
}

export async function logout(): Promise<void> {
  const token = getToken()
  if (token) {
    await fetch(`${BASE}/auth/logout`, {
      method: 'POST',
      headers: { Authorization: token },
    })
  }
  clearToken()
}

export async function listTasks(): Promise<Task[]> {
  const res = await taskFetch(`${BASE}/tasks`)
  if (!res.ok) throw new Error('failed to list tasks')
  const data = await res.json()
  return data ?? []
}

export async function getTask(id: string): Promise<Task> {
  const res = await taskFetch(`${BASE}/tasks/${id}`)
  if (!res.ok) throw new Error('task not found')
  return res.json()
}

export async function createTask(title: string): Promise<Task> {
  const res = await taskFetch(`${BASE}/tasks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  })
  if (!res.ok) throw new Error('failed to create task')
  return res.json()
}

export async function changeStatus(id: string, status: Status): Promise<Task> {
  const res = await taskFetch(`${BASE}/tasks/${id}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  })
  if (!res.ok) throw new Error('failed to change status')
  return res.json()
}
