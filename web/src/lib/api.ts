import type { CoreStatus, EntryGroup, EntryGroupInput, Node, Settings, Status } from '@/types'

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
  return response.json() as Promise<T>
}

export const api = {
  status: () => request<Status>('/status'),
  nodes: () => request<Node[]>('/nodes'),
  importNodes: (uri: string) => request<{ count: number; nodes: Node[] }>('/nodes/import', {
    method: 'POST', body: JSON.stringify({ uri }),
  }),
  updateNode: (node: Node) => request<Node>(`/nodes/${node.id}`, {
    method: 'PUT', body: JSON.stringify(node),
  }),
  deleteNode: (id: string) => request<void>(`/nodes/${id}`, { method: 'DELETE' }),
  clearNodes: () => request<void>('/nodes', { method: 'DELETE' }),
  selectNode: (id: string) => request<void>(`/nodes/${id}/select`, { method: 'POST' }),
  delay: (id: string) => request<{ delay: number }>(`/nodes/${id}/delay`, { method: 'POST' }),
  entryGroups: () => request<EntryGroup[]>('/entry-groups'),
  createEntryGroup: (group: EntryGroupInput) => request<EntryGroup>('/entry-groups', {
    method: 'POST', body: JSON.stringify(group),
  }),
  updateEntryGroup: (group: EntryGroup) => request<EntryGroup>(`/entry-groups/${group.id}`, {
    method: 'PUT', body: JSON.stringify(group),
  }),
  deleteEntryGroup: (id: string) => request<void>(`/entry-groups/${id}`, { method: 'DELETE' }),
  selectEntryGroupNode: (id: string, nodeId: string) => request<void>(`/entry-groups/${id}/select`, {
    method: 'POST', body: JSON.stringify({ nodeId }),
  }),
  settings: () => request<Settings>('/settings'),
  saveSettings: (settings: Settings) => request<Settings>('/settings', {
    method: 'PUT', body: JSON.stringify(settings),
  }),
  coreAction: (action: 'start' | 'stop' | 'restart') => request<CoreStatus>(`/core/${action}`, {
    method: 'POST',
  }),
  coreLog: async () => {
    const response = await fetch('/api/core/log')
    if (!response.ok) throw new Error(await response.text() || response.statusText)
    return response.text()
  },
}
