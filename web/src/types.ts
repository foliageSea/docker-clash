export interface Node {
  id: string
  name: string
  type: string
  server: string
  port: number
  dialerProxy?: string
  options?: Record<string, unknown>
  createdAt: string
}

export interface EntryGroup {
  id: string
  name: string
  type: 'select' | 'fallback'
  nodeIds: string[]
  selectedNodeId?: string
  testUrl?: string
  interval?: number
}

export type EntryGroupInput = Omit<EntryGroup, 'id' | 'selectedNodeId'>

export interface Settings {
  listen: string
  mixedPort: number
  allowLan: boolean
  bindAddress: string
  externalAddress: string
  externalPort: number
  selectedNode?: string
}

export interface CoreStatus {
  running: boolean
  pid?: number
  error?: string
}

export interface Status {
  core: CoreStatus
  settings: Settings
  nodeCount: number
  entryGroupCount: number
}
