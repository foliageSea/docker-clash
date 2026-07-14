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
  selectedNode?: string
}
export interface Status {
  core: { running: boolean; pid?: number; error?: string }
  settings: Settings
  nodeCount: number
  entryGroupCount: number
}
async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(body.error || response.statusText)
  }
  if (response.status === 204) return undefined as T
  return response.json()
}
export const api = {
  status: () => request<Status>('/status'),
  nodes: () => request<Node[]>('/nodes'),
  entryGroups: () => request<EntryGroup[]>('/entry-groups'),
  createEntryGroup: (group: EntryGroupInput) =>
    request<EntryGroup>('/entry-groups', { method: 'POST', body: JSON.stringify(group) }),
  updateEntryGroup: (group: EntryGroup) =>
    request<EntryGroup>(`/entry-groups/${group.id}`, {
      method: 'PUT',
      body: JSON.stringify(group),
    }),
  deleteEntryGroup: (id: string) => request<void>(`/entry-groups/${id}`, { method: 'DELETE' }),
  selectEntryGroupNode: (id: string, nodeId: string) =>
    request<void>(`/entry-groups/${id}/select`, {
      method: 'POST',
      body: JSON.stringify({ nodeId }),
    }),
  importNode: (uri: string) =>
    request<{ count: number; nodes: Node[] }>('/nodes/import', {
      method: 'POST',
      body: JSON.stringify({ uri }),
    }),
  updateNode: (n: Node) =>
    request<Node>(`/nodes/${n.id}`, { method: 'PUT', body: JSON.stringify(n) }),
  deleteNode: (id: string) => request<void>(`/nodes/${id}`, { method: 'DELETE' }),
  clearNodes: () => request<void>('/nodes', { method: 'DELETE' }),
  selectNode: (id: string) => request<void>(`/nodes/${id}/select`, { method: 'POST' }),
  delay: (id: string) => request<{ delay: number }>(`/nodes/${id}/delay`, { method: 'POST' }),
  settings: () => request<Settings>('/settings'),
  saveSettings: (s: Settings) =>
    request<Settings>('/settings', { method: 'PUT', body: JSON.stringify(s) }),
  core: (action: string) => request(`/core/${action}`, { method: 'POST' }),
}
