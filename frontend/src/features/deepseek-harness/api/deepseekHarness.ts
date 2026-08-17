import { apiClient } from '@/api/client'
import type {
  CreateHarnessSessionRequest,
  HarnessInstallSession,
  HarnessProfileResponse
} from '../types/deepseekHarness'

const basePath = '/deepseek-harness'

export async function getHarnessProfile(
  apiKeyId: number,
  signal?: AbortSignal
): Promise<HarnessProfileResponse> {
  const { data } = await apiClient.get<HarnessProfileResponse>(`${basePath}/profile`, {
    params: { api_key_id: apiKeyId },
    signal
  })
  return data
}

export async function createHarnessSession(
  request: CreateHarnessSessionRequest,
  signal?: AbortSignal
): Promise<HarnessInstallSession> {
  const { data } = await apiClient.post<HarnessInstallSession>(`${basePath}/sessions`, request, {
    signal
  })
  return data
}

export async function getHarnessSession(
  sessionId: string,
  signal?: AbortSignal
): Promise<HarnessInstallSession> {
  const { data } = await apiClient.get<HarnessInstallSession>(
    `${basePath}/sessions/${encodeURIComponent(sessionId)}`,
    { signal }
  )
  return data
}
