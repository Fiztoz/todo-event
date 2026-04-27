import type { Status, Task } from './types.ts'

const BASE = '/api'

export async function listTasks(): Promise<Task[]> {
  const res = await fetch(`${BASE}/tasks`)
  if (!res.ok) throw new Error('failed to list tasks')
  const data = await res.json()
  return data ?? []
}

export async function getTask(id: string): Promise<Task> {
  const res = await fetch(`${BASE}/tasks/${id}`)
  if (!res.ok) throw new Error('task not found')
  return res.json()
}

export async function createTask(title: string): Promise<Task> {
  const res = await fetch(`${BASE}/tasks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  })
  if (!res.ok) throw new Error('failed to create task')
  return res.json()
}

export async function changeStatus(id: string, status: Status): Promise<Task> {
  const res = await fetch(`${BASE}/tasks/${id}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  })
  if (!res.ok) throw new Error('failed to change status')
  return res.json()
}
