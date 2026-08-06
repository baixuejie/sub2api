import { apiClient } from '@/api/client'
import type { ModelPlazaResponse } from '../types/modelPlaza'

/** Extension 对核心公开端点的唯一适配入口。 */
export async function loadModelPlaza(options?: { signal?: AbortSignal }): Promise<ModelPlazaResponse> {
  const { data } = await apiClient.get<ModelPlazaResponse>('/model-plaza', {
    signal: options?.signal
  })
  return data
}
