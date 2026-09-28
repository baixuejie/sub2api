import { apiClient } from '@/api/client'
import type {
  GalleryStatus,
  GroupTag,
  PelicanConfig,
  RunDetail,
  RunPage,
  SaveConfig
} from './types'

export interface ListOptions {
  cursor?: string
  topic?: string
  group_id?: number
}
export const pelicanAPI = {
  async config(): Promise<PelicanConfig> {
    return (await apiClient.get('/admin/pelican/config')).data
  },
  async save(config: SaveConfig): Promise<PelicanConfig> {
    return (await apiClient.put('/admin/pelican/config', config)).data
  },
  async clearKey(revision: number): Promise<PelicanConfig> {
    return (await apiClient.delete('/admin/pelican/config/key', { params: { revision } })).data
  },
  async runNow(): Promise<{ id: number; status: string }> {
    return (await apiClient.post('/admin/pelican/run')).data
  },
  async list(options: ListOptions = {}, admin = false, signal?: AbortSignal): Promise<RunPage> {
    return (
      await apiClient.get(admin ? '/admin/pelican/runs' : '/pelican/runs', {
        params: { ...options, limit: 30 },
        signal
      })
    ).data
  },
  async status(signal?: AbortSignal): Promise<GalleryStatus> {
    return (await apiClient.get('/pelican/status', { signal })).data
  },
  async groups(signal?: AbortSignal): Promise<GroupTag[]> {
    return (await apiClient.get('/pelican/groups', { signal })).data
  },
  async artifact(
    id: number,
    signal?: AbortSignal
  ): Promise<{ html: string; policy_version: number }> {
    return (await apiClient.get(`/pelican/runs/${id}/artifact`, { signal })).data
  },
  async detail(id: number): Promise<RunDetail> {
    return (await apiClient.get(`/admin/pelican/runs/${id}`)).data
  },
  async source(id: number): Promise<{ source: string }> {
    return (await apiClient.get(`/admin/pelican/runs/${id}/source`)).data
  }
}
