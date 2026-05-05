import type { Task, TaskStatus } from './types.ts'

const BASE = '/api'

async function handle<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }
  return res.json()
}

export async function listTasks(): Promise<Task[]> {
  return handle(await fetch(`${BASE}/tasks`))
}

export async function createTask(title: string): Promise<Task> {
  return handle(await fetch(`${BASE}/tasks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  }))
}

export async function getTask(id: string): Promise<Task> {
  return handle(await fetch(`${BASE}/tasks/${id}`))
}

export async function changeStatus(id: string, status: TaskStatus): Promise<Task> {
  return handle(await fetch(`${BASE}/tasks/${id}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  }))
}
