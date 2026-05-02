export type Status = 'pending' | 'in_progress' | 'done'

export interface Task {
  id: string
  title: string
  status: Status
  created_at: string
}

export interface Session {
  id: string
  user_id: string
  token: string
  active: boolean
  created_at: string
  expires_at: string
}
