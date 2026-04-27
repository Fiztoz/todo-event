const BASE = '/api'

export async function listTasks() {
  const res = await fetch(`${BASE}/tasks`)
  if (!res.ok) throw new Error('failed to list tasks')
  return res.json()
}

export async function getTask(id) {
  const res = await fetch(`${BASE}/tasks/${id}`)
  if (!res.ok) throw new Error('task not found')
  return res.json()
}

export async function createTask(title) {
  const res = await fetch(`${BASE}/tasks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  })
  if (!res.ok) throw new Error('failed to create task')
  return res.json()
}

export async function changeStatus(id, status) {
  const res = await fetch(`${BASE}/tasks/${id}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  })
  if (!res.ok) throw new Error('failed to change status')
  return res.json()
}
