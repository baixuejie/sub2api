import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from '@/api/client'
import {
  createHarnessSession,
  getHarnessProfile,
  getHarnessSession
} from '../api/deepseekHarness'

afterEach(() => vi.restoreAllMocks())

describe('DeepSeek Harness feature API', () => {
  it('loads a profile using only the selected API key id', async () => {
    const signal = new AbortController().signal
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ data: { profile: {} } })

    await getHarnessProfile(42, signal)

    expect(get).toHaveBeenCalledWith('/deepseek-harness/profile', {
      params: { api_key_id: 42 },
      signal
    })
  })

  it('creates an install session without sending the raw API key', async () => {
    const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { id: 'session-1' } })

    const signal = new AbortController().signal
    await createHarnessSession({ api_key_id: 42, model: 'gpt-5.6-sol' }, signal)

    expect(post).toHaveBeenCalledWith('/deepseek-harness/sessions', {
      api_key_id: 42,
      model: 'gpt-5.6-sol'
    }, { signal })
    expect(JSON.stringify(post.mock.calls[0])).not.toContain('sk-')
  })

  it('encodes the session id when polling status', async () => {
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ data: { id: 'session/1' } })

    await getHarnessSession('session/1')

    expect(get).toHaveBeenCalledWith('/deepseek-harness/sessions/session%2F1', {
      signal: undefined
    })
  })
})
