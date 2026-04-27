export type Status = 'pending' | 'in_progress' | 'done'

export interface Task {
  id: string
  title: string
  status: Status
  created_at: string
}
