import type { User } from './types.ts'

const BASE = '/api'

async function handle<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }
  return res.json()
}

export async function register(name: string, email: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, email }),
  }))
}

export async function verifyEmail(id: string, token: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/${id}/verify-email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token }),
  }))
}

export async function getUser(id: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/${id}`))
}

export async function completeProfile(id: string, bio: string): Promise<User> {
  return handle(await fetch(`${BASE}/users/${id}/complete-profile`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bio }),
  }))
}
