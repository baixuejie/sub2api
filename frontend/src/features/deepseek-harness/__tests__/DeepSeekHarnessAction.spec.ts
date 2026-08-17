import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createPinia, setActivePinia } from 'pinia'
import type { ApiKey } from '@/types'
import DeepSeekHarnessAction from '../components/DeepSeekHarnessAction.vue'
import {
  createHarnessSession,
  getHarnessProfile,
  getHarnessSession
} from '../api/deepseekHarness'
import type { HarnessInstallSession, HarnessProfileResponse } from '../types/deepseekHarness'

vi.mock('../api/deepseekHarness', () => ({
  createHarnessSession: vi.fn(),
  getHarnessProfile: vi.fn(),
  getHarnessSession: vi.fn()
}))

const profile: HarnessProfileResponse = {
  profile: {
    api_key_id: 42,
    api_key_name: 'Codex key',
    key_hint: 'sk-t...alue',
    group_name: 'OpenAI',
    platform: 'openai',
    provider: 'sub2api-openai',
    provider_name: 'Sub2API OpenAI',
    protocol: 'openai-responses',
    base_url: 'https://api.example.com/v1',
    default_model: 'gpt-5.6-sol',
    selected_model: 'gpt-5.6-sol',
    available_models: [
      { id: 'gpt-5.6-sol', name: 'GPT-5.6 Sol', context_window: 1_050_000, max_tokens: 128_000 }
    ]
  },
  helper_downloads: {
    windows_amd64: 'https://example.com/windows.zip',
    windows_arm64: '',
    darwin_amd64: '',
    darwin_arm64: '',
    linux_amd64: '',
    linux_arm64: '',
    releases_page: 'https://example.com/releases'
  },
  required_node: '>=22.19.0',
  dsh_version: '0.1.0-rc.6'
}

const createdSession: HarnessInstallSession = {
  id: 'session-1',
  profile: profile.profile,
  status: 'awaiting_helper',
  stage: 'waiting_for_helper',
  message: 'Waiting',
  progress: 0,
  launch_uri: 'sub2api-harness://bootstrap?server=https%3A%2F%2Fapi.example.com&ticket=ticket-1&operation_id=session-1',
  ticket_expires_at: '2099-01-01T00:00:00Z',
  created_at: '2026-08-17T00:00:00Z',
  updated_at: '2026-08-17T00:00:00Z',
  expires_at: '2099-01-01T00:00:00Z'
}

function mountAction(status: ApiKey['status'] = 'active') {
  const pinia = createPinia()
  setActivePinia(pinia)
  const i18n = createI18n({ legacy: false, locale: 'zh', messages: { zh: {} } })
  return mount(DeepSeekHarnessAction, {
    attachTo: document.body,
    props: { apiKeyId: 42, status },
    global: { plugins: [pinia, i18n] }
  })
}

beforeEach(() => {
  vi.mocked(getHarnessProfile).mockResolvedValue(profile)
  vi.mocked(createHarnessSession).mockReset()
  vi.mocked(getHarnessSession).mockReset()
})

afterEach(() => {
  document.body.innerHTML = ''
  vi.clearAllMocks()
})

describe('DeepSeekHarnessAction', () => {
  it('loads a server-derived profile without receiving a raw API key prop', async () => {
    const wrapper = mountAction()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(getHarnessProfile).toHaveBeenCalledWith(42, expect.any(AbortSignal))
    expect(document.body.textContent).toContain('Codex key')
    expect(document.body.textContent).toContain('gpt-5.6-sol')
    expect(document.body.textContent).toContain('https://api.example.com/v1')
    expect(wrapper.props()).not.toHaveProperty('apiKey')
    wrapper.unmount()
  })

  it('aborts session creation when the dialog closes before the response', async () => {
    let resolveSession!: (value: HarnessInstallSession) => void
    let createSignal: AbortSignal | undefined
    vi.mocked(createHarnessSession).mockImplementation((_request, signal) => {
      createSignal = signal
      return new Promise((resolve) => {
        resolveSession = resolve
      })
    })
    const wrapper = mountAction()
    await wrapper.get('button').trigger('click')
    await flushPromises()

    const startButton = Array.from(document.body.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('安装并启动')
    )
    expect(startButton).toBeDefined()
    startButton?.click()
    await flushPromises()
    expect(createSignal?.aborted).toBe(false)

    const cancelButton = Array.from(document.body.querySelectorAll('button')).find((button) =>
      button.textContent?.trim() === '取消'
    )
    cancelButton?.click()
    expect(createSignal?.aborted).toBe(true)

    resolveSession(createdSession)
    await flushPromises()
    expect(document.body.textContent).not.toContain('Waiting')
    wrapper.unmount()
  })

  it('disables installation for a non-active API key', async () => {
    const wrapper = mountAction('inactive')
    const button = wrapper.get('button')

    expect(button.attributes('disabled')).toBeDefined()
    await button.trigger('click')
    expect(getHarnessProfile).not.toHaveBeenCalled()
    wrapper.unmount()
  })
})
