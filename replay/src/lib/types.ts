export interface ID {
  client: number
  clock: number
}

export interface BlockSnapshot {
  id: ID
  content: string
  deleted: boolean
  left_origin: ID
  right_origin: ID
}

export interface Event {
  type: 'insert' | 'delete' | 'integrate'
  state_vector: Record<string, number>
  blocks: Record<string, BlockSnapshot[]>
}
